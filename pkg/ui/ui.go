package ui

import "gopkg.in/fatih/color.v1"

var (
	// Error is the color for errors
	Error = color.New(color.FgHiRed, color.Bold).PrintfFunc()

	// Prompt is the color for prompts
	Prompt = color.New(color.FgHiWhite, color.Bold).PrintfFunc()

	// Success is the color for sucess
	Success = color.New(color.FgHiGreen, color.Bold).PrintfFunc()

	// Info is the color for prompts
	Info = color.New(color.FgHiWhite, color.Bold).PrintfFunc()

	// Warn is the color for warnings
	Warn = color.New(color.FgHiYellow, color.Bold).PrintfFunc()
)
