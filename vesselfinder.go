package ship_ahoy

import (
	"fmt"
	"github.com/erikbryant/ship_ahoy/database"
	"github.com/erikbryant/web"
	"math"
	"strconv"
	"strings"
)

// getUInt16 converts an array of two bytes to a unit16.
func getUInt16(buf string) (uint16, error) {
	if len(buf) != 2 {
		return 0, fmt.Errorf("String length must be exactly 2. Found: %v", len(buf))
	}
	return uint16(buf[0])<<8 | uint16(buf[1]), nil
}

// getInt32 converts an array of four bytes to an int32.
func getInt32(buf string) (int32, error) {
	if len(buf) != 4 {
		return 0, fmt.Errorf("String length must be exactly 4. Found: %v", len(buf))
	}
	return int32(buf[0])<<24 | int32(buf[1])<<16 | int32(buf[2])<<8 | int32(buf[3]), nil
}

// shipsInRegion returns the ships found in a given lat/lon area via a channel.
func shipsInRegion(latA, lonA, latB, lonB float64, c chan database.Ship) {
	defer close(c)

	// Convert to minutes and trunc after 4 decimal places
	latAs := strconv.Itoa(int(math.Trunc(latA * 600000)))
	lonAs := strconv.Itoa(int(math.Trunc(lonA * 600000)))
	latBs := strconv.Itoa(int(math.Trunc(latB * 600000)))
	lonBs := strconv.Itoa(int(math.Trunc(lonB * 600000)))

	url := "https://www.vesselfinder.com/api/pub/vesselsonmap?bbox=" + lonAs + "%2C" + latAs + "%2C" + lonBs + "%2C" + latBs + "&zoom=12&mmsi=0&show_names=1"

	region, err := web.Request(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(region) < 4 || region[0:4] != "CECP" {
		fmt.Println("Unexpected data returned from web.Request(): ", region)
		return
	}

	// Skip over the "CECP" magic bytes
	region = region[4:]

	// Parse each ship from the list. The list is a binary structure containing:
	//   ??
	//   mmsi
	//   lat
	//   lon
	//   len(name)
	//   name

	for i := 0; i < len(region); {
		// I do not know what these values represent. The VesselFinder
		// website unpacks them and names them thusly, but gives no
		// hint as to their meaning.
		//
		//  111111
		//  54321098 76543210
		//  --------+--------
		// |OOGGGGGG|zzzz    |
		//  --------+--------
		// V := getUInt16(region[i:i+2])
		// i += 2
		// z := (V & 0xF0) >> 4
		// G := (V & 0x3F00) >> 8
		// O := 1
		// if V & 0xC000 == 0xC000 {
		// 	O = 2
		// }
		// if V & 0xC000 == 0x8000 {
		// 	O = 0
		// }
		// fmt.Println("z =", z)
		// fmt.Println("G =", G)
		// fmt.Println("O =", O)
		//
		// Until we can figure out the contents of these first two
		// bytes, skip over them.
		i += 2

		val, err := getInt32(region[i : i+4])
		if err != nil {
			fmt.Println("Error unpacking MMSI:", err)
			break
		}
		mmsi := fmt.Sprintf("%09d", val)
		err = validateMmsi(mmsi)
		if err != nil {
			fmt.Println(err)
			fmt.Printf("Raw data: 0x%x 0x%x 0x%x 0x%x\n", region[i], region[i+1], region[i+2], region[i+3])
			break
		}
		i += 4

		val, err = getInt32(region[i : i+4])
		if err != nil {
			fmt.Println("Error unpacking lat:", err)
			break
		}
		lat := float64(val) / 600000.0
		i += 4

		val, err = getInt32(region[i : i+4])
		if err != nil {
			fmt.Println("Error unpacking lon:", err)
			break
		}
		lon := float64(val) / 600000.0
		i += 4

		nameLen := int(region[i])
		i++

		if i+nameLen > len(region) {
			fmt.Println("Ran off of end of data:", mmsi, nameLen, region[i:])
			break
		}

		name := region[i : i+nameLen]
		i += nameLen

		details, ok := getShipDetails(mmsi, name, lat, lon)
		if !ok {
			continue
		}

		// Push 'details' to channel.
		c <- details
	}
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

// getShipDetails retrieves ship details from the database, if they exist, or from the web if they do not.
func getShipDetails(mmsi string, name string, lat, lon float64) (database.Ship, bool) {
	details, seen := database.LookupShip(mmsi)

	err := validateMmsi(mmsi)
	if err != nil {
		fmt.Println(err)
		return details, false
	}

	mmsiURL := "https://www.vesselfinder.com/api/pub/click/" + mmsi
	response, err := web.RequestJSON(mmsiURL)
	if err != nil || response == nil {
		return details, false
	}
	if web.ToString(response["name"]) != name {
		// We have a non-existant MMSI. Abort.
		return details, false
	}

	// Example response
	//
	// https://www.vesselfinder.com/api/pub/click/367003250
	// {
	// 	".ns":0,                 //
	// 	"a2":"us",               // country of register (abbrv)
	// 	"al":19,                 // length
	// 	"aw":8,                  // width
	// 	"country":"USA",         // country of register
	// 	"cu":246.7,              // course
	// 	"dest":"FALSE RIVER",    // destination
	// 	"draught":33,            // draught
	// 	"dw":0,                  // deadweight(?)
	// 	"etaTS":1588620600,      // ETA timestamp
	// 	"gt":0,                  // gross tonnage
	// 	"imo":0,                 // imo number
	// 	"lc.":0,                 //
	// 	"m9":0,                  //
	// 	"name":"SARAH REED",     // name
	// 	"pic":"0-367003250-cf317c76a96fd9b9f5ae4679c64bd065",
	// 	"r":2,                   //
	// 	"sc.":0,                 //
	// 	"sl":false,              //
	// 	"ss":0.1,                // speed (knots)
	// 	"ts":1587883051          // timestamp (of position received?)
	// 	"type":"Towing vessel",  // AIS type
	// 	"y":0,                   // year built
	// }

	details.MMSI = mmsi
	details.Lat = lat
	details.Lon = lon
	details.IMO = web.ToString(response["imo"])
	details.Name = web.ToString(response["name"])
	details.Type = web.ToString(response["type"])
	details.GT = web.ToInt(response["gt"])
	details.DW = web.ToInt(response["dw"])
	details.DirectLink = directLink(details.Name, details.IMO, mmsi)
	details.Draught = web.ToFloat64(response["draught"]) / 10
	details.Year = web.ToInt(response["y"])
	details.Length = web.ToInt(response["al"])
	details.Beam = web.ToInt(response["aw"])
	details.ShipCourse = web.ToFloat64(response["cu"])
	details.Speed = web.ToFloat64(response["ss"])

	if !seen {
		// fmt.Printf("Found: %s %-25s %s\n", details.MMSI, details.Name, decodeMmsi(details.MMSI))
	}

	database.SaveShip(details)

	return details, true
}
