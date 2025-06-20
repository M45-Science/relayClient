package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"runtime"
	"time"

	"github.com/klauspost/compress/zstd"
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
	interval := time.Second
	if batchingMicroseconds > 0 {
		interval = time.Microsecond * time.Duration(batchingMicroseconds)
	}
	ticker := time.NewTicker(interval)

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
	zw, _ := zstd.NewWriter(&buf, zstd.WithEncoderLevel(zstd.EncoderLevel(compressionLevel)), zstd.WithEncoderConcurrency(runtime.NumCPU()))
	if _, err := zw.Write(tun.packets); err != nil {
		doLog("compressFrame: zstd write error: %v", err)
		_ = zw.Close()
		return nil
	}
	if err := zw.Close(); err != nil {
		doLog("compressFrame: zstd close error: %v", err)
		return nil
	}
	return buf.Bytes()
}

func decompressFrame(data []byte) ([]byte, error) {
	buf := bytes.NewReader(data)
	zr, err := zstd.NewReader(buf, zstd.WithDecoderConcurrency(runtime.NumCPU()))
	if err != nil {
		return nil, fmt.Errorf("decompressFrame: zstd reader error: %v", err)
	}
	defer zr.Close()

	var out bytes.Buffer
	if _, err := io.Copy(&out, zr); err != nil {
		return nil, fmt.Errorf("decompressFrame: zstd decompress copy error: %v", err)
	}
	return out.Bytes(), nil
}
