# shipahoy

Alert when ships of interest are about to enter the part of the bay visible from our apartment.

Our apartment looks over part of the San Francisco bay. Alert if a ship of interest is about to enter that area. One side of that area is the part of the bay to the west of the Golden Gate bridge. The other part is the area to the east of Angel Island/Alcatraz.


# Implementation Notes and Background

Encyclopedia of Marine Terms https://www.wartsila.com/encyclopedia/term/standard-loading-conditions

Most recent ports of call.
https://www.vesselfinder.com/api/pro/portcalls/538007561?s

AIS vessel types.
https://help.marinetraffic.com/hc/en-us/articles/205579997-What-is-the-significance-of-the-AIS-Shiptype-number-

AIS raw data format.
https://www.navcen.uscg.gov/?pageName=AISMessagesA

MMSI format.
https://www.navcen.uscg.gov/index.php?pageName=mtMmsi
https://en.wikipedia.org/wiki/Maritime_identification_digits
https://en.wikipedia.org/wiki/Maritime_Mobile_Service_Identity

MID registry.
https://www.itu.int/en/ITU-R/terrestrial/fmd/Pages/mid.aspx

Shine Micro DB.
http://www.mmsispace.com/livedisplay.php?mmsiresult=636091798
http://www.mmsispace.com/common/getdetails_v3.php?mmsi=369083000

Lat Lon calc.
https://www.movable-type.co.uk/scripts/latlong.html

# VesselFinder REST API Data Requests

Ships in a region response record from VesselFinder. The list is a binary structure.

GET https://www.vesselfinder.com/api/pub/vesselsonmap?bbox=-73581179%2C22791346%2C-73451814%2C22982834&zoom=12&mmsi=0&ref=20639.80311629575&show_names=1

* Header bytes: `CECP`
* One packed record per ship:
  * 2 bytes: (unknown what these are)
  * 4 bytes: mmsi
  * 4 bytes: lat
  * 4 bytes: lon
  * 1 bytes: len(name)
  * n bytes: name

MMSI data of a given ship

```text
GET https://www.vesselfinder.com/api/pub/click/367003250
{
  ".ns": 0,                 navigational status
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
  "ss": 0.1,                speed (knots)
  "ts": 1587883051          timestamp (of position received?)
  "type": "Towing vessel",  AIS type
  "y": 0,                   year built
}
```

# SQL statements

```sql
CREATE TABLE ships (
    mmsi varchar(20),
    imo varchar(20),
    name varchar(128),
    ais int,
    type varchar(128),
    -- t unixtimestamp,
    sar boolean,
    -- dest varchar(255),
    -- etastamp 'Jun 21, 07:30',
    -- ship_speed float,
    -- ship_course float,
    -- timestamp 'Jun 27, 2018 17:48 UTC',
    __id varchar(20),
    -- pn varchar(255),  -- '0-227616590-808bd5b15abc2089364f4d3ccf1e13d6'
    vo int,
    ff boolean,
    direct_link varchar(128),
    draught float,
    year int,
    gt int,
    sizes varchar(50),
    length int not null,
    beam int not null,
    dw int,
    unknown int
 );

 CREATE UNIQUE INDEX mmsi ON ships ( mmsi );

 DELETE FROM ships;

 ALTER TABLE ships MODIFY mmsi varchar(20);
 ALTER TABLE ships ADD ais int AFTER name;

 UPDATE ships SET length = 0 WHERE length IS NULL;
 ALTER TABLE ships MODIFY length INT NOT NULL;

 ALTER TABLE ships DROP COLUMN unknown;

 CREATE TABLE sightings (
    mmsi varchar(20),
    ship_course float,
    timestamp int,  # Unix datetime
    lat float,
    lon float,
    my_lat float,
    my_lon float
 );
```

# Database Backup / Restore

```sh
mysqldump -u ships -p db_name t1 > dump.sql
mysql -u ships -p db_name < dump.sql
```

# Tidal Information

https://tidesandcurrents.noaa.gov/api/

Presidio tidal sensors https://tidesandcurrents.noaa.gov/stationhome.html?id=9414290

Bay Bridge Air Gap sensors https://tidesandcurrents.noaa.gov/map/index.html?id=9414304

Example queries

https://tidesandcurrents.noaa.gov/api/datagetter?date=latest&station=9414290&product=datums&datum=mllw&units=english&time_zone=lst_ldt&application=web_services&format=xml

```xml
<data>
<datum n="MHHW" v="11.817"/>
<datum n="MHW" v="11.208"/>
<datum n="DTL" v="8.897"/>
<datum n="MTL" v="9.160"/>
<datum n="MSL" v="9.097"/>
<datum n="MLW" v="7.113"/>
<datum n="MLLW" v="5.976"/>
<datum n="GT" v="5.841"/>
<datum n="MN" v="4.095"/>
<datum n="DHQ" v="0.609"/>
<datum n="DLQ" v="1.137"/>
<datum n="NAVD" v="5.917"/>
<datum n="LWI" v="2.781"/>
<datum n="HWI" v="24.721"/>
</data>
```

## Mean Lower Low Water for Presidio

https://tidesandcurrents.noaa.gov/api/datagetter?date=latest&station=9414290&product=water_level&datum=mllw&units=english&time_zone=lst_ldt&application=web_services&format=xml

```xml
<data>
<metadata id="9414290" name="San Francisco" lat="37.8063" lon="-122.4659"/>
<observations>
<wl t="2018-10-22 17:00" v="1.458" s="0.062" f="0,0,0,0" q="p"/>
</observations>
</data>
```

## Air gap for Bay Bridge D-E span

https://tidesandcurrents.noaa.gov/api/datagetter?date=latest&station=9414304&product=air_gap&datum=mllw&units=english&time_zone=lst_ldt&application=web_services&format=xml

```xml
<data>
<metadata id="9414304" name="San Francisco-Oakland Bay Bridge Air Gap" lat="37.8044" lon="-122.3728"/>
<observations>
<ag t="2018-10-24 16:48" v="204.400" s="0.121" f="1,0,0,0"/>
</observations>
</data>
```
