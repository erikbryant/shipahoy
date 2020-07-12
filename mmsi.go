package shipahoy

import (
	"fmt"
	"regexp"
	"strings"
)

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
