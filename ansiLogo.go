package main

import (
	"fmt"
	"strings"
)

func showANSILogo() {
	if enableVirtualTerminalProcessing() == nil {
		logo := strings.ReplaceAll(string(logoANSI), "\\e", "\x1b")
		fmt.Println(logo + "\nM45-Science")
	}
}
