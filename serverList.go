package main

import (
	"encoding/binary"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
)

func handleForwardedPorts(tun *tunnelCon, portCounts int) error {
	for _, listener := range listeners {
		listener.Close()
	}

	//Build port list data
	forwardedPorts = []int{}
	portsStr := ""
	listeners = []*net.UDPConn{}
	for p := range portCounts {

		//Read port
		port, err := binary.ReadUvarint(tun.frameReader)
		if err != nil {
			return fmt.Errorf("unable to read forwarded port: %v", err)
		}

		//Name length
		nameLen, err := binary.ReadUvarint(tun.frameReader)
		if err != nil {
			return fmt.Errorf("unable to read name length: %v", err)
		}

		var name []byte
		if nameLen > 0 {
			name = make([]byte, nameLen)
			l, err := io.ReadFull(tun.frameReader, name)
			if err != nil {
				return fmt.Errorf("Unable to read frame data: %v", err)
			}
			if l != int(nameLen) {
				return fmt.Errorf("Unable to read all frame data: %v of %v", l, nameLen)
			}
		}

		//Build list
		forwardedPorts = append(forwardedPorts, int(port))
		forwardedPortsNames = append(forwardedPortsNames, string(name))
		if p != 0 {
			portsStr = portsStr + ", "
		}
		if nameLen > 0 {
			portsStr = portsStr + string(name) + " - "
		}
		portsStr = portsStr + strconv.FormatUint(port, 10)

		laddr := &net.UDPAddr{IP: nil, Port: int(port)}
		conn, err := net.ListenUDP("udp", laddr)
		if err != nil {
			return fmt.Errorf("unable to read from laddr: %v", err)
		}

		listeners = append(listeners, conn)
	}

	log.Printf("Forwarded ports: %v", portsStr)
	outputServerList()

	return nil
}

func outputServerList() {
	data := PageData{Servers: []ServerEntry{}}

	for i, port := range forwardedPorts {
		name := forwardedPortsNames[i]
		server := ServerEntry{Name: name, Addr: clientAddr, Port: port}
		data.Servers = append(data.Servers, server)
	}

	tem := privateServerTemplate
	if PublicClientMode != "true" {
		tem = publicServerTemplate
	}

	tmpl, err := template.New("page").Parse(tem)
	if err != nil {
		log.Fatalf("Failed to parse template: %v", err)
	}

	if PublicClientMode != "true" {
		htmlFileName = "index.html"
	}
	f, err := os.Create(htmlFileName)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	err = tmpl.Execute(f, data)
	if err != nil {
		log.Fatalf("Failed to execute template: %v", err)
	}

	log.Printf("UPDATED %v written successfully... (attempting to open)", htmlFileName)

	if PublicClientMode == "true" {
		if err := openInBrowser(htmlFileName); err != nil {
			log.Printf("Failed to open in browser: %v", err)
		} else {
			log.Printf("The server list should now be open!")
		}
	}
}

func openInBrowser(path string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default: // linux, freebsd, etc.
		cmd = exec.Command("xdg-open", path)
	}

	return cmd.Start()
}

func tinyHTTPServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile("index.html")
		if err != nil {
			http.Error(w, "File not found.", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	log.Println("Serving index.html on port 80...")
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
