package main

// $ go get github.com/go-sql-driver/mysql
//
// $ apt install libasound2-dev
// $ go get github.com/faiface/beep
// $ go get github.com/faiface/beep/mp3
// $ go get github.com/faiface/beep/wav
// $ go get github.com/faiface/beep/speaker
//
// $ go get github.com/erikbryant/aes
// $ go get github.com/erikbryant/web

import (
	"bytes"
	"cloud.google.com/go/texttospeech/apiv1"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/erikbryant/aes"
	"github.com/erikbryant/ship_ahoy/database"
	"github.com/erikbryant/web"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	_ "github.com/go-sql-driver/mysql"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
	"log"
	"math"
	"math/rand"
	"os"
	"regexp"
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

	gcpAuthCrypt = "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"

	myLat float64
	myLon float64

	interestingMMSI = map[string]bool{
		"338691000": true, // Matthew Turner
		"338359786": true, // Randy S Cummings (associated with Pacific Responder?)
	}

	uninterestingAIS = map[string]bool{
		"Fishing vessel": true,
		"Passenger ship": true,
		"Passenger Ship": true,
		"Pleasure craft": true,
		"Sailing vessel": true,
		"Towing vessel":  true,
		"Tug":            true,
		"Unknown":        true,
	}

	uninterestingMMSI = map[string]bool{
		"367123640": true, // Hawk
		"367389640": true, // Oski
		"366990520": true, // Del Norte
		"367566960": true, // F/V Pioneer
		"367469070": true, // Sunset Hornblower
		"338234637": true, // Hewescraft 220 OP
		"366918840": true, // Happy Days
		"338107922": true, // Misty Dawn
		"367703860": true, // Vera Cruz
		"338236492": true, // Round Midnight
		"367517270": true, // Tesa
		"367533950": true, // Sausalito Bmpress
		"366831930": true, // Millennium
		"366864140": true, // Naiad
		"338303816": true, // Coastal 24
		"319421001": true, // RIB 45
	}
)

// init performs the pre-flight setup.
func init() {
	rand.Seed(time.Now().Unix())
}

func decodeGcpAuth() {
	flag.Parse()

	path := os.TempDir() + "/" + "gcp_auth.json"

	gcpAuth := aes.Decrypt(gcpAuthCrypt, *passPhrase)
	err := writeFile(path, gcpAuth)
	if err != nil {
		panic(err)
	}

	// The Google API looks for the auth using an env var.
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", path)
}

// writeFile creates a file and writes the given content to it.
func writeFile(file, content string) error {
	f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.WriteString(content + "\n")
	if err != nil {
		return err
	}

	return nil
}

// myGeo returns the lat/lon pair of the location of the computer running this program.
func myGeo() (lat, lon float64) {
	// myIP := web.Request("http://ifconfig.co/ip") <-- site has malware
	location, err := web.RequestJSON("http://api.ipstack.com/check?access_key=" + geoAPIKey)
	if err != nil {
		fmt.Println("ERROR: Unable to get geo location. Assuming you are home. Message:", err)
		return 37.8007, -122.4097
	}
	if location["error"] != nil {
		fmt.Println("ERROR: Error getting geo location. Assuming you are home. Message:", location["error"])
		return 37.8007, -122.4097
	}
	lat = location["latitude"].(float64)
	lon = location["longitude"].(float64)
	return lat, lon
}

// validateMmsi tests whether an MMSI is valid.
func validateMmsi(mmsi string) error {
	if len(mmsi) != 9 {
		return fmt.Errorf("MMSI length != 9: %s", mmsi)
	}

	matched, err := regexp.MatchString("^[0-9]{9}$", mmsi)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("Invalid MMSI found: %s", mmsi)
	}

	return nil
}

