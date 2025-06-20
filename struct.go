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
	id        int
	source    string
	destPort  int
	lastUsed  time.Time
	listener  *net.UDPConn
	startTime time.Time
	bytesIn   int64
	bytesOut  int64
}

type ServerEntry struct {
	Name string
	Addr string
	Port int
}

type PageData struct {
	Servers          []ServerEntry
	CurrentUsers     int
	PeakUsers        int
	TotalSessions    int
	Uptime           string
	Version          string
	Protocol         int
	BatchInterval    int
	Compression      int
	Sessions         []SessionInfo
	BytesInTotal     int64
	BytesOutTotal    int64
	BytesInTotalStr  string
	BytesOutTotalStr string
}

type SessionInfo struct {
	ID          int
	DestPort    int
	Duration    string
	BytesIn     int64
	BytesOut    int64
	BytesInStr  string
	BytesOutStr string
}

type SavedStats struct {
	StartTime     time.Time
	SessionsTotal int
	PeakUsers     int
	BytesInTotal  int64
	BytesOutTotal int64
}
