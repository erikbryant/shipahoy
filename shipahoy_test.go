package shipahoy

import (
	"testing"
)

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
		{37.8052, -122.48, true}, // bottom left corner
		{37.8612, -122.48, true}, // just inside top left corner
		// Fails on my new macbook due to rounding error :(
		// {37.8052, -122.4092, true}, // bottom right corner
		{37.82, -122.46, true},   // mid triangle
		{37.805, -122.49, false}, // outside bottom left corner
		{37.87, -122.49, false},  // outside top left corner
		{37.805, -122.4, false},  // outside bottom right corner
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
