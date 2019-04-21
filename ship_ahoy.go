package main

// $ go get github.com/go-sql-driver/mysql
//
// $ apt install libasound2-dev
// $ go get github.com/faiface/beep
// $ go get github.com/faiface/beep/mp3
// $ go get github.com/faiface/beep/wav
// $ go get github.com/faiface/beep/speaker

import (
	"./database"
	"flag"
	"fmt"
	"github.com/erikbryant/aes"
	"github.com/erikbryant/web"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

	passPhrase     = flag.String("passPhrase", "", "Passphrase to unlock API key")
	geoAPIKeyCrypt = "2nC/f4XNjMo3Ddmn1b+aHed5ybr01za4plBCWjy+bjLkBIgT4+3QjtugSuq2iItxNRW9OodilLqQ7OG+"
	geoAPIKey      string

	myLat float64
	myLon float64

	uninterestingAIS = map[int]bool{
		0:  true, // Unknown
		6:  true, // Passenger
		31: true, // Tug
		36: true, // Sailing vessel
		37: true, // Pleasure craft
		52: true, // Tug
		60: true, // Passenger ship
		69: true, // Passenger ship
	}

	uninterestingMMSI = map[string]bool{
		"367123640": true, // Hawk
		"367389640": true, // Oski
		"366990520": true, // Del Norte
		"367566960": true, // F/V Pioneer
		"367469070": true, // Sunset Hornblower
		"338234637": true, // HEWESCRAFT 220 OP
		"366918840": true, // Happy Days
		"338107922": true, // Misty Dawn
		"367703860": true, // Vera Cruz
		"338236492": true, // Round Midnight
		"367517270": true, // Tesa
	}
)

func init() {
	rand.Seed(time.Now().Unix())
}

// MyGeo returns the lat/lon pair of the location of the computer running this program.
func MyGeo() (lat, lon float64) {
	// myIP := web.Request("http://ifconfig.co/ip") <-- site has malware
	// myIP := web.Request("https://api.ipify.org")
	// myIP = strings.TrimSpace(myIP)
	location, err := web.RequestJSON("http://api.ipstack.com/check?access_key=" + geoAPIKey)
	if err != nil {
		fmt.Println("ERROR: Unable to get geo location. Assuming you are home. Message:", err)
		return 37.8007, -122.4097
	}
	if location["error"] != nil {
		fmt.Println("ERROR: Unable to get geo location. Assuming you are home. Message:", location["error"])
		return 37.8007, -122.4097
	}
	lat = location["latitude"].(float64)
	lon = location["longitude"].(float64)
	return lat, lon
}

// decodeMmsi() returns a string describing the data encoded in the given MMSI.
func decodeMmsi(mmsi string) string {
	msg := ""

	// https://en.wikipedia.org/wiki/Maritime_Mobile_Service_Identity
	// https://en.wikipedia.org/wiki/Maritime_identification_digits
	// https://www.navcen.uscg.gov/?pageName=mtMmsi#format
	switch mmsi[0] {
	case '0':
		// Ship group, coast station, or group of coast stations
		msg += "Ship group or coast station "
	case '1':
		// For use by SAR aircraft (111MIDaxx)
		msg += "SAR aircraft "
	case '2':
		// MID: Europe
		msg += "Europe "
	case '3':
		// MID: North and Central America and Caribbean
		msg += "North/Central America "
	case '4':
		// MID: Asia
		msg += "Asia "
	case '5':
		// MID: Oceania
		msg += "Oceania "
	case '6':
		// MID: Africa
		msg += "Africa "
	case '7':
		// MID: South America
		msg += "South America "
	case '8':
		// Handheld VHF transceiver with DSC and GNSS
		msg += "Handheld VHF "
	case '9':
		// Devices using a free-form number identity:
		// Search and Rescue Transponders (970yyzzzz)
		// Man overboard DSC and/or AIS devices (972yyzzzz)
		// 406 MHz EPIRBs fitted with an AIS transmitter (974yyzzzz)
		// Craft associated with a parent ship (98MIDxxxx)
		// AtoN (aid to navigation) (99MIDaxxx)
		msg += "Misc "
	}

	// Trailing zeroes.
	//
	// If the ship is fitted with an Inmarsat A ship earth station, or has
	// satellite equipment other than Inmarsat, then the identity needs no
	// trailing zero.
	//
	// If the ship is fitted with an Inmarsat C ship earth station, or it is
	// expected to be so equipped in the foreseeable future, then the identity
	// could have one trailing zero:
	//
	// MIDxxxxx0
	//
	// If the ship is fitted with an Inmarsat B, C or M ship earth station,
	// or it is expected to be so equipped in the foreseeable future, then
	// the identity should have three trailing zeros:
	//
	// MIDxxx000
	if strings.HasSuffix(mmsi, "000") {
		msg += "[B/C/M] "
	} else {
		if strings.HasSuffix(mmsi, "0") {
			msg += "[C] "
		} else {
			msg += "[A or other] "
		}
	}

	return msg
}

