package aismmsi

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
		err := ValidateMmsi(testCase.mmsi)
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
		answer := DecodeMmsi(testCase.mmsi)
		if answer != testCase.expected {
			t.Errorf("ERROR: For %v expected %v, got %v", testCase.mmsi, testCase.expected, answer)
		}
	}
}
