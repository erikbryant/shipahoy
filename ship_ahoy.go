package main

// $ go get github.com/go-sql-driver/mysql
//
// $ apt install libasound2-dev
// $ go get github.com/faiface/beep
// $ go get github.com/faiface/beep/mp3
// $ go get github.com/faiface/beep/wav
// $ go get github.com/faiface/beep/speaker

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"
)

type Ship struct {
	// Stored in db ...
	mmsi        string
	imo         string
	name        string
	ais         int
	Type        string
	sar         bool
	__id        string
	vo          int
	ff          bool
	direct_link string
	draught     float64
	year        int
	gt          int
	sizes       string
	length      int
	beam        int
	dw          int
	unknown     int // Unused.

	// Not stored in db ...
	lat         float64
	lon         float64
	ship_course float64
	speed       float64
}

type Sighting struct {
	mmsi        string
	ship_course float64
	timestamp   int64
	lat         float64
	lon         float64
	my_lat      float64
	my_lon      float64
}

type NoaaDatum struct {
	station string
	product string
	datum   string
	value   string
	s       string
	flags   string
	// processing_level string // "p" - preliminary, "v" - verified
}

var (
	cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

	db                *sql.DB
	uninteresting_ais = map[int]bool{
		0:  true, // Unknown
		6:  true, // Passenger
		31: true, // Tug
		36: true, // Sailing vessel
		37: true, // Pleasure craft
		52: true, // Tug
		60: true, // Passenger ship
		69: true, // Passenger ship
	}

	uninteresting_mmsi = map[string]bool{
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

func db_save_ship(details Ship) {
	sqlString := "INSERT IGNORE INTO ships ( mmsi, imo, name, ais, Type, sar, __id, vo, ff, direct_link, draught, year, gt, sizes, length, beam, dw ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )"

	_, err := db.Exec(sqlString, details.mmsi, details.imo, details.name, details.ais, details.Type, details.sar, details.__id, details.vo, details.ff, details.direct_link, details.draught, details.year, details.gt, details.sizes, details.length, details.beam, details.dw)
	if err != nil {
		fmt.Println("db_save_ship Exec:", err)
	}
}

func db_lookup_ship(mmsi string) (Ship, bool) {
	var details Ship

	sqlString := "SELECT * FROM ships WHERE mmsi = " + mmsi + " LIMIT 1"

	rows := db.QueryRow(sqlString)
	err := rows.Scan(&details.mmsi, &details.imo, &details.name, &details.ais, &details.Type, &details.sar, &details.__id, &details.vo, &details.ff, &details.direct_link, &details.draught, &details.year, &details.gt, &details.sizes, &details.length, &details.beam, &details.dw)
	if err != nil {
		if err != sql.ErrNoRows {
			fmt.Println("lookup_ship Scan:", err)
		}
		return details, false
	}

	return details, true
}

// db_lookup_ship_exists() is [hopefully] faster than loading the entire record like db_lookup_ship() does.
func db_lookup_ship_exists(mmsi string) bool {
	var exists int

	sqlString := "SELECT EXISTS( SELECT mmsi FROM ships WHERE mmsi = " + mmsi + " LIMIT 1 )"

	rows := db.QueryRow(sqlString)
	err := rows.Scan(&exists)
	if err != nil {
		if err != sql.ErrNoRows {
			fmt.Println("lookup_ship_exists Scan:", err)
		}
		return false
	}

	return exists == 1
}

func db_save_sighting(details Ship) {
	my_lat, my_lon := my_geo()

	sqlString := "INSERT IGNORE INTO sightings ( mmsi, ship_course, timestamp, lat, lon, my_lat, my_lon ) VALUES ( ?, ?, ?, ?, ?, ?, ?)"

	_, err := db.Exec(sqlString, details.mmsi, details.ship_course, time.Now().Unix(), details.lat, details.lon, my_lat, my_lon)
	if err != nil {
		fmt.Println("db_save_sighting Exec:", err)
	}
}

func db_lookup_sighting(details Ship) (Sighting, bool) {
	var sighting Sighting

	sqlString := "SELECT * FROM sightings WHERE mmsi = " + details.mmsi + " ORDER BY timestamp DESC LIMIT 1"

	rows := db.QueryRow(sqlString)
	err := rows.Scan(&sighting.mmsi, &sighting.ship_course, &sighting.timestamp, &sighting.lat, &sighting.lon, &sighting.my_lat, &sighting.my_lon)
	if err != nil {
		if err != sql.ErrNoRows {
			fmt.Println("lookup_sighting Scan:", err)
		}
		return sighting, false
	}

	return sighting, true
}

// db_lookup_last_sighting() is [hopefully] faster than db_lookup_sighting() because it only queries the timestamp.
func db_lookup_last_sighting(details Ship) (timestamp int64) {
	sqlString := "SELECT timestamp FROM sightings WHERE mmsi = " + details.mmsi + " ORDER BY timestamp DESC LIMIT 1"

	rows := db.QueryRow(sqlString)
	err := rows.Scan(&timestamp)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println("lookup_last_sighting Scan:", err)
	}

	return
}

func db_count_rows(table string) (int64, bool) {
	var count int64

	sqlString := "SELECT COUNT(*) FROM " + table

	row := db.QueryRow(sqlString)
	err := row.Scan(&count)
	if err != nil {
		if err != sql.ErrNoRows {
			fmt.Println("count_rows Scan:", err)
		}
		return 0, false
	}

	return count, true
}

func web_request(url string) string {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Do:", err)
		return ""
	}

	s, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("ReadAll:", err)
		return ""
	}

	return string(s)
}