// play() plays a given sound file.
func play(file string, wavFile bool) {
	// Open first sample File
	f, err := os.Open(file)

	// Check for errors when opening the file
	if err != nil {
		fmt.Println("Could not open audio file", file)
		return
	}

	// Decode the .mp3 File, if you have a .wav file, use wav.Decode(f)
	var (
		s      beep.StreamSeekCloser
		format beep.Format
	)
	if wavFile {
		s, format, _ = wav.Decode(f)
	} else {
		s, format, _ = mp3.Decode(f)
	}

	// Init the Speaker with the SampleRate of the format and a buffer size.
	speaker.Init(format.SampleRate, format.SampleRate.N(3*time.Second))

	// Channel, which will signal the end of the playback.
	playing := make(chan struct{})

	// Now we Play our Streamer on the Speaker
	speaker.Play(beep.Seq(s, beep.Callback(func() {
		// Callback after the stream Ends
		close(playing)
	})))
	<-playing
}

// alert() prints a message and plays an alert tone.
func alert(details database.Ship, url string) {
	fmt.Printf("\nShip Ahoy!     %s     %v - %s\n\n", url, details, decodeMmsi(details.MMSI))

	if strings.Contains(strings.ToLower(details.Type), "vehicle") {
		go play("meep.wav", true)
	} else {
		go play("ship_horn.mp3", false)
	}
}

// getShipDetails() retrieves ship details from the database, if they exist, or from the web if they do not.
func getShipDetails(mmsi string, ais int) (database.Ship, bool) {
	var (
		length int64
		beam   int64
	)

	details, ok := database.LookupShip(mmsi)
	if ok {
		return details, true
	}

	mmsiURL := "https://www.vesselfinder.com/clickinfo?mmsi=" + mmsi + "&rn=64229.85898456942&_=1524694015667"
	response, err := web.RequestJSON(mmsiURL)
	if err != nil || response == nil {
		return details, false
	}

	details.MMSI = mmsi
	details.IMO = web.ToString(response["imo"])
	details.Name = web.ToString(response["name"])
	details.AIS = ais
	details.Type = web.ToString(response["type"])
	details.SAR = response["sar"].(bool)
	details.ID = web.ToString(response["__id"])
	details.VO = web.ToInt(response["vo"])
	details.FF = response["ff"].(bool)
	details.DirectLink = web.ToString(response["direct_link"])
	details.Draught = web.ToFloat64(response["draught"])
	details.Year = web.ToInt(response["year"])
	details.GT = web.ToInt(response["gt"])
	details.Sizes = web.ToString(response["sizes"])
	details.DW = web.ToInt(response["dw"])

	sizes := strings.Split(details.Sizes, " ")
	if len(sizes) == 4 && sizes[1] == "x" && sizes[3] == "m" {
		length, _ = strconv.ParseInt(sizes[0], 10, 64)
		beam, _ = strconv.ParseInt(sizes[2], 10, 64)
	}
	details.Length = int(length)
	details.Beam = int(beam)

	fmt.Printf("Found: %s %-25s %s\n", details.MMSI, details.Name, decodeMmsi(details.MMSI))
	database.SaveShip(details)

	return details, true
}

// visibleFromApt() returns a bool indicating whether the ship is visible
// from our apartment window.
func visibleFromApt(lat, lon float64) bool {
	// The bounding box for the area visible from our apartment.
	visibleLatA := 37.8052
	visibleLonA := -122.48
	visibleLatB := 37.8613
	visibleLonB := -122.4092

	// Note that A is the bottom left corner and B is the upper
	// right corner, so we need to work out C and D which are the
	// upper left and lower right corners.
	latC := visibleLatB
	latD := visibleLatA
	lonC := visibleLonA
	lonD := visibleLonB

	// Is the ship within the bounding box of our visible area?
	if lat < latD || lat > latC {
		return false
	}
	if lon < lonC || lon > lonD {
		return false
	}

	// Is the ship within our visible triangle (the bottom left
	// triangle of the bounding box)? It is if the latitude is
	// less than the latitude of the box's diagonal.
	// x == longitude, y == latitude
	m := (latC - latD) / (lonC - lonD)
	b := latC - m*lonC
	y := m*lon + b
	if lat > y {
		return false
	}

	return true
}

