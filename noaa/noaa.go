package noaa

import (
	"fmt"
	"github.com/erikbryant/shipahoy/database"
	"github.com/erikbryant/web"
)

// noaaReading reads one datum from a given NOAA station.
func noaaReading(station, product, datum string) (database.NoaaDatum, bool) {
	reading := database.NoaaDatum{
		Station: station,
		Product: product,
		Datum:   datum,
	}
	url := "https://tidesandcurrents.noaa.gov/api/datagetter?date=latest&station=" + reading.Station + "&product=" + reading.Product + "&datum=" + reading.Datum + "&units=english&time_zone=lst_ldt&application=erikbryantology@gmail.com&format=json"

	response, err := web.RequestJSON(url)
	if err != nil {
		fmt.Println("Error getting", reading, err, response)
		return reading, false
	}

	if response["error"] != nil {
		fmt.Println("Error reading", reading, response["error"])
		return reading, false
	}

	data := response["data"].([]interface{})[0].(map[string]interface{})
	reading.Value = data["v"].(string)
	reading.S = data["s"].(string)
	reading.Flags = data["f"].(string)

	return reading, true
}

// Tides looks up instantaneous tide data for a given NOAA station.
func Tides(station string) (database.NoaaDatum, bool) {
	return noaaReading(station, "water_level", "mllw")
}

// AirGap looks up instantaneous air gap (distance from bottom of bridge to water) for a given NOAA station.
func AirGap(station string) (database.NoaaDatum, bool) {
	return noaaReading(station, "air_gap", "mllw")
}
