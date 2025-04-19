package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/Masterminds/semver"
)

const (
	baseURL     = "https://m45sci.xyz/relayClient/"
	UpdateJSON  = "https://m45sci.xyz/relayClient/relayClient.json"
	updateDebug = false
)

type downloadInfo struct {
	Link     string `json:"link"`
	Checksum string `json:"sha256"`
}

type Entry struct {
	Version string         `json:"version"`
	Date    int64          `json:"utc-unixnano"`
	Links   []downloadInfo `json:"links"`
}

func OSString() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return "win", nil
	case "darwin":
		return "mac", nil
	case "linux":
		return "linux", nil
	default:
		return "", fmt.Errorf("did not detect a valid host OS")
	}
}

func CheckUpdate() (bool, error) {
	log.Print("Checking for relayClient updates.")
	jsonBytes, fileName, err := httpGet(UpdateJSON)
	if err != nil {
		return false, err
	}

	if len(jsonBytes) == 0 {
		return false, fmt.Errorf("empty response")
	}
	if updateDebug {
		log.Printf("len: %v, name: %v\n", len(jsonBytes), fileName)
	}

	jsonReader := bytes.NewReader(jsonBytes)
	decoder := json.NewDecoder(jsonReader)
	entries := []Entry{}
	if err := decoder.Decode(&entries); err != nil && err != io.EOF {
		log.Printf("error decoding json: %v\n", err)
		os.Exit(1)
	}

	remoteNewest, err := NewestEntry(entries)
	if err != nil {
		return false, fmt.Errorf("NewestEntry: %v", err)
	}

	ourVersion, err := semver.NewVersion(version)
	remoteVersion, err := semver.NewVersion(remoteNewest.Version)
	if !ourVersion.LessThan(remoteVersion) {
		log.Print("clientRelay is update to date.")
		return false, nil
	}

	log.Printf("Found new version: %v\n", remoteNewest.Version)

	goos, err := OSString()
	if err != nil {
		return false, fmt.Errorf("OSString: %v", err)
	}
	var updateLink *downloadInfo
	for _, link := range remoteNewest.Links {
		if strings.Contains(
			strings.ToLower(link.Link),
			strings.ToLower("-"+goos+"-")) {
			updateLink = &link
			break
		}
	}
	if updateLink == nil {
		return false, fmt.Errorf("No valid download link found")
	} else {
		log.Printf("Downloading: %v\n", baseURL+updateLink.Link)
		data, fileName, err := httpGet(baseURL + updateLink.Link)
		if err != nil {
			return false, fmt.Errorf("httpGet: %v", err)
		}
		if updateDebug {
			log.Printf("Filename: %v, Size: %vb\n", fileName, len(data))
		}
		checksum, err := computeChecksum(data)
		if checksum != updateLink.Checksum {
			return false, fmt.Errorf("file: %v - checksum is invalid.", fileName)
		} else {
			log.Print("Download complete, updating.")
			if err := UnzipToExeDir(data); err != nil {
				log.Fatalf("Extraction failed: %v\n", err)
			}
			log.Printf("Update complete, restarting.")
			relaunch()
		}

		return true, nil
	}
}

// relaunch replaces the current process with a new instance of the same binary.
// It never returns on success; on failure it returns an error.
func relaunch() error {
	// 1) Find the path to the currently running executable
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("cannot find executable path: %w", err)
	}

	// 2) Grab the original args (including os.Args[0]) so the new process is identical
	args := os.Args

	// 3) Inherit the current environment
	env := os.Environ()

	// 4) Exec – on success this never returns, as the Go runtime is replaced
	return syscall.Exec(exePath, args, env)
}

func computeChecksum(data []byte) (string, error) {
	dataReader := bytes.NewReader(data)
	h := sha256.New()
	if _, err := io.Copy(h, dataReader); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func httpGet(input string) ([]byte, string, error) {
	// Set timeout
	hClient := http.Client{
		Timeout: time.Second * 30,
	}

	//Use proxy if provided
	var URL string = input

	//HTTP GET
	req, err := http.NewRequest(http.MethodGet, URL, nil)
	if err != nil {
		return nil, "", errors.New("get failed: " + err.Error())
	}

	//Get response
	res, err := hClient.Do(req)
	if err != nil {
		return nil, "", errors.New("failed to get response: " + err.Error())
	}

	//Check status code
	if res.StatusCode != 200 {
		return nil, "", fmt.Errorf("http status error: %v", res.StatusCode)
	}

	//Close once complete, if valid
	if res.Body != nil {
		defer res.Body.Close()
	}

	//Read all
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, "", errors.New("unable to read response body: " + err.Error())
	}

	//Check data length
	if res.ContentLength > 0 {
		if len(body) != int(res.ContentLength) {
			return nil, "", errors.New("data ended early")
		}
	} else if res.ContentLength != -1 {
		return nil, "", errors.New("content length did not match")
	}

	realurl := res.Request.URL.String()
	parts := strings.Split(realurl, "/")
	query := parts[len(parts)-1]
	parts = strings.Split(query, "?")
	return body, parts[0], nil
}

func NewestEntry(entries []Entry) (*Entry, error) {
	// pair up each Entry with its parsed semver.Version
	type pair struct {
		e   *Entry
		ver *semver.Version
	}
	var pairs []pair
	for i := range entries {
		e := &entries[i]
		v, err := semver.NewVersion(e.Version)
		if err != nil {
			// skip unparsable versions
			continue
		}
		pairs = append(pairs, pair{e: e, ver: v})
	}

	if len(pairs) == 0 {
		return nil, fmt.Errorf("semutil: no valid versions found in entries")
	}

	// sort ascending by version (lowest → highest)
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].ver.LessThan(pairs[j].ver)
	})

	// the last element has the highest version
	return pairs[len(pairs)-1].e, nil
}

// UnzipToExeDir unpacks the zip from data into the directory of the running binary.
func UnzipToExeDir(data []byte) error {
	// figure out where the binary lives
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("os.Executable: %w", err)
	}
	exeDir := filepath.Dir(exePath)
	return UnzipToDir(data, exeDir)
}

// UnzipToDir unpacks the zip archive in data into destDir, preserving folders and file modes.
func UnzipToDir(data []byte, destDir string) error {
	// open zip reader
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return fmt.Errorf("zip.NewReader: %w", err)
	}

	for _, f := range r.File {
		targetPath := filepath.Join(destDir, f.Name)

		if f.FileInfo().IsDir() {
			// create sub‑directory
			if err := os.MkdirAll(targetPath, os.ModePerm); err != nil {
				return fmt.Errorf("mkdir %q: %w", targetPath, err)
			}
			continue
		}

		// make sure parent dir exists
		if err := os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
			return fmt.Errorf("mkdirall %q: %w", filepath.Dir(targetPath), err)
		}

		// open file inside zip
		inFile, err := f.Open()
		if err != nil {
			return fmt.Errorf("open %q in zip: %w", f.Name, err)
		}
		defer inFile.Close()

		// create destination file
		outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("open file %q: %w", targetPath, err)
		}
		defer outFile.Close()

		// copy contents
		if _, err := io.Copy(outFile, inFile); err != nil {
			return fmt.Errorf("copy to %q: %w", targetPath, err)
		}
	}

	return nil
}
