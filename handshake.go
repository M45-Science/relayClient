package main

import (
	"encoding/binary"
	"fmt"
)

func writeHandshakePacket(tun *tunnelCon) {
	var buf []byte
	buf = binary.AppendUvarint(buf, CLIENT_KEY)
	buf = binary.AppendUvarint(buf, protocolVersion)
	buf = binary.AppendUvarint(buf, 0)
	tun.con.Write(buf)
}

func readHandshakePacket(tun *tunnelCon) error {
	//Read key
	key, err := binary.ReadUvarint(tun.frameReader)
	if err != nil {
		return fmt.Errorf("unable to read key: %v", err)
	}
	if key != SERVER_KEY {
		return fmt.Errorf("incorrect key: %X", key)
	}

	//Protocol version
	proto, err := binary.ReadUvarint(tun.frameReader)
	if err != nil {
		return fmt.Errorf("unable to read protocol version: %v", err)
	}
	if proto != protocolVersion {
		doLog("protocol version not compatible: %v.", proto)
		doLog("Please download a new version from %v", downloadURL)
		openInBrowser(downloadURL)
		select {}
	}

	//Compression level
	cl, err := binary.ReadUvarint(tun.frameReader)
	if err != nil {
		return fmt.Errorf("unable to read compressionLevel: %v", err)
	}
	compressionLevel = int(cl)
	if compressionLevel > MaxCompression {
		compressionLevel = MaxCompression
	}

	//Batching Interval
	bm, err := binary.ReadUvarint(tun.frameReader)
	if err != nil {
		return fmt.Errorf("unable to read server id: %v", err)
	}
	batchingMicroseconds = int(bm)

	//Reserved Value A
	rva, err := binary.ReadUvarint(tun.frameReader)
	if err != nil {
		return fmt.Errorf("unable to read server id: %v", err)
	}
	reservedValueA = int(rva)

	//Reserved Value B
	rvb, err := binary.ReadUvarint(tun.frameReader)
	if err != nil {
		return fmt.Errorf("unable to read server id: %v", err)
	}
	reservedValueB = int(rvb)

	//Server ID
	sid, err := binary.ReadUvarint(tun.frameReader)
	if err != nil {
		return fmt.Errorf("unable to read server id: %v", err)
	}
	serverID = int(sid)

	if debugLog {
		doLog("Proto: %v, Compress: %v, Batch: %v, ServerID: %v", proto, compressionLevel, batchingMicroseconds, serverID)
	}

	//Forwarded port count
	forwardedPortCount, err := binary.ReadUvarint(tun.frameReader)
	if err != nil {
		return fmt.Errorf("unable to read forwarded port count: %v", err)
	}

	err = handleForwardedPorts(tun, int(forwardedPortCount))
	if err != nil {
		return fmt.Errorf("handleForwardedPorts: %v", err)
	}

	return nil
}
