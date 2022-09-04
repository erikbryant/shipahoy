package vesselfinder

import (
	"testing"
)

func TestBytesToInt(t *testing.T) {
	testCases := []struct {
		buf      string
		expected int
	}{
		{"\x00", 0x00},
		{"\x01", 0x01},
		{"\x00\x00", 0x0000},
		{"\x00\x01", 0x0001},
		{"\x01\x00", 0x0100},
		{"\x7F\x7F", 0x7F7F},
		{"\x80\x80", 0x8080},
		{"\x00\x00\x00\x00", 0x00000000},
		{"\x00\x00\x00\x01", 0x00000001},
		{"\x00\x00\x01\x00", 0x00000100},
		{"\x00\x01\x00\x00", 0x00010000},
		{"\x01\x00\x00\x00", 0x01000000},
		{"\x10\x00\x00\x00", 0x10000000},
		{"\x00\x10\x00\x00", 0x00100000},
		{"\x00\x00\x10\x00", 0x00001000},
		{"\x00\x00\x00\x10", 0x00000010},
	}

	for _, testCase := range testCases {
		answer := bytesToInt(testCase.buf)
		if answer != testCase.expected {
			t.Errorf("ERROR: For %v expected %v, got %v", testCase.buf, testCase.expected, answer)
		}
	}
}

func TestDirectLink(t *testing.T) {
	testCases := []struct {
		name     string
		imo      string
		mmsi     string
		expected string
	}{
		{"CG Robert Ward", "0", "338926430", "https://www.vesselfinder.com/vessels/CG-ROBERT-WARD-IMO-0-MMSI-338926430"},
		{"M/Y Saint Nicolas", "1008918", "319762000", "https://www.vesselfinder.com/vessels/MY-SAINT-NICOLAS-IMO-1008918-MMSI-319762000"},
	}

	for _, testCase := range testCases {
		answer := directLink(testCase.name, testCase.imo, testCase.mmsi)
		if answer != testCase.expected {
			t.Errorf("ERROR: For %v, %v, %v expected %v, got %v", testCase.name, testCase.imo, testCase.mmsi, testCase.expected, answer)
		}
	}
}
