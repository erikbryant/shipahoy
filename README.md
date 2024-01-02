# shipahoy

Alert when ships of interest are about to enter the part of the bay visible from our apartment.

Our apartment looks over part of the San Francisco bay. Alert if a ship of interest is about to enter that area. One side of that area is the part of the bay to the west of the Golden Gate bridge. The other part is the area to the east of Angel Island/Alcatraz.

## Build Status

![go fmt](https://github.com/erikbryant/shipahoy/actions/workflows/fmt.yml/badge.svg)
![go vet](https://github.com/erikbryant/shipahoy/actions/workflows/vet.yml/badge.svg)
![go test](https://github.com/erikbryant/shipahoy/actions/workflows/test.yml/badge.svg)

## Implementation Notes and Background

* [Encyclopedia of Marine Terms](https://www.wartsila.com/encyclopedia/term/standard-loading-conditions)
* [Most recent ports of call](https://www.vesselfinder.com/api/pro/portcalls/538007561?s)
* [AIS vessel types](https://help.marinetraffic.com/hc/en-us/articles/205579997-What-is-the-significance-of-the-AIS-Shiptype-number-)
* [AIS raw data format](https://www.navcen.uscg.gov/?pageName=AISMessagesA)
* [MMSI format - USCG](https://www.navcen.uscg.gov/index.php?pageName=mtMmsi)
* [MMSI format ID digits](https://en.wikipedia.org/wiki/Maritime_identification_digits)
* [MMSI format - wiki](https://en.wikipedia.org/wiki/Maritime_Mobile_Service_Identity)
* [MID registry](https://www.itu.int/en/ITU-R/terrestrial/fmd/Pages/mid.aspx)
* [Shine Micro DB](http://www.mmsispace.com/livedisplay.php?mmsiresult=636091798)
* [Shine Micro DB](http://www.mmsispace.com/common/getdetails_v3.php?mmsi=369083000)
* [Lat Lon calc](https://www.movable-type.co.uk/scripts/latlong.html)

## Installation

Follow the instructions in the `database` directory to install and configure the database.
