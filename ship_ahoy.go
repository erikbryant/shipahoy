package main

// $ go get github.com/go-sql-driver/mysql
import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"net/http"
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
	unknown     int

	// Not stored in db ...
	lat   float64
	lon   float64
	speed float64
}

var (
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
	}
)

func db_lookup_ship(mmsi string) (Ship, bool) {
	var details Ship

	sqlString := "select * from ships where mmsi = " + mmsi

	rows := db.QueryRow(sqlString)
	err := rows.Scan(&details.mmsi, &details.imo, &details.name, &details.ais, &details.Type, &details.sar, &details.__id, &details.vo, &details.ff, &details.direct_link, &details.draught, &details.year, &details.gt, &details.sizes, &details.length, &details.beam, &details.dw, &details.unknown)
	if err != nil {
		if err != sql.ErrNoRows {
			fmt.Println("lookup_ship Scan:", err)
		}
		return details, false
	}

	return details, true
}

func db_count_rows(table string) (int, bool) {
	rows, err := db.Query("select count(*) from " + table)
	if err != nil {
		fmt.Println("Query:", err)
		return 0, false
	}
	defer rows.Close()
	var (
		count int
	)
	for rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			fmt.Println("count_rows Scan:", err)
			return 0, false
		}
	}
	err = rows.Err()
	if err != nil {
		fmt.Println("Err:", err)
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

// alert() prints a message and plays an alert tone.
func alert(details Ship, url string) {
	fmt.Println("Ship Ahoy!     ", url)
	fmt.Println(details)

	// TODO play tone.
}

func ship_details(mmsi string, ais int) (Ship, bool) {
	var (
		length int64
		beam   int64
	)

	details, ok := db_lookup_ship(mmsi)
	if ok {
		if details.ais != ais {
			update(mmsi, "ais", ais)
			details.ais = ais
		}
		return details, true
	}

	mmsi_url := "https://www.vesselfinder.com/clickinfo?mmsi=" + mmsi + "&rn=64229.85898456942&_=1524694015667"
	response := web_request_map(mmsi_url)
	if response == nil {
		return details, false
	}

	// Copy response into details.
	details.imo = response["imo"].(string)
	details.name = response["name"].(string)
	details.Type = response["type"].(string)
	details.sar = response["sar"].(bool)
	details.__id = response["__id"].(string)
	details.vo = int(response["vo"].(float64))
	details.ff = response["ff"].(bool)
	details.direct_link = response["direct_link"].(string)
	details.draught = response["draught"].(float64)
	year, _ := strconv.ParseInt(response["year"].(string), 10, 32)
	details.year = int(year)
	gt, _ := strconv.ParseInt(response["gt"].(string), 10, 32)
	details.gt = int(gt)
	details.sizes = response["sizes"].(string)
	dw, _ := strconv.ParseInt(response["dw"].(string), 10, 32)
	details.dw = int(dw)
	// details.unknown = response["unknown"].(int)

	// Add the extra fields to details.
	details.mmsi = mmsi
	details.ais = ais
	fmt.Println("Found new ship:", details.mmsi, details.name)
	sizes := strings.Split(details.sizes, " ")
	if len(sizes) == 4 && sizes[1] == "x" && sizes[3] == "m" {
		length, _ = strconv.ParseInt(sizes[0], 10, 64)
		beam, _ = strconv.ParseInt(sizes[2], 10, 64)
	}
	details.length = int(length)
	details.beam = int(beam)
	persist_ship(details)

	return details, true
}

// TODO
func update(_ string, _ string, _ int) {
	return
}

// TODO
func persist_ship(_ Ship) {
	return
}

// TODO
func persist_sighting(Ship) {
	return
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

		details, ok := ship_details(mmsi, int(ais))
		if !ok {
			continue
		}

		details.lat = lat
		details.lon = lon
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
		if details.speed < 4.0 {
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
		// TODO

		// We have passed all the tests! Save and alert.
		persist_sighting(details)
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
	lat, lon := my_geo()
	latA, lonA, latB, lonB := box(lat, lon, 50)

	for {
		fmt.Println("Scanning nearby ...", latA, lonA, latB, lonB)
		// TODO: call something like look_at_ships, but without
		// the check for sightings.
		// look_at_ships(latA, lonA, latB, lonB)
		time.Sleep(10 * time.Second)
	}
}

func scan_apt_visible() {
	lat, lon := 37.82, -122.45 // Center of visible bay
	latA, lonA, latB, lonB := box(lat, lon, 10)

	for {
		fmt.Println("Scanning apt visible ...", latA, lonA, latB, lonB)
		look_at_ships(latA, lonA, latB, lonB)
		time.Sleep(15 * time.Second)
	}
}

func scan_planet() {
	for {
		for latA := 60.0; latA >= -70.0; latA -= 10.0 {
			for lonA := 180.0; lonA >= -180.0; lonA -= 10.0 {
				latB := latA + 10.0
				lonB := lonA + 10.0
				fmt.Println("Scanning planet ...", latA, lonA, latB, lonB)
				look_at_ships(latA, lonA, latB, lonB)
				time.Sleep(5 * time.Second)
			}
			time.Sleep(30 * time.Second)
		}
	}
}

func db_stats() {
	tables := []string{"ships", "sightings"}

	for {
		for _, t := range tables {
			count, ok := db_count_rows(t)
			if ok {
				fmt.Printf("%s: %d\n", t, count)
			}
		}
		time.Sleep(5 * 60 * time.Second)
	}
}

func main() {
	var (
		err error
	)

	db, err = sql.Open("mysql", "ships:shipspassword@tcp(127.0.0.1:3306)/ship_ahoy")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer db.Close()

	// go scan_nearby()
	go scan_apt_visible()
	// go scan_planet()
	go db_stats()

	for {
		time.Sleep(3600 * time.Second)
	}
}
