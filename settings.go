package main

import (
	"compress/gzip"
	"time"
)

const (
	protocolVersion = 1
	SERVER_KEY      = 0xCAFE69C0FFEE
	CLIENT_KEY      = 0xADD069C0FFEE

	publicReconDelaySec  = 2
	privateReconDelaySec = 5

	ephemeralLife   = time.Minute
	ephemeralTicker = time.Second * 15

	maxAttempts       = 60 / publicReconDelaySec
	attemptResetAfter = time.Minute * 5
	tunnelLife        = time.Minute

	bufferSizeUDP = 65 * 1024
	downloadURL   = "https://m45sci.xyz/eu#downloads"
	updateLink    = "https://m45sci.xyz/relayClient.json"
)

var (
	MaxCompression = len(compressionLevels) - 1
)

var compressionLevels []int = []int{
	gzip.NoCompression,
	gzip.HuffmanOnly,
	gzip.BestSpeed,
	gzip.DefaultCompression,
	gzip.BestCompression,
}
