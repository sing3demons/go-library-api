package logger

import (
	"testing"
	"time"
)

func TestGenerateXTid(t *testing.T) {
	nodeName := "testNode"
	xTid := GenerateXTid(nodeName)

	// Check if the length of the generated xTid is 22
	if len(xTid) != 22 {
		t.Errorf("Expected xTid length to be 22, but got %d", len(xTid))
	}

	// Check if the node name part is correct
	expectedNodeName := nodeName[:5]
	if xTid[:5] != expectedNodeName {
		t.Errorf("Expected node name part to be %s, but got %s", expectedNodeName, xTid[:5])
	}

	// Check if the date part is correct
	now := time.Now()
	expectedDate := now.Format("060102")
	if xTid[6:12] != expectedDate {
		t.Errorf("Expected date part to be %s, but got %s", expectedDate, xTid[6:12])
	}

	// Check if the random string part is of correct length
	randomStringPart := xTid[12:]
	if len(randomStringPart) != 10 {
		t.Errorf("Expected random string part length to be 10, but got %d", len(randomStringPart))
	}
}
