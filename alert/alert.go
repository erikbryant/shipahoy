package alert

// $ apt install libasound2-dev

import (
	"encoding/json"
	"fmt"
	"github.com/erikbryant/beepspeak"
	"github.com/erikbryant/shipahoy/aismmsi"
	"github.com/erikbryant/shipahoy/database"
	"math"
	"strings"
	"time"
)

// readableName makes a string more human readable by removing all non alphanumeric and non-punctuation.
func readableName(text string) string {
	text = strings.TrimSpace(text)

	// Symbols do not usually read well.
	text = strings.ReplaceAll(text, "_", " ")
	text = strings.ReplaceAll(text, "/", "")
	text = strings.ReplaceAll(text, "^", "")
	text = strings.ReplaceAll(text, "[", " ")
	text = strings.ReplaceAll(text, "]", " ")
	text = strings.ReplaceAll(text, "\"", "")

	// Specific ships we have seen that read poorly.
	text = strings.ReplaceAll(text, "SEASPAN HAMBURG", "SEA SPAN HAMBURG")
	text = strings.ReplaceAll(text, "RO-ROZAFER", "RO-RO ZAFER")
	text = strings.ReplaceAll(text, "RO-RO ", "ROW ROW ")
	text = strings.ReplaceAll(text, "ERISORT", "AIRY SORT")
	text = strings.ReplaceAll(text, "T0WBOATUS_ALAMEDA", "TOW BOAT U S ALAMEDA")
	text = strings.ReplaceAll(text, "SAILDRONE", "SAIL DRONE")
	text = strings.ReplaceAll(text, "JOHNZIPSER", "JOHN ZIPSER")

	return text
}

// readableCourse takes a compass heading and formats it in a human speakable way.
func readableCourse(heading float64) string {
	course := int(math.Round(heading))
	courseText := ""

	if course%100 == 0 {
		courseText = fmt.Sprintf("%d", course)
	} else if course < 100 {
		courseText = fmt.Sprintf("%d", course)
	} else {
		var hundreds int
		hundreds = course / 100
		tens := course % 100
		if tens < 10 {
			courseText = fmt.Sprintf("%dO%d", hundreds, tens)
		} else {
			courseText = fmt.Sprintf("%d %d", hundreds, tens)
		}
	}

	return courseText
}

// prettify formats and prints the input.
func prettify(i interface{}) string {
	s, err := json.MarshalIndent(i, "", " ")
	if err != nil {
		fmt.Println("Could not Marshal object", i)
	}

	return string(s)
}

// Alert prints a message and plays an alert tone.
func Alert(details database.Ship) error {
	fmt.Printf(
		"\nShip Ahoy!  %s  %s\n%+v\n\n",
		time.Now().Format("Mon Jan 2 15:04:05"),
		aismmsi.DecodeMmsi(details.MMSI),
		prettify(details),
	)

	var sound string
	if strings.Contains(strings.ToLower(details.Type), "vehicle") {
		sound = "meep.wav"
	} else if strings.Contains(strings.ToLower(details.Type), "pilot") {
		sound = "pilot.mp3"
	} else {
		sound = "horn.mp3"
	}
	err := beepspeak.Play("./alert/" + sound)
	if err != nil {
		return err
	}

	summary := fmt.Sprintf("Ship ahoy! %s. %s. Course %s degrees.", readableName(details.Name), details.Type, readableCourse(details.ShipCourse))

	// Hearing, "eleven point zero knots" sounds awkward. Remove the "point zero".
	if math.Trunc(details.Speed) == details.Speed {
		summary = fmt.Sprintf("%s Speed %3.0f knots. ", summary, math.Trunc(details.Speed))
	} else {
		summary = fmt.Sprintf("%s Speed %3.1f knots. ", summary, details.Speed)
	}

	// Read out interesting navigational statuses.
	switch details.NavigationalStatus {
	case 2:
		summary += "Not under command. "
	case 3:
		summary += "Restricted maneuverability. "
	case 4:
		summary += "Constrained by her draught. "
	case 6:
		summary += "Aground. "
	case 7:
		summary += "Engaged in fishing. "
	case 11:
		summary += "Power-driven vessel towing astern. "
	case 12:
		summary += "Power-driven vessel pushing ahead or towing alongside. "
	case 14:
		summary += "AIS-SART, MOB-AIS, or EPIRB-AIS. "
	}

	switch details.Sightings {
	case 0:
		summary += "This is the first sighting. "
	case 1:
		summary += "One previous sighting. "
	default:
		summary = fmt.Sprintf("%s %d previous sightings.", summary, details.Sightings)
	}

	err = beepspeak.Say(summary)
	if err != nil {
		return err
	}

	return nil
}
