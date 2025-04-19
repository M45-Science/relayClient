package main

import (
	"net"
	"sync"
)

var (
	serverID             int
	tunnelServerAddress  string
	clientAddress        string
	forwardedPorts       []int
	forwardedPortsNames  []string
	listeners            []*net.UDPConn
	batchingMicroseconds int
	compressionLevel     int = 1
	reservedValueA       int
	reservedValueB       int

	debugLog, verboseDebug bool

	ephemeralTop          int = 1
	ephemeralIDRecycle    []int
	ephemeralIDRecycleLen int

	ephemeralIDMap   map[int]*ephemeralData    = map[int]*ephemeralData{}
	ephemeralPortMap map[string]*ephemeralData = map[string]*ephemeralData{}
	ephemeralLock    sync.Mutex

	publicMode       bool
	publicClientFlag string
	version          string
	notFirstConnect  bool
)
