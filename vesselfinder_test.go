package shipahoy

import (
	"testing"
)

func TestGetUInt16(t *testing.T) {
	testCases := []struct {
		buf      string
		expected uint16
		error    bool
	}{
		{"\x00\x00", 0x0000, false},
		{"\x00\x01", 0x0001, false},
		{"\x01\x00", 0x0100, false},
		{"\x7F\x7F", 0x7F7F, false},
		{"\x80\x80", 0x8080, false},
		{"\x80", 0x0000, true},
		{"\x80\x80\x80", 0x0000, true},
	}

	for _, testCase := range testCases {
		answer, err := getUInt16(testCase.buf)
		if answer != testCase.expected {
			t.Errorf("ERROR: For %v expected %v, got %v", testCase.buf, testCase.expected, answer)
		}
		if (err != nil) != testCase.error {
			t.Errorf("ERROR: For %v expected error to be %v, got %v", testCase.buf, testCase.error, err)
		}
	}
}

func TestGetInt32(t *testing.T) {
	testCases := []struct {
		buf      string
		expected int32
		error    bool
	}{
		{"\x00\x00\x00\x00", 0x00000000, false},
		{"\x00\x00\x00\x01", 0x00000001, false},
		{"\x00\x00\x01\x00", 0x00000100, false},
		{"\x00\x01\x00\x00", 0x00010000, false},
		{"\x01\x00\x00\x00", 0x01000000, false},
		{"\x10\x00\x00\x00", 0x10000000, false},
		{"\x00\x10\x00\x00", 0x00100000, false},
		{"\x00\x00\x10\x00", 0x00001000, false},
		{"\x00\x00\x00\x10", 0x00000010, false},
		{"\x80\x80\x80", 0x0000, true},
		{"\x80\x80\x80\x80\x80", 0x0000, true},
	}

	for _, testCase := range testCases {
		answer, err := getInt32(testCase.buf)
		if answer != testCase.expected {
			t.Errorf("ERROR: For %v expected %v, got %v", testCase.buf, testCase.expected, answer)
		}
		if (err != nil) != testCase.error {
			t.Errorf("ERROR: For %v expected error to be %v, got %v", testCase.buf, testCase.error, err)
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
