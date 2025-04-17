package main

import (
	"bufio"
	"bytes"
	"net"
	"sync"
	"time"
)

type tunnelCon struct {
	frameReader  *bufio.Reader
	packetReader *bytes.Reader
	con          net.Conn

	packets       []byte
	packetsLength int
	packetLock    sync.Mutex

	lastUsed time.Time
}

type ephemeralData struct {
	id       int
	source   string
	destPort int
	lastUsed time.Time
	listener *net.UDPConn
}

type ServerEntry struct {
	Name string
	Addr string
	Port int
}

type PageData struct {
	Servers []ServerEntry
}
