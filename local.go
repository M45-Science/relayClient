package main

import (
	"encoding/binary"
	"net"
	"strconv"
	"strings"
	"time"
)

// TO DO: Add contexts and deadlines
func handleListeners(tun *tunnelCon) {
	for _, port := range listeners {
		go func(p *net.UDPConn) {
			for p != nil {
				// Read payload
				buf := make([]byte, bufferSizeUDP)
				n, addr, err := p.ReadFromUDP(buf)
				if err != nil {
					doLog("Error reading: %v", err)
					return
				}

				// Check ephemeral map
				ephemeralLock.Lock()
				var newSession *ephemeralData
				session := ephemeralPortMap[addr.String()]

				// New session, create
				if session == nil {
					newSession = &ephemeralData{
						id: ephemeralTop, source: addr.String(),
						destPort: getPortStr(p.LocalAddr().String()),
						lastUsed: time.Now(), listener: port}

					ephemeralPortMap[addr.String()] = newSession
					ephemeralIDMap[ephemeralTop] = newSession

					ephemeralTop++
					session = newSession
					doLog("NEW SESSION ID: %v: %vb: %v -> %v\n", newSession.id, n, newSession.source, newSession.destPort)
				} else {
					session.lastUsed = time.Now()
					if verboseLog {
						doLog("Session ID: %v: %vb: %v -> %v\n", session.id, n, session.source, session.destPort)
					}
				}
				ephemeralLock.Unlock()

				/* New client, tell server clientID destination */
				var header []byte
				if newSession != nil {
					header = binary.AppendUvarint(header, 0)
					header = binary.AppendUvarint(header, uint64(newSession.destPort))
				}

				//Write standard header
				header = binary.AppendUvarint(header, uint64(session.id))
				header = binary.AppendUvarint(header, uint64(n))
				tun.Write(append(header, buf[:n]...))
			}
		}(port)
	}
}

// Get port from address string
func getPortStr(input string) int {
	parts := strings.Split(input, ":")
	numparts := len(parts)
	portStr := parts[numparts-1]
	port, _ := strconv.ParseUint(portStr, 10, 64)
	return int(port)
}
