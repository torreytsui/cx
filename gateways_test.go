package main

import (
	"flag"

	"github.com/cloud66/cli"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Gateways command", func() {
	Context("with a customer having an active gateway", func() {

		var flagSet *flag.FlagSet

		BeforeEach(func() {
			//capture output and mock API endpoint
			StartCaptureStdout()
			MockApiGetCall("/accounts.json", 200, "./mocks/accounts/list.json")
			MockApiGetCall("/accounts/139/gateways.json", 200, "./mocks/gateways/list.json")
			MockApiPutCall("/accounts/139/gateways/1.json", 200, "./mocks/gateways/list.json")

			flagSet = flag.NewFlagSet("test", 0)
		})

		It("can list the gateways", func() {
			// set the enviroment flag to production
			flagSet.String("org", "test_org", "")

			// run context with the CLI
			context := cli.NewContext(nil, flagSet, nil)
			runListGateways(context)

			// read stdout
			output := StopCaptureStdout()

			// check the actual output
			Expect(output[1]).To(HavePrefix("aws_bastion  ec2-usr   1.1.1.1  2.2.2.2"))
			Expect(output[1]).To(HaveSuffix("close"))
			//Expect(len(output)-1).To(Equal(3))
		})

		It("can open the gateway", func() {
			// set the enviroment flag to production
			flagSet.String("name", "aws_bastion", "")
			flagSet.String("ttl", "1h", "")
			flagSet.String("key", "./mocks/gateways/test.pem", "")
			flagSet.String("org", "test_org", "")

			// run context with the CLI
			context := cli.NewContext(nil, flagSet, flagSet)
			context.SetAllFlags()

			runOpenGateway(context)

			// read stdout
			output := StopCaptureStdout()

			// check the actual output
			Expect(output[0]).To(Equal("Gateway opened successfully!"))
		})

	})
})
