package database

//
// Consumers of this package need to add this import line:
//
//	_ "github.com/go-sql-driver/mysql"
//

import (
	"database/sql"
	"fmt"
	"time"
)

// Ship holds all the information we get back from the web service about a single ship.
type Ship struct {
	// Stored in db ...
	MMSI       string
	IMO        string
	Name       string
	AIS        int // not currently available
	Type       string
	SAR        bool
	ID         string // deprecated
	VO         int    // deprecated
	FF         bool   // deprecated
	DirectLink string
	Draught    float64
	Year       int
	GT         int
	Sizes      string // deprecated
	Length     int
	Beam       int
	DW         int
	unknown    int // unused

	// Not stored in db ...
	Lat                float64
	Lon                float64
	ShipCourse         float64
	Speed              float64
	Sightings          int64
	NavigationalStatus int
}

// Sighting holds the relevant information about a ship sighting.
type Sighting struct {
	MMSI               string
	ShipCourse         float64
	Timestamp          int64
	Lat                float64
	Lon                float64
	MyLat              float64
	MyLon              float64
	NavigationalStatus int64
}

// NoaaDatum holds the information we get back from the NOAA web service.
type NoaaDatum struct {
	Station string
	Product string
	Datum   string
	Value   string
	S       string
	Flags   string
	// processing_level string // "p" - preliminary, "v" - verified
}

var (
	db *sql.DB
)

// Open opens a connection to the database.
func Open() error {
	var err error

	db, err = sql.Open("mysql", "ships:shipspassword@tcp(127.0.0.1:3306)/ship_ahoy")
	return err
}

// Close closes the connection to the database opened by Open.
func Close() {
	db.Close()
}

// SaveShip writes ship details to the database.
func SaveShip(details Ship) {
	sqlString := "INSERT IGNORE INTO ships ( mmsi, imo, name, ais, Type, sar, __id, vo, ff, direct_link, draught, year, gt, sizes, length, beam, dw ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )"

	_, err := db.Exec(sqlString, details.MMSI, details.IMO, details.Name, details.AIS, details.Type, details.SAR, details.ID, details.VO, details.FF, details.DirectLink, details.Draught, details.Year, details.GT, details.Sizes, details.Length, details.Beam, details.DW)
	if err != nil {
		fmt.Println("dbSaveShip Exec:", err)
	}
}

// LookupShip reads ship details from the database.
func LookupShip(mmsi string) (Ship, bool) {
	var details Ship

	sqlString := "SELECT * FROM ships WHERE mmsi = " + mmsi + " LIMIT 1"

	rows := db.QueryRow(sqlString)
	err := rows.Scan(&details.MMSI, &details.IMO, &details.Name, &details.AIS, &details.Type, &details.SAR, &details.ID, &details.VO, &details.FF, &details.DirectLink, &details.Draught, &details.Year, &details.GT, &details.Sizes, &details.Length, &details.Beam, &details.DW)
	if err != nil {
		if err != sql.ErrNoRows {
			fmt.Println("lookup_ship Scan:", err)
		}
		return details, false
	}

	details.Sightings = CountSightings(details.MMSI)

	return details, true
}

// LookupShipExists is [hopefully] faster than loading the entire record like dbLookupShip() does.
func LookupShipExists(mmsi string) bool {
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

// SaveSighting writes the ship sighting details to the database.
func SaveSighting(details Ship, myLat, myLon float64) {
	sqlString := "INSERT IGNORE INTO sightings ( mmsi, ship_course, timestamp, lat, lon, my_lat, my_lon ) VALUES ( ?, ?, ?, ?, ?, ?, ?)"

	_, err := db.Exec(sqlString, details.MMSI, details.ShipCourse, time.Now().Unix(), details.Lat, details.Lon, myLat, myLon)
	if err != nil {
		fmt.Println("dbSaveSighting Exec:", err)
	}
}

// LookupSighting reads sighting details from the database.
func LookupSighting(details Ship) (Sighting, bool) {
	var sighting Sighting

	sqlString := "SELECT * FROM sightings WHERE mmsi = " + details.MMSI + " ORDER BY timestamp DESC LIMIT 1"

	rows := db.QueryRow(sqlString)
	err := rows.Scan(&sighting.MMSI, &sighting.ShipCourse, &sighting.Timestamp, &sighting.Lat, &sighting.Lon, &sighting.MyLat, &sighting.MyLon)
	if err != nil {
		if err != sql.ErrNoRows {
			fmt.Println("lookup_sighting Scan:", err)
		}
		return sighting, false
	}

	return sighting, true
}

// LookupLastSighting is [hopefully] faster than dbLookupSighting() because it only queries the timestamp.
func LookupLastSighting(details Ship) (timestamp int64) {
	sqlString := "SELECT timestamp FROM sightings WHERE mmsi = " + details.MMSI + " ORDER BY timestamp DESC LIMIT 1"

	rows := db.QueryRow(sqlString)
	err := rows.Scan(&timestamp)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println("lookup_last_sighting Scan:", err)
	}

	return
}

// CountSightings counts the number of times we have seen this ship.
func CountSightings(mmsi string) (count int64) {
	sqlString := "SELECT COUNT(*) FROM sightings WHERE mmsi = " + mmsi

	rows := db.QueryRow(sqlString)
	err := rows.Scan(&count)
	if err != nil && err != sql.ErrNoRows {
		fmt.Println("lookup_last_sighting Scan:", err)
	}

	return
}

// CountRows returns the number of rows in the given table.
func CountRows(table string) (int64, bool) {
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

// TableStats prints interesting statistics about the size of the database.
func TableStats() map[string]int64 {
	tables := []string{"ships", "sightings"}
	counts := make(map[string]int64)

	for _, table := range tables {
		count, ok := CountRows(table)
		if ok {
			counts[table] = count
		}
	}

	return counts
}