// decodeMmsi returns a string describing the data encoded in the given MMSI.
func decodeMmsi(mmsi string) string {
	err := validateMmsi(mmsi)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	msg := ""
	mid := mmsi[0:3]

	// https://en.wikipedia.org/wiki/Maritime_Mobile_Service_Identity
	// https://en.wikipedia.org/wiki/Maritime_identification_digits
	// https://www.navcen.uscg.gov/?pageName=mtMmsi#format
	switch mmsi[0] {
	case '0':
		// 0: Ship group
		// 00: Coast radio station
		if mmsi[1] == '0' {
			msg += "Coast radio station "
			mid = mmsi[2:5]
		} else {
			msg += "Ship group "
			mid = mmsi[1:4]
		}
	case '1':
		// 111: For use by SAR aircraft (111MIDaxx)
		if mmsi[0:3] == "111" {
			msg += "SAR aircraft "
			mid = mmsi[3:6]
		} else {
			msg += "Unknown type 1 "
			mid = mmsi[1:4]
		}
	case '8':
		// Handheld VHF transceiver with DSC and GNSS
		msg += "Handheld VHF "
		mid = mmsi[1:4]
	case '9':
		// Devices using a free-form number identity:
		//   Search and Rescue Transponders (970yyzzzz)
		//   Man overboard DSC and/or AIS devices (972yyzzzz)
		//   406 MHz EPIRBs fitted with an AIS transmitter (974yyzzzz)
		//   Craft associated with a parent ship (98MIDxxxx)
		//   AtoN (aid to navigation) (99MIDaxxx)
		msg += "Misc "

		switch mmsi[0:2] {
		case "98":
			msg += "craft associated with parent ship "
			mid = mmsi[2:5]
		case "99":
			msg += "aid to navigation "
			mid = mmsi[2:5]
		default:
			switch mmsi[0:3] {
			case "970":
				msg += "SAR transponder "
				mid = mmsi[3:6]
			case "972":
				msg += "man overboard device "
				mid = mmsi[3:6]
			case "974":
				msg += "EPIRB with AIS transmitter "
				mid = mmsi[3:6]
			}
		}
	}

	switch mid[0] {
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
	default:
		msg += "Invalid MID " + mid + " "
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
		msg += "[Immarsat B/C/M]"
	} else {
		if strings.HasSuffix(mmsi, "0") {
			msg += "[Immarsat C] "
		} else {
			msg += "[Immarsat A or other] "
		}
	}

	return strings.TrimSpace(msg)
}

// playStream plays a given audio stream.
func playStream(s beep.StreamSeekCloser, format beep.Format) {
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

// play plays a given sound file. MP3 and WAV are supported.
func play(file string) {
	f, err := os.Open(file)
	if err != nil {
		fmt.Println("Could not open audio file", file)
		return
	}

	var (
		s      beep.StreamSeekCloser
		format beep.Format
	)

	if strings.HasSuffix(file, ".wav") {
		s, format, err = wav.Decode(f)
		if err != nil {
			fmt.Println("Could not decode WAV audio file", file, err)
			return
		}
	} else {
		s, format, err = mp3.Decode(f)
		if err != nil {
			fmt.Println("Could not decode MP3 audio file", file, err)
			return
		}
	}

	playStream(s, format)
}

// readable makes a string more human readable by removing all non alphanumeric and non-punctuation.
func readable(text string) string {
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "_", " ")
	text = strings.ReplaceAll(text, "/", "")
	text = strings.ReplaceAll(text, "^", "")
	text = strings.ReplaceAll(text, "[", " ")
	text = strings.ReplaceAll(text, "]", " ")
	text = strings.ReplaceAll(text, "\"", "")

	// Specific ships we have seen that read poorly.
	text = strings.ReplaceAll(text, "SEASPAN HAMBURG", "SEA SPAN HAMBURG")

	return text
}

// say converts text to speech and then plays it.
func say(text string) {
	text = readable(text)

	ctx := context.Background()

	c, err := texttospeech.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	req := &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{
				Text: text,
			},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			// https://cloud.google.com/text-to-speech/docs/voices
			LanguageCode: "en-US",
			Name:         "en-US-Standard-C",
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_LINEAR16,
			SpeakingRate:  1.0,
		},
	}
	resp, err := c.SynthesizeSpeech(ctx, req)
	if err != nil {
		log.Fatal(err)
	}

	r := bytes.NewReader(resp.GetAudioContent())
	s, format, err := wav.Decode(r)
	if err != nil {
		fmt.Println("Could not decode WAV speech stream", text, err)
		return
	}
	playStream(s, format)
}

// prettify formats and prints the input.
func prettify(i interface{}) string {
	s, err := json.MarshalIndent(i, "", " ")
	if err != nil {
		fmt.Println("Could not Marshal object", i)
	}

	return string(s)
}

// alert prints a message and plays an alert tone.
func alert(details database.Ship) {
	fmt.Printf(
		"\nShip Ahoy!  %s  %s\n%+v\n\n",
		time.Now().Format("Mon Jan 2 15:04:05"),
		decodeMmsi(details.MMSI),
		prettify(details),
	)

	if strings.Contains(strings.ToLower(details.Type), "vehicle") {
		play("meep.wav")
	} else if strings.Contains(strings.ToLower(details.Type), "pilot") {
		play("pilot.mp3")
	} else {
		play("ship_horn.mp3")
	}

	summary := fmt.Sprintf("Ship ahoy! %s. %s. Course %3.f degrees.", details.Name, details.Type, details.ShipCourse)

	// Hearing, "eleven point zero knots" sounds awkward. Remove the "point zero".
	if math.Trunc(details.Speed) == details.Speed {
		summary = fmt.Sprintf("%s Speed %3.0f knots.", summary, math.Trunc(details.Speed))
	} else {
		summary = fmt.Sprintf("%s Speed %3.1f knots.", summary, details.Speed)
	}

	switch details.Sightings {
	case 0:
		summary = fmt.Sprintf("%s This is the first sighting.", summary)
	case 1:
		summary = fmt.Sprintf("%s One previous sighting.", summary)
	default:
		summary = fmt.Sprintf("%s %d previous sightings.", summary, details.Sightings)
	}

	say(summary)
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

// visibleFromApt returns a bool indicating whether the ship is visible
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

	region := web.Request(url)
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

// lookAtShips looks for interesting ships in a given lat/lon region.
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
		if details.Speed < 2.0 {
			continue
		}

		// Always show interesting ships.
		if !interestingMMSI[details.MMSI] {
			// Skip uninteresting ships.
			if uninterestingAIS[details.Type] || uninterestingMMSI[details.MMSI] {
				continue
			}
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
		alert(details)
	}
}

