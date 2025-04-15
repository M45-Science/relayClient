package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

func connectHandler() {

	if PublicClientMode == "true" {
		lastConnect := time.Now()
		for attempts := 0; attempts < maxAttempts; attempts++ {
			if attempts != 0 {
				time.Sleep(time.Duration(publicReconDelaySec) * time.Second)
			}

			ephemeralLock.Lock()
			ephemeralTop = 1
			ephemeralIDMap = map[int]*ephemeralData{}
			ephemeralPortMap = map[string]*ephemeralData{}
			ephemeralLock.Unlock()
			connectTunnel()

			//Eventually reset tries
			if time.Since(lastConnect) > attemptResetAfter {
				attempts = 0
			}
			lastConnect = time.Now()
		}

		log.Printf("Too many unsuccsessful connection attempts (%v), stopping.\nQuit then relaunch to try again.", maxAttempts)
		time.Sleep(time.Hour * 24)
	} else {
		for {
			ephemeralLock.Lock()
			ephemeralTop = 1
			ephemeralIDMap = map[int]*ephemeralData{}
			ephemeralPortMap = map[string]*ephemeralData{}
			ephemeralLock.Unlock()
			connectTunnel()

			time.Sleep(time.Second * privateReconDelaySec)
		}
	}
}

func connectTunnel() {
	log.Printf("Connecting to %v...", tunnelServerAddr)
	var err error
	con, err := net.Dial("tcp", tunnelServerAddr)
	if err != nil {
		log.Printf("Unable to connect: %v", err)
		return
	}

	tun := &tunnelCon{con: con, frameReader: bufio.NewReader(con)}
	writeHandshakePacket(tun)

	err = frameHandler(tun)
	if err != nil {
		log.Printf("readFrame: %v", err)
	}
}

func (tun *tunnelCon) readPacket() error {
	sessionID, err := binary.ReadUvarint(tun.packetReader)
	if err != nil {
		return fmt.Errorf("unable to read sessionID: %v", err)
	}

	payloadLen, err := binary.ReadUvarint(tun.packetReader)
	if err != nil {
		return fmt.Errorf("unable to read payload length: %v", err)
	}

	//Read in full payload
	var payload = make([]byte, payloadLen)
	l, err := io.ReadFull(tun.packetReader, payload)
	if err != nil {

		return fmt.Errorf("unable to read payload: %v", err)
	}
	if payloadLen != uint64(l) {
		return fmt.Errorf("Failed reading whole packet from tunnel: read %v of %v", l, payloadLen)
	}

	if verboseLog {
		log.Printf("SessionID: %v, expect: %v got %v", sessionID, payloadLen, l)
	}

	//Lookup destination via ID
	ephemeralLock.Lock()
	dest := ephemeralIDMap[int(sessionID)]
	ephemeralLock.Unlock()

	if dest == nil {
		return fmt.Errorf("Received response for invalid ID: %v", sessionID)
	} else {
		addr, err := net.ResolveUDPAddr("udp", dest.source)
		if err != nil {
			return fmt.Errorf("unable to resolve destination: %v", err)
		}

		w, err := dest.listener.WriteToUDP(payload, addr)
		if err != nil {
			return fmt.Errorf("unable to write payload: %v", err)
		}
		if w != int(payloadLen) {
			return fmt.Errorf("Only wrote %vb of %vb to %v", w, payloadLen, dest.destPort)
		}
		if verboseLog {
			log.Printf("Wrote %vb to %v", w, dest.destPort)
		}
	}

	return nil
}
