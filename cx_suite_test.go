package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
	"bytes"
	"time"
	"os"
	"os/exec"
)

var binPath string = "cx"

func FileExists(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func TestCx(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cx Suite")
}

var _ = BeforeSuite(func() {
	command := exec.Command("git", "describe", "--abbrev=0", "--tags")
	command_out, _ := command.Output()
	version := bytes.TrimRight(command_out, "\n")
	current_date := time.Now().Format("2006-01-02")


	err := exec.Command("go", "build", "-ldflags","-X \"main.VERSION=" + string(version) + "\" -X \"main.BUILD_DATE=" + current_date + "\"").Run()
	Expect(err).NotTo(HaveOccurred())
	Expect(FileExists(binPath)).To(BeTrue())
})

var _ = AfterSuite(func() {
	err := exec.Command("rm", binPath).Run()
	Expect(err).NotTo(HaveOccurred())
})

