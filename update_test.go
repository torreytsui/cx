package main

import (
	. "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

var _ = Describe("Update", func() {
	Context("a new version", func() {
		It("should retreive the newest version", func() {
			_, err := findLatestVersion()
			//Expect(string(response.Version)).To(Equal("0.1.39"))
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