func web_request_map(url string) map[string]interface{} {
	s := web_request(url)

	var m interface{}

	dec := json.NewDecoder(strings.NewReader(string(s)))
	err := dec.Decode(&m)
	if err != nil {
		fmt.Println("Decode:", err)
		return nil
	}

	// If the web request was successful we should get back a
	// map in JSON form. If it failed we should get back an error
	// message in string form. Make sure we got a map.
	f, ok := m.(map[string]interface{})
	if !ok {
		fmt.Println(string(s))
		return nil
	}

	return f
}

func decode_mmsi(mmsi string) string {
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
func alert(details Ship, url string) {
	fmt.Printf("\nShip Ahoy!     %s     %v\n\n", url, details, "-", decode_mmsi(details.mmsi))

	if strings.Contains(strings.ToLower(details.Type), "vehicle") {
		go play("meep.wav", true)
	} else {
		go play("ship_horn.mp3", false)
	}
}

func to_int(val interface{}) (result int) {
	switch val.(type) {
	case int:
		result = val.(int)
	case int64:
		result = int(val.(int64))
	case string:
		tmp, _ := strconv.ParseInt(val.(string), 10, 32)
		result = int(tmp)
	case float64:
		result = int(val.(float64))
	default:
		fmt.Println("Unknown type", val)
		result = val.(int) // Force a panic.
	}

	return result
}

func to_string(val interface{}) (result string) {
	switch val.(type) {
	case int:
		result = strconv.FormatInt(int64(val.(int)), 10)
	case int64:
		result = strconv.FormatInt(val.(int64), 10)
	case string:
		result = val.(string)
	case float64:
		result = strconv.FormatFloat(val.(float64), 'f', 8, 64)
	default:
		fmt.Println("Unknown type", val)
		result = val.(string) // Force a panic.
	}

	return result
}

func to_float64(val interface{}) (result float64) {
	switch val.(type) {
	case int:
		result = float64(val.(int))
	case int64:
		result = float64(val.(int64))
	case string:
		result, _ = strconv.ParseFloat(val.(string), 64)
	case float64:
		result = val.(float64)
	default:
		fmt.Println("Unknown type", val)
		result = val.(float64) // Force a panic.
	}

	return result
}

func get_ship_details(mmsi string, ais int) (Ship, bool) {
	var (
		length int64
		beam   int64
	)

	details, ok := db_lookup_ship(mmsi)
	if ok {
		return details, true
	}

	mmsi_url := "https://www.vesselfinder.com/clickinfo?mmsi=" + mmsi + "&rn=64229.85898456942&_=1524694015667"
	response := web_request_map(mmsi_url)
	if response == nil {
		return details, false
	}

	details.mmsi = mmsi
	details.imo = to_string(response["imo"])
	details.name = to_string(response["name"])
	details.ais = ais
	details.Type = to_string(response["type"])
	details.sar = response["sar"].(bool)
	details.__id = to_string(response["__id"])
	details.vo = to_int(response["vo"])
	details.ff = response["ff"].(bool)
	details.direct_link = to_string(response["direct_link"])
	details.draught = to_float64(response["draught"])
	details.year = to_int(response["year"])
	details.gt = to_int(response["gt"])
	details.sizes = to_string(response["sizes"])
	details.dw = to_int(response["dw"])

	sizes := strings.Split(details.sizes, " ")
	if len(sizes) == 4 && sizes[1] == "x" && sizes[3] == "m" {
		length, _ = strconv.ParseInt(sizes[0], 10, 64)
		beam, _ = strconv.ParseInt(sizes[2], 10, 64)
	}
	details.length = int(length)
	details.beam = int(beam)

	fmt.Println("Found:", details.mmsi, details.name, "\t-", decode_mmsi(details.mmsi))
	db_save_ship(details)

	return details, true
}

// visible_from_apt() returns a bool indicating whether the ship is visible
// from our apartment window.
func visible_from_apt(lat, lon float64) bool {
	// The bounding box for the area visible from our apartment.
	visible_latA := 37.8052
	visible_lonA := -122.48
	visible_latB := 37.8613
	visible_lonB := -122.4092

	// Note that A is the bottom left corner and B is the upper
	// right corner, so we need to work out C and D which are the
	// upper left and lower right corners.
	latC := visible_latB
	latD := visible_latA
	lonC := visible_lonA
	lonD := visible_lonB

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

func ships_in_region(latA, lonA, latB, lonB float64, c chan Ship) {
	defer close(c)

	latAs := strconv.FormatFloat(latA, 'f', 8, 64)
	lonAs := strconv.FormatFloat(lonA, 'f', 8, 64)
	latBs := strconv.FormatFloat(latB, 'f', 8, 64)
	lonBs := strconv.FormatFloat(lonB, 'f', 8, 64)

	url := "https://www.vesselfinder.com/vesselsonmap?bbox=" + lonAs + "%2C" + latAs + "%2C" + lonBs + "%2C" + latBs + "&zoom=12&mmsi=0&show_names=1&ref=35521.28976544603&pv=6"

	region := web_request(url)
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
		ship_course, _ := strconv.ParseFloat(fields[2], 64)
		ship_course /= 10.0
		speed, _ := strconv.ParseFloat(fields[3], 64)
		speed /= 10.0 // SOG
		ais, _ := strconv.ParseInt(fields[4], 10, 64)
		mmsi := fields[5]
		// name := fields[6]
		// unknown, _ := strconv.ParseInt(fields[7], 10, 64)

		details, ok := get_ship_details(mmsi, int(ais))
		if !ok {
			continue
		}

		details.lat = lat
		details.lon = lon
		details.ship_course = ship_course
		details.speed = speed

		// Push 'details' to channel.
		c <- details
	}
}

func look_at_ships(latA, lonA, latB, lonB float64) {
	// Open channel
	c := make(chan Ship, 10)

	go ships_in_region(latA, lonA, latB, lonB, c)

	// Read from channel
	for {
		// Read 'details' from channel.
		details, ok := <-c
		if !ok {
			break
		}

		// Only alert for ships that are moving.
		if details.speed < 1.0 {
			continue
		}

		// Skip 'uninteresting' ships.
		if uninteresting_ais[details.ais] || uninteresting_mmsi[details.mmsi] {
			continue
		}

		// Only alert for ships visible from our apartment.
		if !visible_from_apt(details.lat, details.lon) {
			continue
		}

		// If we have recently seen this ship, skip it.
		now := time.Now().Unix()
		elapsed := now - db_lookup_last_sighting(details)
		if elapsed < 30*60 {
			// The ship is still crossing the visible area.
			// No need to alert a second time.
			continue
		}

		// We have passed all the tests! Save and alert.
		db_save_sighting(details)
		url := "https://www.vesselfinder.com/?mmsi=" + details.mmsi + "&zoom=13"
		alert(details, url)
	}
}

func my_geo() (lat, lon float64) {
	my_ip := web_request("http://ifconfig.co/ip")
	my_ip = strings.TrimSpace(my_ip)
	url := "https://ipstack.com/ipstack_api.php?ip=" + my_ip
	location := web_request_map(url)
	lat = location["latitude"].(float64)
	lon = location["longitude"].(float64)
	return lat, lon
}

// bbox() returns a bounding box of the circle with center of the
// current location and radius of 'nmiles' nautical miles.
// Returns (latA, lonA, latB, lonB) Where A is the bottom left
// corner and B is the upper right corner.
func box(lat, lon float64, nmiles float64) (latA, lonA, latB, lonB float64) {
	// Convert nautical miles to decimal degrees.
	delta := nmiles / 60.0

	bbox_latA := lat - delta
	bbox_lonA := lon - delta
	bbox_latB := lat + delta
	bbox_lonB := lon + delta

	return bbox_latA, bbox_lonA, bbox_latB, bbox_lonB
}

func scan_nearby() {
	// TODO: If the bounding region of 'nearby' overlaps the bounding
	// region of scan_apt_visible then do not scan 'nearby',
	for {
		lat, lon := my_geo()
		latA, lonA, latB, lonB := box(lat, lon, 30)

		// Open channel.
		c := make(chan Ship, 10)

		go ships_in_region(latA, lonA, latB, lonB, c)

		// Read from channel.
		for {
			_, ok := <-c
			if !ok {
				break
			}
		}

		time.Sleep(5 * 60 * time.Second)
	}
}

func scan_apt_visible() {
	lat, lon := 37.82, -122.45 // Center of visible bay
	latA, lonA, latB, lonB := box(lat, lon, 10)

	for {
		look_at_ships(latA, lonA, latB, lonB)
		time.Sleep(2 * 60 * time.Second)
	}
}

func scan_planet() {
	for {
		step := 10
		lonA := float64(rand.Intn(360-step) - 180)
		latA := float64(rand.Intn(360-step) - 180)

		latB := latA + float64(step)
		lonB := lonA + float64(step)

		// Open channel.
		c := make(chan Ship, 10)

		go ships_in_region(latA, lonA, latB, lonB, c)

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

func tides() {
	reading := NoaaDatum{
		station: "9414290",
		product: "water_level",
		datum:   "mllw",
	}

	url := "https://tidesandcurrents.noaa.gov/api/datagetter?date=latest&station=" + reading.station + "&product=" + reading.product + "&datum=" + reading.datum + "&units=english&time_zone=lst_ldt&application=erikbryantology@gmail.com&format=json"

	for {
		response := web_request_map(url)
		data := response["data"].([]interface{})[0].(map[string]interface{})
		reading.value = data["v"].(string)
		reading.s = data["s"].(string)
		reading.flags = data["f"].(string)
		fmt.Println("Reading:", reading)
		time.Sleep(10 * 60 * time.Second)
	}
}

func air_gap() {
	reading := NoaaDatum{
		station: "9414304",
		product: "air_gap",
		datum:   "mllw",
	}

	url := "https://tidesandcurrents.noaa.gov/api/datagetter?date=latest&station=" + reading.station + "&product=" + reading.product + "&datum=" + reading.datum + "&units=english&time_zone=lst_ldt&application=erikbryantology@gmail.com&format=json"

	for {
		response := web_request_map(url)
		data := response["data"].([]interface{})[0].(map[string]interface{})
		reading.value = data["v"].(string)
		reading.s = data["s"].(string)
		reading.flags = data["f"].(string)
		fmt.Println("Air gap:", reading)
		time.Sleep(10 * 60 * time.Second)
	}
}

func db_stats() {
	tables := []string{"ships", "sightings"}

	for {
		msg := "## "
		for _, t := range tables {
			count, ok := db_count_rows(t)
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

	var err error

	db, err = sql.Open("mysql", "ships:shipspassword@tcp(127.0.0.1:3306)/ship_ahoy")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer db.Close()

	go scan_nearby()
	go scan_apt_visible()
	go scan_planet()
	go tides()
	go air_gap()
	go db_stats()

	for {
		time.Sleep(3 * 60 * time.Second)
		if *cpuprofile != "" {
			break
		}
	}
}