# VesselFinder REST API Data Requests

Ships in a region response record from VesselFinder. The list is a binary structure.

```text
GET https://www.vesselfinder.com/api/pub/vesselsonmap?bbox=-73581179%2C22791346%2C-73451814%2C22982834&zoom=12&mmsi=0&ref=20639.80311629575&show_names=1
```

* Header bytes: `CECP`
* One packed record per ship:
  * 2 bytes: (unknown what these are)
  * 4 bytes: mmsi
  * 4 bytes: lat
  * 4 bytes: lon
  * 1 bytes: len(name)
  * n bytes: name

VesselFinder can also return detailed information about a given MMSI.

```text
GET https://www.vesselfinder.com/api/pub/click/367003250
{
  ".ns": 0,                 navigational status
                             -1 = not defined in the spec, but VesselFinder has used it
                              0 = under way using engine
                              1 = at anchor
                              2 = not under command
                              3 = restricted maneuverability
                              4 = constrained by her draught
                              5 = moored
                              6 = aground
                              7 = engaged in fishing
                              8 = under way sailing
                              9 = reserved for future amendment of navigational status for ships carrying DG, HS, or MP, or IMO hazard or pollutant category C, high speed craft (HSC)
                              10 = reserved for future amendment of navigational status for ships carrying dangerous goods (DG), harmful substances (HS) or marine pollutants (MP), or IMO hazard or pollutant category A, wing in ground (WIG)
                              11 = power-driven vessel towing astern (regional use)
                              12 = power-driven vessel pushing ahead or towing alongside (regional use)
                              13 = reserved for future use
                              14 = AIS-SART (active), MOB-AIS, EPIRB-AIS
                              15 = undefined = default (also used by AIS-SART, MOB-AIS and EPIRB-AIS under test)
  "a2": "us",               country of register (abbrv)
  "al": 19,                 length
  "aw": 8,                  width
  "country": "USA",         country of register
  "cu": 246.7,              course
  "dest": "FALSE RIVER",    destination
  "draught": 33,            draught
  "drm": 13.9
  "dw": 0,                  deadweight
  "etaTS": 1588620600,      ETA timestamp
  "gt": 0,                  gross tonnage
  "imo": 0,                 imo number
  "lc.": 0,                 load condition(???)
  "m9": 0,
  "name": "SARAH REED",     name
  "pic": "0-367003250-...", path to thumbnail image https://static.vesselfinder.net/ship-photo/0-367003250-cf317c76a96fd9b9f5ae4679c64bd065/0
  "r": 2,
  "sc.": 0,                 status: 0=underway, 1=at anchor, 2=at anchor(?)
  "sl": false,              newer position available via satellite?
  "slts": 174463992300      newer position timestamp
  "ss": 0.1,                speed (knots)
  "ts": 1587883051          timestamp of last position received
  "type": "Towing vessel",  AIS type
  "y": 0,                   year built
}
```
