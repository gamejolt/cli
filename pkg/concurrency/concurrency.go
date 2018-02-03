package concurrency

import (
	"context"
	"sync"
)

const (
	// OpResume signifies a resumed operation
	OpResume = iota

	// OpPause signifies a paused operation
	OpPause = iota

	// OpCancel signifies a canceled operation
	OpCancel = iota
)

type resumer interface {
	pause(self bool)
	resume(self bool)
}

// Resumer comment
type Resumer interface {
	resumer
	Pause()
	Resume()
	Cancel()
	Done() <-chan struct{}
	Wait() <-chan int
}

var nextResumableId int

// Resumable handles a resumable operation
type Resumable struct {
	id            int
	ctx           context.Context
	interruptible bool
	cancel        context.CancelFunc
	mu            sync.Mutex
	aborted       bool
	onResume      chan bool
	isSelfPaused  bool
	isPaused      int
	children      map[resumer]bool
}

// NewResumable creates a new Resumable with an optional parent
func NewResumable(parent *Resumable) *Resumable {
	return newResumable(nil, parent, false)
}

// NewInterruptible creates a new Resumable that gets canceled when paused with an optional parent
func NewInterruptible(parent *Resumable) *Resumable {
	return newResumable(nil, parent, true)
}

// NewResumableFromContext creates a new Resumable from a context
func NewResumableFromContext(ctx context.Context) *Resumable {
	return newResumable(ctx, nil, false)
}

// NewInterruptibleFromContext creates a new Resumable that gets canceled when paused from a context
func NewInterruptibleFromContext(ctx context.Context) *Resumable {
	return newResumable(ctx, nil, true)
}

// NOTE: if the parent is specified, the parent's context is used instead
func newResumable(ctx context.Context, parent *Resumable, interruptible bool) *Resumable {
	if parent == nil {
		if ctx == nil {
			ctx = context.Background()
		}

		ctx, cancel := context.WithCancel(ctx)
		child := &Resumable{
			id:            nextResumableId,
			ctx:           ctx,
			aborted:       ctx.Err() != nil, // if ctx.Err() is non nil it means the context has already been canceled, so set aborted to true
			interruptible: interruptible,
		}
		nextResumableId++

		newCancel := func() {
			child.aborted = true
			cancel()
		}
		child.cancel = newCancel

		go func() {
			<-child.ctx.Done()
			child.aborted = true
		}()

		return child
	}

	parent.mu.Lock()
	defer parent.mu.Unlock()

	ctx, cancel := context.WithCancel(parent.ctx)
	isPaused := parent.isPaused
	if parent.isSelfPaused {
		isPaused++
	}
	var onResume chan bool
	if isPaused > 0 {
		onResume = make(chan bool)
	}

	child := &Resumable{
		id:            nextResumableId,
		ctx:           ctx,
		aborted:       ctx.Err() != nil,
		onResume:      onResume,
		isPaused:      isPaused,
		interruptible: interruptible,
	}
	nextResumableId++

	newCancel := func() {
		child.aborted = true
		cancel()
	}
	child.cancel = newCancel

	go func() {
		<-child.ctx.Done()
		child.aborted = true
	}()

	if parent.children == nil {
		parent.children = make(map[resumer]bool)
	}
	parent.children[child] = true

	if interruptible && isPaused > 0 {
		child.cancel()
	}

	return child
}

// Context returns the resumable context
func (r *Resumable) Context() context.Context {
	return r.ctx
}

func (r *Resumable) pause(self bool) {
	if r.interruptible {
		r.cancel()
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if self {
		if r.isSelfPaused {
			return // already paused
		}
		r.isSelfPaused = true
	} else {
		r.isPaused++
	}

	// If just got paused
	if r.onResume == nil {
		r.onResume = make(chan bool)
	}

	for c := range r.children {
		c.pause(false)
	}
}

// Pause comment
func (r *Resumable) Pause() {
	r.pause(true)
}

func (r *Resumable) resume(self bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.aborted {
		return
	}

	if self {
		if !r.isSelfPaused {
			return
		}
		r.isSelfPaused = false
	} else {
		r.isPaused--
	}

	if !r.isSelfPaused && r.isPaused == 0 && r.onResume != nil {
		close(r.onResume)
		r.onResume = nil
	}

	for c := range r.children {
		c.resume(false)
	}
}

// Resume comment
func (r *Resumable) Resume() {
	r.resume(true)
}

// Cancel comment
func (r *Resumable) Cancel() {
	r.cancel()
}

// Done comment
func (r *Resumable) Done() <-chan struct{} {
	return r.ctx.Done()
}

// IsRunning comment
func (r *Resumable) IsRunning() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return !r.aborted && !r.isSelfPaused && r.isPaused == 0
}

// Wait comment
func (r *Resumable) Wait() <-chan int {
	ch := make(chan int, 1)

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.aborted {
		ch <- OpCancel
		close(ch)
		return ch
	}

	if !r.isSelfPaused && r.isPaused == 0 {
		ch <- OpResume
		close(ch)
		return ch
	}

	go func() {
		select {
		case <-r.ctx.Done():
			ch <- OpCancel
			close(ch)
		case <-r.onResume:
			ch <- OpResume
			close(ch)
		}
	}()

	return ch
}

type chainFunc func() error

// ChainResumableTasks comment
func ChainResumableTasks(resumable Resumer, tasks ...chainFunc) <-chan error {
	ch := make(chan error)

	go func() {
		for _, task := range tasks {
			if resume := <-resumable.Wait(); resume == OpCancel {
				ch <- context.Canceled
				return
			}

			if err := task(); err != nil {
				ch <- err
				return
			}
		}
		ch <- nil
	}()

	return ch
}
