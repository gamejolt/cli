package concurrency

import (
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
func TestCancel(t *testing.T) {
	r := NewResumable(nil)

	count := 0
	go func() {
		for {
			select {
			case <-r.Done():
				return
			case <-time.After(10 * time.Millisecond):
				count++
			}
		}
	}()

	<-time.After(55 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to be 5, received: %d", count)
	}

	r.Cancel()
	<-time.After(45 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to remain 5, received: %d", count)
	}
}

func TestPause(t *testing.T) {
	r := NewResumable(nil)

	count := 0
	go func() {
		for {
			<-time.After(10 * time.Millisecond)

			if resume := <-r.Wait(); resume == OpCancel {
				return
			}
			count++
		}
	}()

	<-time.After(55 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to be 5, received: %d", count)
	}

	r.Pause()
	<-time.After(45 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to remain 5, received: %d", count)
	}
}

func TestResume(t *testing.T) {
	r := NewResumable(nil)
	r.Pause()

	count := 0
	go func() {
		for {
			if resume := <-r.Wait(); resume == OpCancel {
				return
			}

			<-time.After(10 * time.Millisecond)
			count++
		}
	}()

	<-time.After(45 * time.Millisecond)
	if count != 0 {
		t.Errorf("Expected count to be 0, received: %d", count)
	}

	r.Resume()
	<-time.After(55 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to be 5, received: %d", count)
	}
}

func TestInterruptibleCancel(t *testing.T) {
	r := NewInterruptible(nil)

	count := 0
	go func() {
		for {
			select {
			case <-r.Done():
				return
			case <-time.After(10 * time.Millisecond):
				count++
			}
		}
	}()

	<-time.After(55 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to be 5, received: %d", count)
	}

	r.Cancel()
	<-time.After(45 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to remain 5, received: %d", count)
	}
}

func TestInterruptiblePause(t *testing.T) {
	r := NewInterruptible(nil)

	count := 0
	go func() {
		for {
			select {
			case <-r.Done():
				return
			case <-time.After(10 * time.Millisecond):
				count++
			}
		}
	}()

	<-time.After(55 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to be 5, received: %d", count)
	}

	r.Pause()
	<-time.After(45 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to remain 5, received: %d", count)
	}
}

func TestInterruptibleResume(t *testing.T) {
	r := NewInterruptible(nil)
	r.Pause()

	count := 0
	go func() {
		for {
			<-time.After(10 * time.Millisecond)
			if r.IsRunning() {
				count++
			}
		}
	}()

	<-time.After(45 * time.Millisecond)
	if count != 0 {
		t.Errorf("Expected count to be 0, received: %d", count)
	}

	r.Resume()
	<-time.After(55 * time.Millisecond)
	if count != 0 {
		t.Errorf("Expected count to remain 0, received: %d", count)
	}
}

func TestParentCancel(t *testing.T) {
	r := NewResumable(nil)
	c := NewResumable(r)

	count := 0
	go func() {
		for {
			select {
			case <-c.Done():
				return
			case <-time.After(10 * time.Millisecond):
				count++
			}
		}
	}()

	<-time.After(55 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to be 5, received: %d", count)
	}

	r.Cancel()
	<-time.After(45 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to remain 5, received: %d", count)
	}
}

func TestParentPause(t *testing.T) {
	r := NewResumable(nil)
	c := NewResumable(r)

	count := 0
	go func() {
		for {
			<-time.After(10 * time.Millisecond)

			if resume := <-c.Wait(); resume == OpCancel {
				return
			}
			count++
		}
	}()

	<-time.After(55 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to be 5, received: %d", count)
	}

	r.Pause()
	<-time.After(45 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to remain 5, received: %d", count)
	}
}

func TestParentResume(t *testing.T) {
	r := NewResumable(nil)
	c := NewResumable(r)
	r.Pause()

	count := 0
	go func() {
		for {
			if resume := <-c.Wait(); resume == OpCancel {
				return
			}

			<-time.After(10 * time.Millisecond)
			count++
		}
	}()

	<-time.After(45 * time.Millisecond)
	if count != 0 {
		t.Errorf("Expected count to be 0, received: %d", count)
	}

	r.Resume()
	<-time.After(55 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to be 5, received: %d", count)
	}
}

func TestParentCancelBeforeChild(t *testing.T) {
	r := NewResumable(nil)
	r.Cancel()
	c := NewResumable(r)

	count := 0
	go func() {
		for {
			select {
			case <-c.Done():
				return
			case <-time.After(10 * time.Millisecond):
				count++
			}
		}
	}()

	<-time.After(55 * time.Millisecond)
	if count != 0 {
		t.Errorf("Expected count to be 0, received: %d", count)
	}
}

func TestParentPauseBeforeChild(t *testing.T) {
	r := NewResumable(nil)
	r.Pause()
	c := NewResumable(r)

	count := 0
	go func() {
		for {
			<-time.After(10 * time.Millisecond)

			if resume := <-c.Wait(); resume == OpCancel {
				return
			}
			count++
		}
	}()

	<-time.After(45 * time.Millisecond)
	if count != 0 {
		t.Errorf("Expected count to be 0, received: %d", count)
	}
}

func TestParentResumeBeforeChild(t *testing.T) {
	r := NewResumable(nil)
	r.Pause()
	c := NewResumable(r)

	count := 0
	go func() {
		for {
			if resume := <-c.Wait(); resume == OpCancel {
				return
			}

			<-time.After(10 * time.Millisecond)
			count++
		}
	}()

	<-time.After(45 * time.Millisecond)
	if count != 0 {
		t.Errorf("Expected count to be 0, received: %d", count)
	}

	r.Resume()
	<-time.After(55 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to be 5, received: %d", count)
	}
}

func TestParentCancelInterruptible(t *testing.T) {
	r := NewResumable(nil)
	c := NewInterruptible(r)

	count := 0
	go func() {
		for {
			select {
			case <-c.Done():
				return
			case <-time.After(10 * time.Millisecond):
				count++
			}
		}
	}()

	<-time.After(55 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to be 5, received: %d", count)
	}

	r.Cancel()
	<-time.After(45 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to remain 5, received: %d", count)
	}
}

func TestParentPauseInterruptible(t *testing.T) {
	r := NewResumable(nil)
	c := NewInterruptible(r)

	count := 0
	go func() {
		for {
			select {
			case <-c.Done():
				return
			case <-time.After(10 * time.Millisecond):
				count++
			}
		}
	}()

	<-time.After(55 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to be 5, received: %d", count)
	}

	r.Pause()
	<-time.After(45 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to remain 5, received: %d", count)
	}
}

func TestParentResumeInterruptible(t *testing.T) {
	r := NewResumable(nil)
	c := NewInterruptible(r)
	r.Pause()

	count := 0
	go func() {
		for {
			if resume := <-r.Wait(); resume == OpCancel {
				return
			}

			<-time.After(10 * time.Millisecond)
			count++
		}
	}()

	count2 := 0
	go func() {
		for {
			if resume := <-c.Wait(); resume == OpCancel {
				return
			}

			<-time.After(10 * time.Millisecond)
			count2++
		}
	}()

	<-time.After(45 * time.Millisecond)
	if count != 0 {
		t.Errorf("Expected count to be 0, received: %d", count)
	}
	if count2 != 0 {
		t.Errorf("Expected child count to be 0, received: %d", count2)
	}

	r.Resume()
	<-time.After(55 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to be 5, received: %d", count)
	}
	if count2 != 0 {
		t.Errorf("Expected child count to remain 0, received: %d", count2)
	}
}

func TestParentCancelBeforeInterruptibleChild(t *testing.T) {
	r := NewResumable(nil)
	r.Cancel()
	c := NewInterruptible(r)

	count := 0
	go func() {
		for {
			select {
			case <-c.Done():
				return
			case <-time.After(10 * time.Millisecond):
				count++
			}
		}
	}()

	<-time.After(55 * time.Millisecond)
	if count != 0 {
		t.Errorf("Expected count to be 0, received: %d", count)
	}
}

func TestParentPauseBeforeInterruptibleChild(t *testing.T) {
	r := NewResumable(nil)
	r.Pause()
	c := NewInterruptible(r)

	count := 0
	go func() {
		for {
			<-time.After(10 * time.Millisecond)

			if resume := <-c.Wait(); resume == OpCancel {
				return
			}
			count++
		}
	}()

	<-time.After(45 * time.Millisecond)
	if count != 0 {
		t.Errorf("Expected count to be 0, received: %d", count)
	}
}

func TestParentResumeBeforeInterruptibleChild(t *testing.T) {
	r := NewResumable(nil)
	r.Pause()
	c := NewInterruptible(r)

	count := 0
	go func() {
		for {
			if resume := <-r.Wait(); resume == OpCancel {
				return
			}

			<-time.After(10 * time.Millisecond)
			count++
		}
	}()

	count2 := 0
	go func() {
		for {
			if resume := <-c.Wait(); resume == OpCancel {
				return
			}

			<-time.After(10 * time.Millisecond)
			count2++
		}
	}()

	<-time.After(45 * time.Millisecond)
	if count != 0 {
		t.Errorf("Expected count to be 0, received: %d", count)
	}
	if count2 != 0 {
		t.Errorf("Expected child count to be 0, received: %d", count2)
	}

	r.Resume()
	<-time.After(55 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to be 5, received: %d", count)
	}
	if count2 != 0 {
		t.Errorf("Expected child count to remain 0, received: %d", count2)
	}
}

func TestInterruptibleParentCancel(t *testing.T) {
	r := NewInterruptible(nil)
	c := NewResumable(r)

	count := 0
	go func() {
		for {
			select {
			case <-c.Done():
				return
			case <-time.After(10 * time.Millisecond):
				count++
			}
		}
	}()

	<-time.After(55 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to be 5, received: %d", count)
	}

	r.Cancel()
	<-time.After(45 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to remain 5, received: %d", count)
	}
}

func TestInterruptibleParentPause(t *testing.T) {
	r := NewInterruptible(nil)
	c := NewResumable(r)

	count := 0
	go func() {
		for {
			<-time.After(10 * time.Millisecond)

			if resume := <-c.Wait(); resume == OpCancel {
				return
			}
			count++
		}
	}()

	<-time.After(55 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to be 5, received: %d", count)
	}

	r.Pause()
	<-time.After(45 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to remain 5, received: %d", count)
	}
}

func TestInterruptibleParentResume(t *testing.T) {
	r := NewInterruptible(nil)
	c := NewResumable(r)
	r.Pause()

	// TODO: perhaps figure out a way to make cancel propogation syncronous
	// Give time for the cancel to propogate to the child
	<-time.After(1 * time.Millisecond)

	count := 0
	go func() {
		for {
			if resume := <-c.Wait(); resume == OpCancel {
				return
			}

			<-time.After(10 * time.Millisecond)
			count++
		}
	}()

	<-time.After(45 * time.Millisecond)
	if count != 0 {
		t.Errorf("Expected count to be 0, received: %d", count)
	}

	r.Resume()
	<-time.After(55 * time.Millisecond)
	if count != 0 {
		t.Errorf("Expected count to remain 0, received: %d", count)
	}
}

func TestInterruptibleParentCancelBeforeChild(t *testing.T) {
	r := NewInterruptible(nil)
	r.Cancel()
	c := NewResumable(r)

	count := 0
	go func() {
		for {
			select {
			case <-c.Done():
				return
			case <-time.After(10 * time.Millisecond):
				count++
			}
		}
	}()

	<-time.After(55 * time.Millisecond)
	if count != 0 {
		t.Errorf("Expected count to be 0, received: %d", count)
	}
}

func TestInterruptibleParentPauseBeforeChild(t *testing.T) {
	r := NewInterruptible(nil)
	r.Pause()
	c := NewResumable(r)

	count := 0
	go func() {
		for {
			<-time.After(10 * time.Millisecond)

			if resume := <-c.Wait(); resume == OpCancel {
				return
			}
			count++
		}
	}()

	<-time.After(45 * time.Millisecond)
	if count != 0 {
		t.Errorf("Expected count to be 0, received: %d", count)
	}
}

func TestInterruptibleParentResumeBeforeChild(t *testing.T) {
	r := NewInterruptible(nil)
	r.Pause()
	c := NewResumable(r)

	count := 0
	go func() {
		for {
			if resume := <-c.Wait(); resume == OpCancel {
				return
			}

			<-time.After(10 * time.Millisecond)
			count++
		}
	}()

	<-time.After(45 * time.Millisecond)
	if count != 0 {
		t.Errorf("Expected count to be 0, received: %d", count)
	}

	r.Resume()
	<-time.After(55 * time.Millisecond)
	if count != 0 {
		t.Errorf("Expected count to remain 0, received: %d", count)
	}
}

func TestInterruptibleParentCancelInterruptible(t *testing.T) {
	r := NewInterruptible(nil)
	c := NewInterruptible(r)

	count := 0
	go func() {
		for {
			select {
			case <-c.Done():
				return
			case <-time.After(10 * time.Millisecond):
				count++
			}
		}
	}()

	<-time.After(55 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to be 5, received: %d", count)
	}

	r.Cancel()
	<-time.After(45 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to remain 5, received: %d", count)
	}
}

func TestInterruptibleParentPauseInterruptible(t *testing.T) {
	r := NewInterruptible(nil)
	c := NewInterruptible(r)

	count := 0
	go func() {
		for {
			select {
			case <-c.Done():
				return
			case <-time.After(10 * time.Millisecond):
				count++
			}
		}
	}()

	<-time.After(55 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to be 5, received: %d", count)
	}

	r.Pause()
	<-time.After(45 * time.Millisecond)
	if count != 5 {
		t.Errorf("Expected count to remain 5, received: %d", count)
	}
}

func TestInterruptibleParentResumeInterruptible(t *testing.T) {
	r := NewInterruptible(nil)
	c := NewInterruptible(r)
	r.Pause()

	// TODO: perhaps figure out a way to make cancel propogation syncronous
	// Give time for the cancel to propogate to the child
	<-time.After(1 * time.Millisecond)

	count := 0
	go func() {
		for {
			if resume := <-r.Wait(); resume == OpCancel {
				return
			}

			<-time.After(10 * time.Millisecond)
			count++
		}
	}()

	count2 := 0
	go func() {
		for {
			if resume := <-c.Wait(); resume == OpCancel {
				return
			}

			<-time.After(10 * time.Millisecond)
			count2++
		}
	}()

	<-time.After(45 * time.Millisecond)
	if count != 0 {
		t.Errorf("Expected count to be 0, received: %d", count)
	}
	if count2 != 0 {
		t.Errorf("Expected child count to be 0, received: %d", count2)
	}

	r.Resume()
	<-time.After(55 * time.Millisecond)
	if count != 0 {
		t.Errorf("Expected count to be 5, received: %d", count)
	}
	if count2 != 0 {
		t.Errorf("Expected child count to remain 0, received: %d", count2)
	}
}

func TestInterruptibleParentCancelBeforeInterruptibleChild(t *testing.T) {
	r := NewInterruptible(nil)
	r.Cancel()
	c := NewInterruptible(r)

	count := 0
	go func() {
		for {
			select {
			case <-c.Done():
				return
			case <-time.After(10 * time.Millisecond):
				count++
			}
		}
	}()

	<-time.After(55 * time.Millisecond)
	if count != 0 {
		t.Errorf("Expected count to be 0, received: %d", count)
	}
}

func TestInterruptibleParentPauseBeforeInterruptibleChild(t *testing.T) {
	r := NewInterruptible(nil)
	r.Pause()
	c := NewInterruptible(r)

	count := 0
	go func() {
		for {
			<-time.After(10 * time.Millisecond)

			if resume := <-c.Wait(); resume == OpCancel {
				return
			}
			count++
		}
	}()

	<-time.After(45 * time.Millisecond)
	if count != 0 {
		t.Errorf("Expected count to be 0, received: %d", count)
	}
}

func TestInterruptibleParentResumeBeforeInterruptibleChild(t *testing.T) {
	r := NewInterruptible(nil)
	r.Pause()
	c := NewInterruptible(r)

	count := 0
	go func() {
		for {
			if resume := <-r.Wait(); resume == OpCancel {
				return
			}

			<-time.After(10 * time.Millisecond)
			count++
		}
	}()

	count2 := 0
	go func() {
		for {
			if resume := <-c.Wait(); resume == OpCancel {
				return
			}

			<-time.After(10 * time.Millisecond)
			count2++
		}
	}()

	<-time.After(45 * time.Millisecond)
	if count != 0 {
		t.Errorf("Expected count to be 0, received: %d", count)
	}
	if count2 != 0 {
		t.Errorf("Expected child count to be 0, received: %d", count2)
	}

	r.Resume()
	<-time.After(55 * time.Millisecond)
	if count != 0 {
		t.Errorf("Expected count to be 5, received: %d", count)
	}
	if count2 != 0 {
		t.Errorf("Expected child count to remain 0, received: %d", count2)
	}
}
