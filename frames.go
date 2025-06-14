package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

func frameHandler(tun *tunnelCon) error {

	err := readHandshakePacket(tun)
	if err != nil {
		return fmt.Errorf("Unable to read handshake: %v", err)
	}
	go tun.batchWriter()
	go handleListeners(tun)

	tun.readFrames()

	return nil
}

func (tun *tunnelCon) readFrames() {
	err := readFrameHeader(tun)
	if err != nil {
		if debugLog {
                   doLog("%v", err)
		}
		return
	}

	for tun.packetReader != nil && tun.packetReader.Len() > 0 {
		err := tun.readPacket()
		if err != nil {
			if debugLog {
                               doLog("%v", err)
			}
			return
		}
	}

	tun.lastUsed = time.Now()

	tun.readFrames()
}

func readFrameHeader(tun *tunnelCon) error {
	frameLength, err := binary.ReadUvarint(tun.frameReader)
	if err != nil {
		return fmt.Errorf("unable to read frameLength: %v", err)
	}
	frameData := make([]byte, frameLength)
	l, err := io.ReadFull(tun.frameReader, frameData)
	if err != nil {
		return fmt.Errorf("Unable to read frame data: %v", err)
	}
	if l != int(frameLength) {
		return fmt.Errorf("Unable to read all frame data: %v of %v", l, frameLength)
	}

	if compressionLevel > 0 {
		frameData, err = decompressFrame(frameData)
		if err != nil {
			return err
		}
	}

	tun.packetReader = bytes.NewReader(frameData)
	return nil
}
