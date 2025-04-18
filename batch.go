package main

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

func (tun *tunnelCon) write(buf []byte) {
	if tun == nil {
		return
	}

	tun.packetLock.Lock()
	if batchingMicroseconds == 0 {
		tun.packets = buf
		tun.packetsLength = len(buf)
		err := writeBatch(tun)
		if err != nil {
			doLog("tun write: %v", err)
		}
	} else {
		tun.packets = append(tun.packets, buf...)
		tun.packetsLength += len(buf)
	}
	tun.packetLock.Unlock()
}

func (tun *tunnelCon) batchWriter() {
	if batchingMicroseconds == 0 {
		return
	}
	ticker := time.NewTicker(time.Microsecond * time.Duration(batchingMicroseconds))

	if debugLog {
		defer doLog("batchWriter: exit")
	}

	for range ticker.C {
		if tun == nil || tun.con == nil {
			return
		}
		if time.Since(tun.lastUsed) > tunnelLife {
			doLog("Idle, disconnected")
			tun.delete(true)
			return
		}
		tun.packetLock.Lock()
		err := writeBatch(tun)
		tun.packetLock.Unlock()
		if err != nil {
			return
		}
	}

}

func writeBatch(tun *tunnelCon) error {
	if tun.packetsLength == 0 {
		return nil
	}

	var dataToWrite []byte
	if compressionLevel > 0 {
		dataToWrite = compressFrame(tun)
	} else {
		dataToWrite = tun.packets
	}

	var header []byte
	header = binary.AppendUvarint(header, uint64(len(dataToWrite)))
	l, err := tun.con.Write(append(header, dataToWrite...))
	if err != nil {
		return fmt.Errorf("writeBatch: Write error: %v", err)
	}

	if l < len(dataToWrite) {
		return fmt.Errorf("writeBatch: Partial write: wrote %d of %d bytes", l, len(dataToWrite))
	}

	tun.packets = nil
	tun.packetsLength = 0
	return nil
}

func compressFrame(tun *tunnelCon) []byte {
	var buf bytes.Buffer
	gz, _ := gzip.NewWriterLevel(&buf, compressionLevels[compressionLevel])
	if _, err := gz.Write(tun.packets); err != nil {
		doLog("compressFrame: gzip write error: %v", err)
		_ = gz.Close()
		return nil
	}
	if err := gz.Close(); err != nil {
		doLog("compressFrame: gzip close error: %v", err)
		return nil
	}
	return buf.Bytes()
}

func decompressFrame(data []byte) ([]byte, error) {
	buf := bytes.NewReader(data)
	gz, err := gzip.NewReader(buf)
	if err != nil {
		return nil, fmt.Errorf("decompressFrame: gzip reader error: %v", err)
	}
	defer gz.Close()

	var out bytes.Buffer
	if _, err := io.Copy(&out, gz); err != nil {
		return nil, fmt.Errorf("decompressFrame: gzip decompress copy error: %v", err)
	}
	return out.Bytes(), nil
}
