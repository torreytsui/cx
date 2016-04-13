//helpers
package helpers

import (
	"os"
	"strings"
	"io/ioutil"
)

var oldStdout *os.File = os.Stdout
var writeFile *os.File = nil
var readFile *os.File = nil


func StartCaptureStdout() {
	//capture stdout
    r, w, _ := os.Pipe()
    writeFile = w
    readFile = r
    os.Stdout = w
}

 func StopCaptureStdout() []string {
	writeFile.Close()
    out, _ := ioutil.ReadAll(readFile)
    os.Stdout = oldStdout
    outputLines := strings.Split(string(out[:]), "\n")
    return outputLines
 }