package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"bitbucket.org/kardianos/osext"
	"github.com/inconshreveable/go-update"
)

var cmdUpdate = &Command{
	Run:      runUpdate,
	Usage:    "update [-v <version>]",
	Category: "cx",
	Long: `This command runs automatically. You should not need to run it manually

  -v forces a specific version to be downloaded.
  `,
}

type CxDownload struct {
	Version  string `json:"version"`
	Platform string `json:"platform"`
	Arch     string `json:"architecture"`
	SHA      string `json:"sha"`
	File     string `json:"file"`
}

type CxLatest struct {
	Version string `json:"latest"`
}

var (
	flagForcedVersion string
	currentPlatform   string
	currentArch       string
)

var ErrHashMismatch = errors.New("mismatch SHA")
var ErrNoUpdateAvailable = errors.New("no update available")

const (
	DOWNLOAD_URL = "http://downloads.cloud66.com/cx/"
)

func init() {
	debugMode = os.Getenv("CXDEBUG") != ""

	if os.Getenv("CX_PLATFORM") == "" {
		currentPlatform = runtime.GOOS
	} else {
		currentPlatform = os.Getenv("CX_PLATFORM")
	}

	if os.Getenv("CX_ARCH") == "" {
		currentArch = runtime.GOARCH
	} else {
		currentArch = os.Getenv("CX_ARCH")
	}
}

func runUpdate(cmd *Command, args []string) {
	if len(args) > 0 {
		if args[0] == "-v" {
			flagForcedVersion = args[1]
		} else {
			cmd.printUsage()
		}
	}

	if debugMode {
		if flagForcedVersion == "" {
			fmt.Println("No forced version")
		} else {
			fmt.Printf("Forced version is %s\n", flagForcedVersion)
		}
	}
	updateIt, err := needUpdate()
	if err != nil {
		if debugMode {
			fmt.Printf("Cannot verify need for update %v\n", err)
		}
		return
	}
	if !updateIt {
		if debugMode {
			fmt.Println("No need for update")
		}
		return
	}

	// houston we have an update. which one do we need?
	download, err := getVersionManifest(flagForcedVersion)
	if err != nil {
		if debugMode {
			fmt.Printf("Error fetching manifest %v\n", err)
		}
	}
	if download == nil {
		if debugMode {
			fmt.Println("Found no matching download for the current OS and ARCH")
		}
		return
	}

	err = download.update()
	if err != nil {
		if debugMode {
			fmt.Printf("Failed to update: %v\n", err)
		}
		return
	}
}

func needUpdate() (bool, error) {
	// get the latest version from remote
	if debugMode {
		fmt.Println("Checking for latest version")
	}
	latest, err := findLatestVersion()
	if err != nil {
		return false, err
	}

	if debugMode {
		fmt.Printf("Found %s as the latest version\n", latest.Version)
	}

	if flagForcedVersion != "" {
		if debugMode {
			fmt.Printf("Forcing update to %s\n", flagForcedVersion)
		}
		return true, nil
	} else {
		flagForcedVersion = latest.Version
	}

	if VERSION == latest.Version {
		return false, nil
	}

	return true, nil
}

func getVersionManifest(version string) (*CxDownload, error) {
	resp, err := http.Get(DOWNLOAD_URL + "cx_" + version + ".json")
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error fetching version manifest: %d", resp.StatusCode)
	}
	var manifest []CxDownload
	if err = json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return nil, err
	}

	// find our OS and ARCH
	for _, download := range manifest {
		if download.Platform == currentPlatform && download.Arch == currentArch {
			return &download, nil
		}
	}

	return nil, nil
}

func backgroundRun() {
	b, err := needUpdate()
	if err != nil {
		return
	}
	if b {
		if err := update.SanityCheck(); err != nil {
			if debugMode {
				fmt.Println("Will not be able to replace the executable")
			}
			// fail
			return
		}
		self, err := osext.Executable()
		if err != nil {
			// fail update, couldn't figure out path to self
			return
		}
		l := exec.Command("logger", "-tcx")
		c := exec.Command(self, "update")
		if w, err := l.StdinPipe(); err == nil && l.Start() == nil {
			c.Stdout = w
			c.Stderr = w
		}
		c.Start()
	}
}

