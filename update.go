package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver"
)

const (
	baseURL    = "https://m45sci.xyz/relayClient/"
	UpdateJSON = "https://m45sci.xyz/relayClient/relayClient.json"
)

type UpdateEntry struct {
	Version   string            `json:"version"`
	Date      time.Time         `json:"date"`
	Links     []string          `json:"links"`
	Checksums map[string]string `json:"checksums,omitempty"`
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
		return "", fmt.Errorf("not a supported OS")
	}
}

func CheckUpdate() (bool, error) {
	jsonBytes, fileName, err := httpGet(UpdateJSON)
	if err != nil {
		return false, err
	}

	if len(jsonBytes) == 0 {
		return false, fmt.Errorf("empty response")
	}
	fmt.Printf("len: %v, name: %v\n", len(jsonBytes), fileName)

	jsonReader := bytes.NewReader(jsonBytes)
	decoder := json.NewDecoder(jsonReader)
	entries := []UpdateEntry{}
	if err := decoder.Decode(&entries); err != nil && err != io.EOF {
		fmt.Fprintf(os.Stderr, "Error decoding JSON: %v\n", err)
		os.Exit(1)
	}

	newest, err := NewestEntry(entries)
	if err != nil {
		return false, fmt.Errorf("NewestEntry: %v", err)
	}

	fmt.Printf("Newest Version: %v\n", newest.Version)

	goos, err := OSString()
	if err != nil {
		return false, fmt.Errorf("OSString: %v", err)
	}
	dlLink := ""
	for _, link := range newest.Links {
		if strings.Contains(
			strings.ToLower(link),
			strings.ToLower("-"+goos+"-")) {
			dlLink = link
			break
		}
	}
	if dlLink == "" {
		return false, fmt.Errorf("No valid download link found")
	} else {
		fmt.Printf("Downloading: %v\n", baseURL+dlLink)
		data, fileName, err := httpGet(baseURL + dlLink)
		if err != nil {
			return false, fmt.Errorf("httpGet: %v", err)
		}
		fmt.Printf("Filename: %v, Size: %vb", fileName, len(data))
		return true, nil
	}
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

func NewestEntry(entries []UpdateEntry) (*UpdateEntry, error) {
	// pair up each Entry with its parsed semver.Version
	type pair struct {
		e   *UpdateEntry
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
