//go:build windows
// +build windows

package main

import (
	"os"

	"golang.org/x/sys/windows"
)

func enableVirtualTerminalProcessing() error {
	// Get the handle to standard output
	handle := windows.Handle(os.Stdout.Fd())

	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		return err
	}

	// Enable ENABLE_VIRTUAL_TERMINAL_PROCESSING
	const ENABLE_VIRTUAL_TERMINAL_PROCESSING = 0x0004
	mode |= ENABLE_VIRTUAL_TERMINAL_PROCESSING

	return windows.SetConsoleMode(handle, mode)
}