// box returns a bounding box of the circle with center of the
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

// scanNearby continually scans for ships within a given radius of this computer.
func scanNearby(sleepSecs time.Duration) {
	// TODO: If the bounding region of 'nearby' overlaps the bounding
	// region of scan_apt_visible then do not scan 'nearby'.
	lat, lon := myGeo()
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

		time.Sleep(sleepSecs)
	}
}

// scanAptVisible continually scans for ships visible from our apartment.
func scanAptVisible(sleepSecs time.Duration) {
	lat, lon := 37.82, -122.45 // Center of visible bay
	latA, lonA, latB, lonB := box(lat, lon, 10)

	for {
		lookAtShips(latA, lonA, latB, lonB)
		time.Sleep(sleepSecs)
	}
}

// scanPlanet continually scans the entire planet for heretofore unseen ships.
func scanPlanet(sleepSecs time.Duration) {
	for {
		// Pick a random lat/lon box on the surface of the planet.
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

		time.Sleep(sleepSecs)
	}
}

// noaaReading reads one datum from a given NOAA station.
func noaaReading(url string, reading *database.NoaaDatum) bool {
	response, err := web.RequestJSON(url)
	if err != nil {
		fmt.Println("Error getting data: ", err)
		return false
	}
	if response["error"] != nil {
		fmt.Println("Error reading data: ", response["error"])
		return false
	}
	data := response["data"].([]interface{})[0].(map[string]interface{})
	reading.Value = data["v"].(string)
	reading.S = data["s"].(string)
	reading.Flags = data["f"].(string)
	return true
}

// tides looks up instantaneous tide data for a given NOAA station.
func tides(sleepSecs time.Duration) {
	reading := database.NoaaDatum{
		Station: "9414290",
		Product: "water_level",
		Datum:   "mllw",
	}

	url := "https://tidesandcurrents.noaa.gov/api/datagetter?date=latest&station=" + reading.Station + "&product=" + reading.Product + "&datum=" + reading.Datum + "&units=english&time_zone=lst_ldt&application=erikbryantology@gmail.com&format=json"

	for {
		// Sleep at the start of the loop to avoid spamming the API
		// in the case where the API is returning errors
		time.Sleep(sleepSecs)

		ok := noaaReading(url, &reading)
		if !ok {
			continue
		}
		fmt.Println("Reading:", reading)
	}
}

// airGap looks up instantaneous air gap (distance from bottom of bridge to water) for a given NOAA station.
func airGap(sleepSecs time.Duration) {
	reading := database.NoaaDatum{
		Station: "9414304",
		Product: "air_gap",
		Datum:   "mllw",
	}

	url := "https://tidesandcurrents.noaa.gov/api/datagetter?date=latest&station=" + reading.Station + "&product=" + reading.Product + "&datum=" + reading.Datum + "&units=english&time_zone=lst_ldt&application=erikbryantology@gmail.com&format=json"

	for {
		// Sleep at the start of the loop to avoid spamming the API
		// in the case where the API is returning errors
		time.Sleep(sleepSecs)

		ok := noaaReading(url, &reading)
		if !ok {
			continue
		}
		fmt.Println("Reading:", reading)
	}
}

// dbStats prints interesting statistics about the size of the database.
func dbStats(sleepSecs time.Duration) {
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
		time.Sleep(sleepSecs)
	}
}

func main() {
	flag.Parse()

	decodeGcpAuth()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	geoAPIKey = aes.Decrypt(geoAPIKeyCrypt, *passPhrase)
	myLat, myLon = myGeo()

	database.Open()
	defer database.Close()

	// go scanNearby(5 * 60 * time.Second)
	go scanAptVisible(2 * 60 * time.Second)
	go scanPlanet(2 * 60 * time.Second)
	go tides(10 * 60 * time.Second)
	go airGap(10 * 60 * time.Second)
	go dbStats(10 * 60 * time.Second)

	for {
		time.Sleep(3 * 60 * time.Second)
		if *cpuprofile != "" {
			break
		}
	}
}