// shipsInRegion() returns the ships found in a given lat/lon area via a channel.
func shipsInRegion(latA, lonA, latB, lonB float64, c chan database.Ship) {
	defer close(c)

	latAs := strconv.FormatFloat(latA, 'f', 8, 64)
	lonAs := strconv.FormatFloat(lonA, 'f', 8, 64)
	latBs := strconv.FormatFloat(latB, 'f', 8, 64)
	lonBs := strconv.FormatFloat(lonB, 'f', 8, 64)

	url := "https://www.vesselfinder.com/vesselsonmap?bbox=" + lonAs + "%2C" + latAs + "%2C" + lonBs + "%2C" + latBs + "&zoom=12&mmsi=0&show_names=1&ref=35521.28976544603&pv=6"

	region := web.Request(url)
	if len(region) < 10 {
		return
	}

	ships := strings.Split(region, "\n")
	for _, ship := range ships {
		fields := strings.Split(ship, "\t")
		// Skip the trailing line with its magic number.
		if len(fields) < 2 {
			continue
		}

		// https://api.vesselfinder.com/docs/response-ais.html
		lat, _ := strconv.ParseFloat(fields[0], 64)
		lat /= 600000.0
		lon, _ := strconv.ParseFloat(fields[1], 64)
		lon /= 600000.0
		shipCourse, _ := strconv.ParseFloat(fields[2], 64)
		shipCourse /= 10.0
		speed, _ := strconv.ParseFloat(fields[3], 64)
		speed /= 10.0 // SOG
		ais, _ := strconv.ParseInt(fields[4], 10, 64)
		mmsi := fields[5]
		// name := fields[6]
		// unknown, _ := strconv.ParseInt(fields[7], 10, 64)

		details, ok := getShipDetails(mmsi, int(ais))
		if !ok {
			continue
		}

		details.Lat = lat
		details.Lon = lon
		details.ShipCourse = shipCourse
		details.Speed = speed

		// Push 'details' to channel.
		c <- details
	}
}

// lookAtShips() looks for interesting ships in a given lat/lon region.
func lookAtShips(latA, lonA, latB, lonB float64) {
	// Open channel
	c := make(chan database.Ship, 10)

	go shipsInRegion(latA, lonA, latB, lonB, c)

	// Read from channel
	for {
		// Read 'details' from channel.
		details, ok := <-c
		if !ok {
			break
		}

		// Only alert for ships that are moving.
		if details.Speed < 1.0 {
			continue
		}

		// Skip 'uninteresting' ships.
		if uninterestingAIS[details.AIS] || uninterestingMMSI[details.MMSI] {
			continue
		}

		// Only alert for ships visible from our apartment.
		if !visibleFromApt(details.Lat, details.Lon) {
			continue
		}

		// If we have recently seen this ship, skip it.
		now := time.Now().Unix()
		elapsed := now - database.LookupLastSighting(details)
		if elapsed < 30*60 {
			// The ship is still crossing the visible area.
			// No need to alert a second time.
			continue
		}

		// We have passed all the tests! Save and alert.
		database.SaveSighting(details, myLat, myLon)
		url := "https://www.vesselfinder.com/?mmsi=" + details.MMSI + "&zoom=13"
		alert(details, url)
	}
}

// box() returns a bounding box of the circle with center of the
// current location and radius of 'nmiles' nautical miles.
// Returns (latA, lonA, latB, lonB) Where A is the bottom left
// corner and B is the upper right corner.
func box(lat, lon float64, nmiles float64) (latA, lonA, latB, lonB float64) {
	// Convert nautical miles to decimal degrees.
	delta := nmiles / 60.0

	bboxLatA := lat - delta
	bboxLonA := lon - delta
	bboxLatB := lat + delta
	bboxLonB := lon + delta

	return bboxLatA, bboxLonA, bboxLatB, bboxLonB
}

