package main

import (
	"flag"
	"os/exec"
	"github.com/cloud66/cli"
	. "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

var startProgramTest *exec.Cmd

var _ = Describe("Register a new server", func() {
	Context("with a local (private) IP adresss", func() {
		var flagSet *flag.FlagSet

		BeforeEach(func() {
			underTest = true
			//capture output and mock API endpoint
			StartCaptureStdout()
			MockApiGetCall("/accounts/139.json", 200, "./mocks/accounts/account.json")
			MockApiGetCall("/accounts.json", 200, "./mocks/accounts/list.json")
			flagSet = flag.NewFlagSet("test", 0)
		})


		It("should use the special header", func() {
			// set the enviroment flag to production
			flagSet.String("org", "test_org", "")
			flagSet.String("server", "192.168.1.10", "")
			flagSet.String("user", "root", "")
			flagSet.String("force-local-ip", "true", "")



			// run context with the CLI
			context := cli.NewContext(nil, flagSet, nil)
			runRegisterServer(context)

			// read stdout
			output := StopCaptureStdout()
			Expect(startProgramTest.Args[17]).To(HavePrefix("'curl --header \"X-Force-Local-IP: true\" -s http://app.cloud66.com/xxx| bash -s'"))
			Expect(output[0]).To(HavePrefix("Register server(s) done."))
		})
	})
})
