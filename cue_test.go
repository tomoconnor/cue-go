package cue

import (
	"os"
	"testing"
)

func TestPackage(t *testing.T) {
	filename := "test.cue"

	file, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Failed to open file. %s", err.Error())
	}

	_, err = Parse(file)
	if err != nil {
		t.Fatalf("Failed to parse file. %s", err.Error())
	}
}
