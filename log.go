package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var (
	logDesc *os.File
	logName string
)

func doLog(format string, args ...any) {

	if logDesc == nil {
		return
	}

	ctime := time.Now()
	_, filename, line, _ := runtime.Caller(1)

	var text string
	if args == nil {
		text = format
	} else {
		text = fmt.Sprintf(format, args...)
	}

	date := fmt.Sprintf("%2v:%2v.%2v", ctime.Hour(), ctime.Minute(), ctime.Second())
	buf := fmt.Sprintf("%v: %15v:%5v: %v\n", date, filepath.Base(filename), line, text)
	_, err := logDesc.WriteString(buf)
	fmt.Print(buf)
	if err != nil {
		fmt.Print("DoLog: WriteString failure")
		logDesc = nil
		return
	}
}

/* Prep everything for the cw log */
func startLog() {

	t := time.Now().UTC()

	/* Create our log file names */
	logName = fmt.Sprintf("log/log-%v-%v-%v.log", t.Day(), t.Month(), t.Year())

	/* Make log directory */
	errr := os.MkdirAll("log", os.ModePerm)
	if errr != nil {
		fmt.Print(errr.Error())
		return
	}

	/* Open log files */
	bdesc, errb := os.OpenFile(logName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	/* Handle file errors */
	if errb != nil {
		fmt.Printf("An error occurred when attempting to create log file. Details: %s", errb)
		return
	}

	if logDesc != nil {
		doLog("Rotating log.")
		logDesc.Close()
	}

	/* Save descriptors, open/closed elsewhere */
	logDesc = bdesc

}

func autoRotateLogs() {
	//Rotate when date changes
	startDay := time.Now().UTC().Day()
	for {
		currentDay := time.Now().UTC().Day()
		if currentDay != startDay {
			startDay = currentDay
			startLog()
		}
		time.Sleep(time.Second)
	}
}
