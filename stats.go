package main

import (
	"encoding/json"
	"os"
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
