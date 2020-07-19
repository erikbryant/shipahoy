package vesselfinder

import (
	"encoding/json"
	"fmt"
	"github.com/erikbryant/shipahoy/aismmsi"
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
		// I do not know what the first two bytes represent. The VesselFinder
		// website unpacks them and names them thusly, but gives no hint as
		// to their meaning.
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
		// fmt.Println("z =", z, "G =", G, "O =", O)
		//
		// Until we can decode the contents of these first two bytes,
		// skip over them.
		i += 2

		val, err := getInt32(region[i : i+4])
		if err != nil {
			fmt.Println("Error unpacking MMSI:", err)
			break
		}
		mmsi := fmt.Sprintf("%09d", val)
		err = aismmsi.ValidateMmsi(mmsi)
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

// prettify formats and prints the input.
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
	response, err := web.RequestJSON(mmsiURL)
	if err != nil || response == nil {
		return nil, false
	}

	// Example response
	//
	// https://www.vesselfinder.com/api/pub/click/367003250
	// {
	// 	".ns": 0,
	// 	"a2": "us",               country of register (abbrv)
	// 	"al": 19,                 length
	// 	"aw": 8,                  width
	// 	"country": "USA",         country of register
	// 	"cu": 246.7,              course
	// 	"dest": "FALSE RIVER",    destination
	// 	"draught": 33,            draught
	// 	"dw": 0,                  deadweight
	// 	"etaTS": 1588620600,      ETA timestamp
	// 	"gt": 0,                  gross tonnage
	// 	"imo": 0,                 imo number
	// 	"lc.": 0,                 load condition(???)
	// 	"m9": 0,
	// 	"name": "SARAH REED",     name
	// 	"pic": "0-367003250-cf317c76a96fd9b9f5ae4679c64bd065", // path to thumbnail image https://static.vesselfinder.net/ship-photo/0-367003250-cf317c76a96fd9b9f5ae4679c64bd065/0
	// 	"r": 2,
	// 	"sc.": 0,                 status: 0=underway, 1=at anchor, 2=at anchor
	// 	"sl": false,              newer position available via satellite?
	// 	"ss": 0.1,                speed (knots)
	// 	"ts": 1587883051          timestamp (of position received?)
	// 	"type": "Towing vessel",  AIS type
	// 	"y": 0,                   year built
	// }

	// So far, we have only seen 0, 1, and 2. If we get any other values we
	// should alert on that.
	if web.ToInt(response["sc."]) > 2 || web.ToInt(response["sc."]) < 0 {
		fmt.Println("################# sc. > 2 || sc. < 0")
		fmt.Println(directLink(web.ToString(response["name"]), web.ToString(response["imo"]), mmsi))
		fmt.Println("  .ns:", response[".ns"])
		fmt.Println("  lc.:", response["lc."])
		fmt.Println("   m9:", response["m9"])
		fmt.Println("    r:", response["r"])
		fmt.Println("  sc.:", response["sc."])
		fmt.Println()
	}

	// I suspect that lc. is a sub-status of sc. That is, if sc. is != 0
	// (that is, the ship is at anchor) lc. will contain the details about
	// why sc. is not zero.
	if web.ToInt(response["lc."]) > 0 && web.ToInt(response["sc."]) < 1 {
		fmt.Println("################# lc. > 0 && sc. < 1")
		fmt.Println(directLink(web.ToString(response["name"]), web.ToString(response["imo"]), mmsi))
		fmt.Println("  .ns:", response[".ns"])
		fmt.Println("  lc.:", response["lc."])
		fmt.Println("   m9:", response["m9"])
		fmt.Println("    r:", response["r"])
		fmt.Println("  sc.:", response["sc."])
		fmt.Println()
	}

	// So far, we have only see m9==0 and r==2. Alert if there are any
	// other values.
	if web.ToInt(response["m9"]) != 0 || web.ToInt(response["r"]) != 2 {
		fmt.Println("################# m9 != 0 || r != 2")
		fmt.Println(directLink(web.ToString(response["name"]), web.ToString(response["imo"]), mmsi))
		fmt.Printf("  .ns: %d 0x%x\n", web.ToInt(response[".ns"]), web.ToInt(response[".ns"]))
		fmt.Printf("  lc.: %d 0x%x\n", web.ToInt(response["lc."]), web.ToInt(response["lc."]))
		fmt.Println("   m9:", response["m9"])
		fmt.Println("    r:", response["r"])
		fmt.Println("  sc.:", response["sc."])
		fmt.Println()
	}

	// What are all the possible values for .ns?
	ns := web.ToInt(response[".ns"])
	if ns != 15 && (ns < -1 || ns > 12) {
		fmt.Println("################# ns not in [-1..12, 15]")
		fmt.Println(directLink(web.ToString(response["name"]), web.ToString(response["imo"]), mmsi))
		fmt.Printf("  .ns: %d 0x%x\n", web.ToInt(response[".ns"]), web.ToInt(response[".ns"]))
		fmt.Printf("  lc.: %d 0x%x\n", web.ToInt(response["lc."]), web.ToInt(response["lc."]))
		fmt.Println("   m9:", response["m9"])
		fmt.Println("    r:", response["r"])
		fmt.Println("  sc.:", response["sc."])
		fmt.Println()
	}

	if mmsi == "319762000" && web.ToFloat64(response["ss"]) > 100.0 {
		// The M/Y SAINT NICHOLAS sometimes reports bad AIS data.
		// It takes the form of course: 360, speed: 102.3
		// Ignore those reports.
		fmt.Println("################### M/Y has high speed... what is her 'sc.'?")
		fmt.Println(directLink(web.ToString(response["name"]), web.ToString(response["imo"]), mmsi))
		fmt.Println(prettify(response))
		fmt.Println()
		return nil, false
	}

	return response, true
}
