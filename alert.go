package main

import (
	"encoding/json"
	"fmt"
	"github.com/erikbryant/beepspeak"
	"github.com/erikbryant/ship_ahoy/database"
	"math"
	"strings"
	"time"
)

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

// alert prints a message and plays an alert tone.
func alert(details database.Ship) {
	fmt.Printf(
		"\nShip Ahoy!  %s  %s\n%+v\n\n",
		time.Now().Format("Mon Jan 2 15:04:05"),
		decodeMmsi(details.MMSI),
		prettify(details),
	)

	if strings.Contains(strings.ToLower(details.Type), "vehicle") {
		beepspeak.Play("meep.wav")
	} else if strings.Contains(strings.ToLower(details.Type), "pilot") {
		beepspeak.Play("pilot.mp3")
	} else {
		beepspeak.Play("ship_horn.mp3")
	}

	summary := fmt.Sprintf("Ship ahoy! %s. %s. Course %s degrees.", details.Name, details.Type, readableCourse(details.ShipCourse))

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

	beepspeak.Say(summary)
}
