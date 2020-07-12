package ship_ahoy

import (
	"fmt"
	"github.com/erikbryant/aes"
	"github.com/erikbryant/beepspeak"
	"github.com/erikbryant/ship_ahoy/database"
	"github.com/erikbryant/web"
	_ "github.com/go-sql-driver/mysql"
	"math/rand"
	"strconv"
	"time"
)

var (
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
		"338099564": true, // Wings
	}
)

// init performs the pre-flight setup.
func init() {
	rand.Seed(time.Now().Unix())
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
	for {
		stats := database.TableStats()

		if len(stats) > 0 {
			msg := "## "
			for table, count := range stats {
				msg += table + ": " + strconv.FormatInt(count, 10) + " "
			}
			msg += "##"
			fmt.Println(msg)
		}

		time.Sleep(sleepSecs)
	}
}

func Doit(passPhrase string, cpuprofile string) {
	var err error

	geoAPIKey, err = aes.Decrypt(geoAPIKeyCrypt, passPhrase)
	if err != nil {
		fmt.Println(err)
		return
	}
	myLat, myLon = myGeo()

	beepspeak.InitSay(gcpAuthCrypt, passPhrase)

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
		if cpuprofile != "" {
			break
		}
	}

}
