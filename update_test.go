package main

import (
	"testing"
)

func TestFindLatestVersion(t *testing.T) {
	_, err := findLatestVersion()
	if err != nil {
		t.Errorf("Error %v\n", err)
	}
}
