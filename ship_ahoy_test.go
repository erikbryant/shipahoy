package main

import (
	"testing"
)

func TestValidateMmsi(t *testing.T) {
	testCases := []struct {
		mmsi  string
		valid bool
	}{
		{"123456789", true},
		{"12345678", false},
		{"1234567890", false},
		{"a23456789", false},
	}

	for _, testCase := range testCases {
		err := validateMmsi(testCase.mmsi)
		isValid := err == nil
		if isValid != testCase.valid {
			t.Errorf("ERROR: For %v expected %v, got %v", testCase.mmsi, testCase.valid, isValid)
		}
	}
}

func TestDecodeMmsi(t *testing.T) {
	testCases := []struct {
		mmsi     string
		expected string
	}{
		// 0
		{"041211111", "Ship group Asia [Immarsat A or other]"},
		{"004121111", "Coast radio station Asia [Immarsat A or other]"},

		// 1
		{"144411111", "Unknown type 1 Asia [Immarsat A or other]"},
		{"111444111", "SAR aircraft Asia [Immarsat A or other]"},

		// 2-7
		{"200111111", "Europe [Immarsat A or other]"},
		{"300111111", "North/Central America [Immarsat A or other]"},
		{"400111111", "Asia [Immarsat A or other]"},
		{"500111111", "Oceania [Immarsat A or other]"},
		{"600111111", "Africa [Immarsat A or other]"},
		{"700111111", "South America [Immarsat A or other]"},

		// 8
		{"855511111", "Handheld VHF Oceania [Immarsat A or other]"},

		// 9
		{"970777111", "Misc SAR transponder South America [Immarsat A or other]"},
		{"972777111", "Misc man overboard device South America [Immarsat A or other]"},
		{"974777111", "Misc EPIRB with AIS transmitter South America [Immarsat A or other]"},
		{"987771111", "Misc craft associated with parent ship South America [Immarsat A or other]"},
		{"997771111", "Misc aid to navigation South America [Immarsat A or other]"},

		// Immarsat
		{"666111111", "Africa [Immarsat A or other]"},
		{"666111110", "Africa [Immarsat C]"},
		{"666111000", "Africa [Immarsat B/C/M]"},
	}

	for _, testCase := range testCases {
		answer := decodeMmsi(testCase.mmsi)
		if answer != testCase.expected {
			t.Errorf("ERROR: For %v expected %v, got %v", testCase.mmsi, testCase.expected, answer)
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

// TODO:
// Because this hardcodes the same values that are hardcoded into
// visibleFromApt() it makes it feel like a change detector test.
// What about factoring the hardcoded rectangle boundaries into
// a global area that this function can pull from?
func TestVisibleFromApt(t *testing.T) {
	testCases := []struct {
		lat      float64
		lon      float64
		expected bool
	}{
		{1.1, 2.2, false},
		{37.8052, -122.48, true},   // bottom left corner
		{37.8613, -122.48, true},   // top left corner
		{37.8052, -122.4092, true}, // bottom right corner
		{37.82, -122.46, true},     // mid triangle
		{37.805, -122.49, false},   // outside bottom left corner
		{37.87, -122.49, false},    // outside top left corner
		{37.805, -122.4, false},    // outside bottom right corner
	}

	for _, testCase := range testCases {
		answer := visibleFromApt(testCase.lat, testCase.lon)
		if answer != testCase.expected {
			t.Errorf("ERROR: For %v, %v expected %v, got %v", testCase.lat, testCase.lon, testCase.expected, answer)
		}
	}
}

func TestBox(t *testing.T) {
	testCases := []struct {
		lat          float64
		lon          float64
		nmiles       float64
		expectedLatA float64
		expectedLonA float64
		expectedLatB float64
		expectedLonB float64
	}{
		{1, 2, 0, 1, 2, 1, 2},
		{100, 100, 60, 99, 99, 101, 101},
		{-100, -100, 60, -101, -101, -99, -99},
	}

	for _, testCase := range testCases {
		latA, lonA, latB, lonB := box(testCase.lat, testCase.lon, testCase.nmiles)
		if latA != testCase.expectedLatA || lonA != testCase.expectedLonA || latB != testCase.expectedLatB || lonB != testCase.expectedLonB {
			t.Errorf("ERROR: For %v, %v, %v expected %v, %v, %v, %v, got %v, %v, %v, %v", testCase.lat, testCase.lon, testCase.nmiles, testCase.expectedLatA, testCase.expectedLonA, testCase.expectedLatB, testCase.expectedLonB, latA, lonA, latB, lonB)
		}
	}
}
