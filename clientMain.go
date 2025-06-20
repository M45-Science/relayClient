package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func main() {
	startTime = time.Now()
	loadSavedStats()
	startStatsUpdater()
	go func() {
		err := restoreBinaryName()
		if err != nil {
			fmt.Println(err.Error())
		}
	}()

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

	startLog()
	go autoRotateLogs()
	showANSILogo()
	doLog("[START] goRelay client started: %v", version)

	if publicMode {
		didUpdate, err := CheckUpdate()
		if err != nil {
			doLog("CheckUpdate: %v", err)
		}

		if didUpdate {
			os.Exit(0)
		}
	}

	go tunnelHandler()
	go cleanEphemeralMaps()

	<-sigs
	doLog("[QUIT] relayClient Shutting down.")
}

func restoreBinaryName() error {
	// 1) Locate current executable
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find executable path: %w", err)
	}
	dir := filepath.Dir(exePath)
	base := filepath.Base(exePath) // e.g. "update_binary.exe" or "myprog"
	ext := filepath.Ext(base)      // e.g. ".exe" or ""

	// 2) Check if it's the updater name
	nameOnly := strings.TrimSuffix(base, ext)
	if nameOnly != "update_binary" {
		// nothing to do
		//fmt.Println("not an update")
		return nil
	}

	// 3) Compute the target name
	targetName := "M45-Relay-Client" + ext
	targetPath := filepath.Join(dir, targetName)

	// Sleep, just in case
	time.Sleep(time.Second * 2)

	// 4) Perform rename
	if err := os.Rename(exePath, targetPath); err != nil {
		return fmt.Errorf("failed to rename %q to %q: %w", exePath, targetPath, err)
	}

	return nil
}
