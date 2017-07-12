package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestCX(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CX")
}

var _ = BeforeSuite(func() {
})

var _ = AfterSuite(func() {
})