func (download *CxDownload) update() error {
	bin, err := download.fetchAndVerify()
	if err != nil {
		return err
	}

	err, errRecover := update.FromStream(bytes.NewBuffer(bin))
	if errRecover != nil {
		return fmt.Errorf("update and recovery errors: %q %q\n", err, errRecover)
	}
	if err != nil {
		return err
	}
	fmt.Printf("Updated v%s -> v%s.\n", VERSION, download.Version)
	return nil
}

func (download *CxDownload) fetchAndVerify() ([]byte, error) {
	bin, err := download.fetchBin()
	if err != nil {
		return nil, err
	}
	return bin, nil
}

func (download *CxDownload) fetchBin() ([]byte, error) {
	r, err := fetch(DOWNLOAD_URL + download.File)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	buf, err := download.decompress(r)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (download *CxDownload) decompress(r io.ReadCloser) ([]byte, error) {
	// for darwin and windows the files are zipped
	if download.Platform == "windows" || download.Platform == "darwin" {
		if debugMode {
			fmt.Printf("Decompressing for %s\n", download.Platform)
		}
		// write it to disk and unzip from there
		dest, err := ioutil.TempFile("", "cx")
		defer os.Remove(dest.Name())
		if err != nil {
			return nil, err
		}

		if debugMode {
			fmt.Printf("Using temp file %s\n", dest.Name())
		}
		writer, err := os.Create(dest.Name())
		if err != nil {
			return nil, err
		}
		defer writer.Close()

		io.Copy(writer, r)
		// now unzip it
		zipper, err := zip.OpenReader(dest.Name())
		if err != nil {
			return nil, err
		}
		defer r.Close()

		for _, f := range zipper.File {
			if debugMode {
				fmt.Printf("Zipped file %s\n", f.Name)
			}
			var targetFile string
			if download.Platform == "windows" {
				targetFile = "cx.exe"
			} else {
				targetFile = "cx_" + flagForcedVersion + "_" + currentPlatform + "_" + currentArch + "/cx"
			}

			if f.Name == targetFile {
				rc, err := f.Open()
				if err != nil {
					return nil, err
				}
				defer rc.Close()

				buf := new(bytes.Buffer)
				if _, err = io.Copy(buf, rc); err != nil {
					return nil, err
				}

				// we are done
				return buf.Bytes(), nil
			}
		}
	}

	// for linux they are tarred and gzipped
	if download.Platform == "linux" {
		buf := new(bytes.Buffer)

		gz, err := gzip.NewReader(r)
		if err != nil {
			return nil, err
		}
		if _, err = io.Copy(buf, gz); err != nil {
			return nil, err
		}

		untar := new(bytes.Buffer)
		// now untar
		tr := tar.NewReader(buf)

		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}
			if debugMode {
				fmt.Printf("Gziped file %s\n", hdr.Name)
			}
			if hdr.Name == "cx_"+flagForcedVersion+"_linux_"+currentArch+"/cx" {
				// this is the executable
				if _, err := io.Copy(untar, tr); err != nil {
					return nil, err
				}
			}
		}

		return untar.Bytes(), nil
	}
	panic("unreached")
}

func fetch(url string) (io.ReadCloser, error) {
	if debugMode {
		fmt.Printf("Downloading %s\n", url)
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case 200:
		return resp.Body, nil
	case 401, 403, 404:
		return nil, ErrNoUpdateAvailable
	default:
		return nil, fmt.Errorf("bad http status from %s: %v", url, resp.Status)
	}
}

func findLatestVersion() (*CxLatest, error) {
	path := DOWNLOAD_URL + "cx_latest.json"
	if debugMode {
		fmt.Printf("Dowloading cx manifest from %s\n", path)
	}
	resp, err := http.Get(path)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf("error fetching latest version manifest: %d", resp.StatusCode)
	}
	var latest CxLatest
	if err = json.NewDecoder(resp.Body).Decode(&latest); err != nil {
		return nil, err
	}

	return &latest, nil
}
