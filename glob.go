package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"net"
	"sync"
	"time"
)

const (
	protocolVersion = 1
	SERVER_KEY      = 0xCAFE69C0FFEE
	CLIENT_KEY      = 0xADD069C0FFEE

	publicReconDelaySec  = 2
	privateReconDelaySec = 5
	maxAttempts          = 60 / publicReconDelaySec
	attemptResetAfter    = time.Minute * 5
	ephemeralLife        = time.Minute
	ephemeralTicker      = time.Second * 15

	bufferSizeUDP = 65 * 1024
	downloadURL   = "https://m45sci.xyz/eu#downloads"
)

var htmlFileName = "connect-links.html"

var MaxCompression = len(compressionLevels) - 1

var compressionLevels []int = []int{
	gzip.NoCompression,
	gzip.HuffmanOnly,
	gzip.BestSpeed,
	gzip.DefaultCompression,
	gzip.BestCompression,
}

var (
	serverID             int
	tunnelServerAddr     string
	clientAddr           string
	forwardedPorts       []int
	forwardedPortsNames  []string
	listeners            []*net.UDPConn
	batchingMicroseconds int
	compressionLevel     int = defaultCompression
	reservedValueA       int
	reservedValueB       int

	PublicClientMode      string
	verboseLog, forceHTML bool

	ephemeralTop     int                       = 1
	ephemeralIDMap   map[int]*ephemeralData    = map[int]*ephemeralData{}
	ephemeralPortMap map[string]*ephemeralData = map[string]*ephemeralData{}
	ephemeralLock    sync.Mutex
)

type tunnelCon struct {
	frameReader  *bufio.Reader
	packetReader *bytes.Reader
	con          net.Conn

	packets       []byte
	packetsLength int
	packetLock    sync.Mutex
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

type AddressData struct {
	ip   string
	port int
}
