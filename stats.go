package main

import (
	"encoding/json"
	"os"
	"time"

	"github.com/dustin/go-humanize"
)

func loadSavedStats() {
	f, err := os.Open(statsFilename)
	if err != nil {
		if !os.IsNotExist(err) {
			doLog("loadSavedStats: %v", err)
		}
		return
	}
	defer f.Close()
	var s SavedStats
	if err := json.NewDecoder(f).Decode(&s); err != nil {
		doLog("loadSavedStats: %v", err)
		return
	}
	ephemeralLock.Lock()
	if !s.StartTime.IsZero() {
		startTime = s.StartTime
	}
	if s.PeakUsers > 0 {
		ephemeralPeak = s.PeakUsers
	}
	if s.SessionsTotal > 0 {
		ephemeralSessionsTotal = s.SessionsTotal
	}
	bytesInTotal = s.BytesInTotal
	bytesOutTotal = s.BytesOutTotal
	ephemeralLock.Unlock()
}

func saveStats(data PageData) {
	tmp := statsFilename + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		doLog("saveStats: %v", err)
		return
	}
	enc := json.NewEncoder(f)
	s := SavedStats{
		StartTime:     startTime,
		SessionsTotal: data.TotalSessions,
		PeakUsers:     data.PeakUsers,
		BytesInTotal:  data.BytesInTotal,
		BytesOutTotal: data.BytesOutTotal,
	}
	if err := enc.Encode(&s); err != nil {
		doLog("saveStats: %v", err)
		f.Close()
		return
	}
	f.Close()
	if err := os.Rename(tmp, statsFilename); err != nil {
		doLog("saveStats rename: %v", err)
	}
}

func gatherStats() PageData {
	ephemeralLock.Lock()
	current := len(ephemeralIDMap)
	peak := ephemeralPeak
	total := ephemeralSessionsTotal
	sessions := []SessionInfo{}
	for _, s := range ephemeralIDMap {
		sess := SessionInfo{
			ID:          s.id,
			DestPort:    s.destPort,
			Duration:    time.Since(s.startTime).Round(time.Second).String(),
			BytesIn:     s.bytesIn,
			BytesOut:    s.bytesOut,
			BytesInStr:  humanize.Bytes(uint64(s.bytesIn)),
			BytesOutStr: humanize.Bytes(uint64(s.bytesOut)),
		}
		sessions = append(sessions, sess)
	}
	inTotal := bytesInTotal
	outTotal := bytesOutTotal
	ephemeralLock.Unlock()

	return PageData{
		CurrentUsers:     current,
		PeakUsers:        peak,
		TotalSessions:    total,
		Uptime:           time.Since(startTime).Round(time.Second).String(),
		BatchInterval:    batchingMicroseconds,
		Compression:      compressionLevel,
		Sessions:         sessions,
		BytesInTotal:     inTotal,
		BytesOutTotal:    outTotal,
		BytesInTotalStr:  humanize.Bytes(uint64(inTotal)),
		BytesOutTotalStr: humanize.Bytes(uint64(outTotal)),
	}
}
