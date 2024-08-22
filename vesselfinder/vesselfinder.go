package vesselfinder

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/erikbryant/shipahoy/aismmsi"
	"github.com/erikbryant/web"
)

// bytesToInt converts a string (taken as an array of bytes) to an int.
func bytesToInt(buf string) int {
	result := 0
	for i := 0; i < len(buf); i++ {
		result = result<<8 | int(buf[i])
	}
	return result
}

// directLink builds a link to details about a given ship.
func directLink(name, imo, mmsi string) string {
	// Some ship names have '/' in them. For instance, motor yachts
	// sometimes have a 'M/Y' prefix. Rather than URL encode the slash,
	// VesselFinder removes it. Do the same here.
	n := strings.ReplaceAll(name, "/", "")
	n = strings.ReplaceAll(n, " ", "-")
	n = strings.ToUpper(n)

	return ("https://www.vesselfinder.com/vessels/" + n + "-IMO-" + imo + "-MMSI-" + mmsi)
}

// ShipsInRegion returns the ships found in a given lat/lon area via a channel.
func ShipsInRegion(latA, lonA, latB, lonB float64, c chan map[string]interface{}) {
	defer close(c)

	// Convert to minutes and trunc after 4 decimal places
	latAs := strconv.Itoa(int(math.Trunc(latA * 600000)))
	lonAs := strconv.Itoa(int(math.Trunc(lonA * 600000)))
	latBs := strconv.Itoa(int(math.Trunc(latB * 600000)))
	lonBs := strconv.Itoa(int(math.Trunc(lonB * 600000)))

	url := "https://www.vesselfinder.com/api/pub/mp2?bbox=" + lonAs + "%2C" + latAs + "%2C" + lonBs + "%2C" + latBs + "&zoom=12&mmsi=0&show_names=1"

	region, err := web.RequestBody(url, map[string]string{})
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(region) < 4 {
		fmt.Println("Too little data returned: ", region)
		return
	}
	if region[0:2] != "CE" {
		fmt.Println("Unexpected data returned: ", region)
		return
	}

	// Code adapted from the drawShipsOnMapBinary function found in
	// https://static.vesselfinder.net/web/main3.js?4.22b1&v6

	rLen := len(region)

	offset := 0
	if len(region) >= 12 {
		offset = bytesToInt(region[2:3])
	}

	for i := 4 + offset; i < rLen; {
		// Unknown what B means
		B := bytesToInt(region[i : i+2])
		i += 2

		// The MMSI
		mmsi := strconv.Itoa(bytesToInt(region[i : i+4]))
		i += 4

		// Lat/Lon
		j := 600000.0
		lat := float64(bytesToInt(region[i:i+4])) / j
		i += 4
		lon := float64(bytesToInt(region[i:i+4])) / j
		i += 4

		// Unused
		i += 1

		// The ship's name (a counted string)
		nameLen := bytesToInt(region[i : i+1])
		i += 1
		if i+nameLen > rLen {
			break
		}
		// Default to MMSI
		name := mmsi
		if nameLen > 0 {
			name = region[i : i+nameLen]
			i += nameLen
		}

		// Unknown what D means
		D := false
		if D {
			// Unused
			i += 2
			i += 2
			i += 2
			i += 2
			i += 2
		}

		re := (2 & B) != 0
		if re && !D {
			// Unused
			i += 2
		}

		response, ok := getShipDetails(mmsi)
		if !ok {
			continue
		}

		if web.ToString(response["name"]) != name {
			// We have a non-existent MMSI.
			continue
		}

		response["directLink"] = directLink(web.ToString(response["name"]), web.ToString(response["imo"]), mmsi)
		response["lat"] = lat
		response["lon"] = lon
		response["mmsi"] = mmsi

		// Push 'response' to channel.
		c <- response
	}
}

// prettify returns the input formatted as a printable string.
func prettify(i interface{}) string {
	s, err := json.MarshalIndent(i, "", " ")
	if err != nil {
		fmt.Println("Could not Marshal object", i)
	}

	return string(s)
}

// getShipDetails retrieves ship details for a given mmsi from the web.
func getShipDetails(mmsi string) (map[string]interface{}, bool) {
	err := aismmsi.ValidateMmsi(mmsi)
	if err != nil {
		fmt.Println(err)
		return nil, false
	}

	mmsiURL := "https://www.vesselfinder.com/api/pub/click/" + mmsi
	response, err := web.RequestJSON(mmsiURL, map[string]string{})
	if err != nil || response == nil {
		return nil, false
	}

	if response[".ns"] == nil {
		// Sometimes '.ns' is returned as nil. If so, the whole
		// response is garbage.
		return nil, false
	}

	if web.ToInt(response[".ns"]) < 0 {
		// We sometimes get -1 back from VesselFinder. That is not a valid
		// navigational status. On the VesselFinder website they show the
		// ship as 'at anchor', so we will do the same thing.
		response[".ns"] = 1
	}

	if mmsi == "319762000" && web.ToFloat64(response["ss"]) > 100.0 {
		// The M/Y SAINT NICHOLAS sometimes reports bad AIS data.
		// It takes the form of course: 360, speed: 102.3, status: at anchor.
		// Ignore those reports.
		return nil, false
	}

	return response, true
}
