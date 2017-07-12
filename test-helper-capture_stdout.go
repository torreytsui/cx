//helpers
package main

import (
	"io/ioutil"
	"os"
	"strings"
)

var oldStdoutHelper *os.File = os.Stdout
var writeFileHelper *os.File = nil
var readFileHelper *os.File = nil

func StartCaptureStdout() {
	//capture stdout
	r, w, _ := os.Pipe()
	writeFileHelper = w
	readFileHelper = r
	os.Stdout = w
}

func StopCaptureStdout() []string {
	writeFileHelper.Close()
	out, _ := ioutil.ReadAll(readFileHelper)
	os.Stdout = oldStdoutHelper
	outputLines := strings.Split(string(out[:]), "\n")
	return outputLines
}
