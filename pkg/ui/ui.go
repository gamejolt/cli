package ui

import "gopkg.in/fatih/color.v1"

var (
	// ErrorCol is the color for errors
	ErrorCol = color.New(color.FgHiRed, color.Bold).FprintfFunc()

	// PromptCol is the color for prompts
	PromptCol = color.New(color.FgHiWhite, color.Bold).FprintfFunc()

	// SuccessCol is the color for sucess
	SuccessCol = color.New(color.FgHiGreen, color.Bold).FprintfFunc()

	// InfoCol is the color for prompts
	InfoCol = color.New(color.FgHiWhite, color.Bold).FprintfFunc()

	// WarnCol is the color for warnings
	WarnCol = color.New(color.FgHiYellow, color.Bold).FprintfFunc()
)

// Error prints an error
func Error(format string, a ...interface{}) {
	ErrorCol(color.Output, format, a...)
}

// Prompt prints a prompt
func Prompt(format string, a ...interface{}) {
	PromptCol(color.Output, format, a...)
}

// Success prints a success message
func Success(format string, a ...interface{}) {
	SuccessCol(color.Output, format, a...)
}

// Info prints an info message
func Info(format string, a ...interface{}) {
	InfoCol(color.Output, format, a...)
}

// Warn prints a warning message
func Warn(format string, a ...interface{}) {
	WarnCol(color.Output, format, a...)
}
