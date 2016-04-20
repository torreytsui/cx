package main

import (
	. "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    "github.com/h2non/gock"
)

var _ = Describe("Update", func() {
	Context("a new version", func() {
		It("should retreive the newest version", func() {
			defer gock.Off()
			gock.New("http://downloads.cloud66.com/cx").
				Get("/cx_latest.json").
    			Reply(200).
    			BodyString("{\"latest\":\"0.1.40\"}")

  			response, err := findLatestVersion()
			Expect(string(response.Version)).To(Equal("0.1.40"))
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