// scanNearby() continually scans for ships within a given radius of this computer.
func scanNearby() {
	// TODO: If the bounding region of 'nearby' overlaps the bounding
	// region of scan_apt_visible then do not scan 'nearby',
	lat, lon := MyGeo()
	latA, lonA, latB, lonB := box(lat, lon, 30)

	// Open channel.
	c := make(chan database.Ship, 10)

	for {
		go shipsInRegion(latA, lonA, latB, lonB, c)

		// Read from channel.
		for {
			_, ok := <-c
			if !ok {
				break
			}

			// TODO: Add alerting to notify ships are near.
		}

		time.Sleep(5 * 60 * time.Second)
	}
}

// scanAptVisible() continually scans for ships visible from our apartment.
func scanAptVisible() {
	lat, lon := 37.82, -122.45 // Center of visible bay
	latA, lonA, latB, lonB := box(lat, lon, 10)

	for {
		lookAtShips(latA, lonA, latB, lonB)
		time.Sleep(2 * 60 * time.Second)
	}
}

// scanPlanet() continually scans the entire planet for heretofore unseen ships.
func scanPlanet() {
	for {
		step := 10
		lonA := float64(rand.Intn(360-step) - 180)
		latA := float64(rand.Intn(360-step) - 180)

		latB := latA + float64(step)
		lonB := lonA + float64(step)

		// Open channel.
		c := make(chan database.Ship, 10)

		go shipsInRegion(latA, lonA, latB, lonB, c)

		// Read from channel.
		for {
			_, ok := <-c
			if !ok {
				break
			}
		}

		time.Sleep(5 * time.Second)
	}
}

// tides() looks up instantaneous tide data for a given NOAA station.
func tides() {
	reading := database.NoaaDatum{
		Station: "9414290",
		Product: "water_level",
		Datum:   "mllw",
	}

	url := "https://tidesandcurrents.noaa.gov/api/datagetter?date=latest&station=" + reading.Station + "&product=" + reading.Product + "&datum=" + reading.Datum + "&units=english&time_zone=lst_ldt&application=erikbryantology@gmail.com&format=json"

	for {
		response, err := web.RequestJSON(url)
		if err != nil {
			fmt.Println("Unable to get tide data: ", err)
			continue
		}
		if response["error"] != nil {
			fmt.Println("Unable to get tide data: ", response["error"])
			continue
		}
		data := response["data"].([]interface{})[0].(map[string]interface{})
		reading.Value = data["v"].(string)
		reading.S = data["s"].(string)
		reading.Flags = data["f"].(string)
		fmt.Println("Reading:", reading)
		time.Sleep(10 * 60 * time.Second)
	}
}

// airGap() looks up instantaneous air gap (distance from bottom of bridge to water) for a given NOAA station.
func airGap() {
	reading := database.NoaaDatum{
		Station: "9414304",
		Product: "air_gap",
		Datum:   "mllw",
	}

	url := "https://tidesandcurrents.noaa.gov/api/datagetter?date=latest&station=" + reading.Station + "&product=" + reading.Product + "&datum=" + reading.Datum + "&units=english&time_zone=lst_ldt&application=erikbryantology@gmail.com&format=json"

	for {
		response, err := web.RequestJSON(url)
		if err != nil {
			fmt.Println("Unable to get air gap data: ", err)
			continue
		}
		if response["error"] != nil {
			// fmt.Println("Unable to get air gap data: ", response["error"])
			continue
		}
		data := response["data"].([]interface{})[0].(map[string]interface{})
		reading.Value = data["v"].(string)
		reading.S = data["s"].(string)
		reading.Flags = data["f"].(string)
		fmt.Println("Air gap:", reading)
		time.Sleep(10 * 60 * time.Second)
	}
}

// dbStats() prints interesting statistics about the size of the database.
func dbStats() {
	tables := []string{"ships", "sightings"}

	for {
		msg := "## "
		for _, t := range tables {
			count, ok := database.CountRows(t)
			if ok {
				msg += t + ": " + strconv.FormatInt(count, 10) + " "
			}
		}
		msg += "##"
		fmt.Println(msg)
		time.Sleep(10 * 60 * time.Second)
	}
}

func main() {
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	geoAPIKey = aes.Decrypt(geoAPIKeyCrypt, *passPhrase)
	myLat, myLon = MyGeo()

	database.Open()
	defer database.Close()

	// go scanNearby()
	go scanAptVisible()
	go scanPlanet()
	go tides()
	go airGap()
	go dbStats()

	for {
		time.Sleep(3 * 60 * time.Second)
		if *cpuprofile != "" {
			break
		}
	}
}
