package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

func tunnelConnect() {
	ephemeralLock.Lock()
	ephemeralTop = 1
	ephemeralIDMap = map[int]*ephemeralData{}
	ephemeralPortMap = map[string]*ephemeralData{}
	ephemeralIDRecycle = []int{}
	ephemeralIDRecycleLen = 0
	ephemeralLock.Unlock()
	connectTunnel()
}

func connectHandler() {

	if publicMode {
		tunnelConnect()
	} else {
		for {
			tunnelConnect()
			time.Sleep(time.Second * privateReconDelaySec)
		}
	}
}

func connectTunnel() {
	doLog("[Connecting] to %v...", tunnelServerAddr)
	var err error
	con, err := net.Dial("tcp", tunnelServerAddr)
	if err != nil {
		doLog("Unable to connect: %v", err)
		return
	}

	tun := &tunnelCon{con: con, frameReader: bufio.NewReader(con), lastUsed: time.Now()}
	writeHandshakePacket(tun)

	err = frameHandler(tun)
	if err != nil {
		doLog("frameHandler: %v", err)
	}
	tun.delete()
}

func (tun *tunnelCon) delete() {
	doLog("[Disconnected]")
	tun.con.Close()
	tun.con = nil
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
		return fmt.Errorf("failed reading whole packet from tunnel: read %vb of %vb", l, payloadLen)
	}

	//Lookup destination via ID
	ephemeralLock.Lock()
	dest := ephemeralIDMap[int(sessionID)]
	if dest != nil {
		dest.lastUsed = time.Now()
	}
	ephemeralLock.Unlock()

	if dest == nil {
		return fmt.Errorf("received response for invalid ID: %v", sessionID)
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
			return fmt.Errorf("only wrote %vb of %vb to %v", w, payloadLen, dest.destPort)
		}
		if verboseDebug {
			doLog("wrote %vb to %v", w, dest.destPort)
		}
	}

	return nil
}
