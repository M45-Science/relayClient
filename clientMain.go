package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	flag.StringVar(&tunnelServerAddress, "serverAddr", "m45sci.xyz:30000", "server:port")
	flag.StringVar(&clientAddress, "clientAddr", "127.0.0.1", "address for this proxy")
	flag.BoolVar(&debugLog, "debugLog", false, "debug logging")
	flag.BoolVar(&verboseDebug, "verboseDebug", false, "full debug logging")
	flag.Parse()
	if verboseDebug {
		//Verbose also enables debug
		debugLog = true
	}
	if publicClientFlag == "true" {
		//Convert ldflag to bool
		publicMode = true
	}

	_, err := CheckUpdate()
	if err != nil {
		log.Println(err)
	}

	startLog()
	go autoRotateLogs()
	showANSILogo()
	doLog("[START] goRelay client started.")

	go tunnelHandler()
	go cleanEphemeralMaps()

	<-sigs
	doLog("[QUIT] relayClient Shutting down.")
}
