package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
)

func handleForwardedPorts(tun *tunnelCon) error {
	for _, listener := range listeners {
		listener.Close()
	}

	//Forwarded port count
	portCounts, err := binary.ReadUvarint(tun.frameReader)
	if err != nil {
		return fmt.Errorf("unable to read forwarded port count: %v", err)
	}

	//Build port list data
	forwardedPorts = []int{}
	portsStr := ""
	listeners = []*net.UDPConn{}
	for p := range portCounts {

		//Read port
		port, err := binary.ReadUvarint(tun.frameReader)
		if err != nil {
			return fmt.Errorf("unable to read forwarded port: %v", err)
		}

		//Name length
		nameLen, err := binary.ReadUvarint(tun.frameReader)
		if err != nil {
			return fmt.Errorf("unable to read name length: %v", err)
		}

		//Read name
		var name []byte
		if nameLen > 0 {
			name = make([]byte, nameLen)
			l, err := io.ReadFull(tun.frameReader, name)
			if err != nil {
				return fmt.Errorf("Unable to read frame data: %v", err)
			}
			if l != int(nameLen) {
				return fmt.Errorf("Unable to read all frame data: %v of %v", l, nameLen)
			}
		}

		//Build list
		forwardedPorts = append(forwardedPorts, int(port))
		forwardedPortsNames = append(forwardedPortsNames, string(name))
		if p != 0 {
			portsStr = portsStr + ", "
		}
		if nameLen > 0 {
			portsStr = portsStr + string(name) + " - "
		}
		portsStr = portsStr + strconv.FormatUint(port, 10)

		//Add listener
		laddr := &net.UDPAddr{IP: nil, Port: int(port)}
		conn, err := net.ListenUDP("udp", laddr)
		if err != nil {
			return fmt.Errorf("unable to read from laddr: %v", err)
		}
		listeners = append(listeners, conn)
	}

	doLog("Forwarded ports: %v", portsStr)
	outputServerList()
	startServerListUpdater()

	return nil
}

func startServerListUpdater() {
	htmlUpdaterOnce.Do(func() {
		go func() {
			for {
				outputServerList()
				ephemeralLock.Lock()
				active := len(ephemeralIDMap) > 0
				ephemeralLock.Unlock()
				if active {
					time.Sleep(htmlUpdateActive)
					continue
				}
				loops := int(htmlUpdateIdle / htmlUpdateActive)
				for i := 0; i < loops; i++ {
					time.Sleep(htmlUpdateActive)
					ephemeralLock.Lock()
					if len(ephemeralIDMap) > 0 {
						active = true
					}
					ephemeralLock.Unlock()
					if active {
						break
					}
				}
			}
		}()
	})
}

func outputServerList() {
	ephemeralLock.Lock()
	current := len(ephemeralIDMap)
	peak := ephemeralPeak
	total := ephemeralSessionsTotal
	sessions := []SessionInfo{}
	for _, s := range ephemeralIDMap {
		sess := SessionInfo{
			ID:          s.id,
			DestPort:    s.destPort,
			Duration:    time.Since(s.startTime).Round(time.Second).String(),
			BytesIn:     s.bytesIn,
			BytesOut:    s.bytesOut,
			BytesInStr:  humanize.Bytes(uint64(s.bytesIn)),
			BytesOutStr: humanize.Bytes(uint64(s.bytesOut)),
		}
		sessions = append(sessions, sess)
	}
	inTotal := bytesInTotal
	outTotal := bytesOutTotal
	ephemeralLock.Unlock()

	data := PageData{
		Servers:          []ServerEntry{},
		CurrentUsers:     current,
		PeakUsers:        peak,
		TotalSessions:    total,
		Uptime:           time.Since(startTime).Round(time.Second).String(),
		Version:          version,
		Protocol:         protocolVersion,
		BatchInterval:    batchingMicroseconds,
		Compression:      compressionLevel,
		Sessions:         sessions,
		BytesInTotal:     inTotal,
		BytesOutTotal:    outTotal,
		BytesInTotalStr:  humanize.Bytes(uint64(inTotal)),
		BytesOutTotalStr: humanize.Bytes(uint64(outTotal)),
	}

	for i, port := range forwardedPorts {
		name := forwardedPortsNames[i]
		server := ServerEntry{Name: name, Addr: clientAddress, Port: port}
		data.Servers = append(data.Servers, server)
	}

	htmlFileName := privateIndexFilename
	parsedTemplate := privateServerTemplate
	if publicMode {
		parsedTemplate = publicServerTemplate
		htmlFileName = publicIndexFilename
	}

	tmpName := htmlFileName + ".tmp"
	f, err := os.Create(tmpName)
	if err != nil {
		doLog("Failed to create file: %v", err)
		os.Exit(1)
	}
	defer f.Close()

	err = parsedTemplate.Execute(f, data)
	if err != nil {
		doLog("Failed to execute template: %v", err)
		os.Exit(1)
	}

	if err := os.Rename(tmpName, htmlFileName); err != nil {
		doLog("Failed to rename output: %v", err)
		os.Exit(1)
	}

	saveStats(data)

	doLog("%v written successfully.", htmlFileName)

	if publicMode {
		if err := openInBrowser(htmlFileName); err != nil {
			doLog("Failed to open in browser: %v", err)
		}
	}
}

func openInBrowser(path string) error {
	if notFirstConnect {
		return nil
	}
	notFirstConnect = true

	var cmd *exec.Cmd

	doLog("Opening link: %v", path)

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default: // linux, freebsd, etc.
		cmd = exec.Command("xdg-open", path)
	}

	return cmd.Start()
}
