package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
)

const (
	meep               = 1
	defaultCompression = 1
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	flag.StringVar(&tunnelServerAddr, "serverAddr", "m45sci.xyz:30000", "server:port")
	flag.StringVar(&clientAddr, "clientAddr", "127.0.0.1", "address for this proxy")
	flag.BoolVar(&forceHTML, "openList", false, "Write "+htmlFileName+" and then attempt to open it.")
	flag.BoolVar(&verboseLog, "verboseLog", false, "debug logging")
	flag.Parse()

	if runtime.GOOS != "windows" {
		logo := strings.ReplaceAll(logoANSI, "\\e", "\x1b")
		fmt.Println(logo + "\nM45-Science")
	}
	startLog()
	autoRotateLogs()
	doLog("[START] goRelay client started.")

	go connectHandler()
	go cleanEphemeralMaps()

	<-sigs
	doLog("[QUIT] Server shutting down: Signal: %v", sigs)
}

const logoANSI = `\e[48;5;0m          \e[48;5;254m  \e[48;5;255m  \e[48;5;243m  \e[48;5;242m  \e[48;5;0m                        \e[m
\e[48;5;0m        \e[48;5;255m  \e[48;5;246m  \e[48;5;233m  \e[48;5;145m  \e[48;5;253m  \e[48;5;8m  \e[48;5;0m                      \e[m
\e[48;5;0m        \e[48;5;237m  \e[48;5;145m  \e[48;5;188m  \e[48;5;245m  \e[48;5;250m  \e[48;5;24m  \e[48;5;81m  \e[48;5;153m  \e[48;5;23m  \e[48;5;0m                \e[m
\e[48;5;0m        \e[48;5;235m  \e[48;5;253m  \e[48;5;1m  \e[48;5;52m  \e[48;5;255m  \e[48;5;232m  \e[48;5;15m  \e[48;5;153m  \e[48;5;38m  \e[48;5;0m                \e[m
\e[48;5;0m        \e[48;5;248m  \e[48;5;236m  \e[48;5;23m  \e[48;5;66m  \e[48;5;255m  \e[48;5;74m  \e[48;5;189m  \e[48;5;44m  \e[48;5;246m  \e[48;5;24m  \e[48;5;0m              \e[m
\e[48;5;0m        \e[48;5;8m  \e[48;5;217m  \e[48;5;101m  \e[48;5;0m  \e[48;5;237m  \e[48;5;80m  \e[48;5;45m  \e[48;5;235m  \e[48;5;0m  \e[48;5;232m  \e[48;5;23m  \e[48;5;233m  \e[48;5;0m          \e[m
\e[48;5;0m      \e[48;5;52m  \e[48;5;131m  \e[48;5;132m  \e[48;5;160m  \e[48;5;52m  \e[48;5;234m  \e[48;5;24m  \e[48;5;239m  \e[48;5;0m  \e[48;5;80m  \e[48;5;23m  \e[48;5;31m  \e[48;5;251m  \e[48;5;87m  \e[48;5;232m  \e[48;5;233m  \e[48;5;232m  \e[48;5;0m  \e[m
\e[48;5;0m  \e[48;5;52m    \e[48;5;131m  \e[48;5;167m  \e[48;5;168m  \e[48;5;167m  \e[48;5;88m  \e[48;5;124m  \e[48;5;167m  \e[48;5;131m  \e[48;5;52m  \e[48;5;0m      \e[48;5;23m    \e[48;5;235m  \e[48;5;0m    \e[48;5;23m  \e[m
\e[48;5;0m  \e[48;5;234m  \e[48;5;131m  \e[48;5;167m  \e[48;5;124m  \e[48;5;181m  \e[48;5;167m  \e[48;5;124m  \e[48;5;88m  \e[48;5;161m  \e[48;5;167m  \e[48;5;131m  \e[48;5;52m  \e[48;5;0m      \e[48;5;236m  \e[48;5;23m    \e[48;5;232m  \e[48;5;23m  \e[m
\e[48;5;240m  \e[48;5;236m  \e[48;5;168m    \e[48;5;124m  \e[48;5;181m  \e[48;5;167m  \e[48;5;124m  \e[48;5;88m  \e[48;5;160m  \e[48;5;210m  \e[48;5;131m  \e[48;5;235m  \e[48;5;117m  \e[48;5;153m  \e[48;5;74m  \e[48;5;195m  \e[48;5;233m  \e[48;5;0m  \e[48;5;24m  \e[48;5;0m  \e[m
\e[48;5;235m  \e[48;5;1m  \e[48;5;88m  \e[48;5;1m  \e[48;5;167m      \e[48;5;124m  \e[48;5;52m    \e[48;5;88m  \e[48;5;236m    \e[48;5;117m  \e[48;5;24m  \e[48;5;234m  \e[48;5;238m  \e[48;5;232m  \e[48;5;23m    \e[48;5;74m  \e[m
\e[48;5;0m  \e[48;5;1m  \e[48;5;124m  \e[48;5;52m  \e[48;5;88m  \e[48;5;1m  \e[48;5;125m  \e[48;5;1m  \e[48;5;52m    \e[48;5;88m  \e[48;5;1m  \e[48;5;95m  \e[48;5;31m  \e[48;5;24m  \e[48;5;23m    \e[48;5;74m  \e[48;5;0m  \e[48;5;24m  \e[48;5;32m  \e[m
\e[48;5;0m    \e[48;5;238m  \e[48;5;1m  \e[48;5;52m            \e[48;5;1m  \e[48;5;52m  \e[48;5;236m  \e[48;5;153m  \e[48;5;81m  \e[48;5;38m  \e[48;5;74m  \e[48;5;234m  \e[48;5;0m      \e[m
\e[48;5;0m      \e[48;5;237m  \e[48;5;52m    \e[48;5;237m  \e[48;5;234m  \e[48;5;52m  \e[48;5;124m  \e[48;5;233m  \e[48;5;0m  \e[48;5;236m  \e[48;5;23m    \e[48;5;31m  \e[48;5;38m  \e[48;5;0m        \e[m`
