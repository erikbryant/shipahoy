# Marine Traffic REST API

https://www.marinetraffic.com/en/ais-api-services/documentation/

## Ships in Region

https://www.marinetraffic.com/getData/get_data_json_4/z:14/X:1309/Y:3165/station:0

* z - The zoom level
* X - ??
* Y - ??

### Response

The API returns an array of ships.

```json
{
  "type": 1,
  "data":{
    "rows":[
         {}
    ],
    "areaShips": 17
  }
}
```

The following information is provided for each ship.

```txt
{
   "LAT": "37.80998",          Latitude
   "LON": "-122.4215",         Longitude
   "SPEED": "0",               Knots (x10)
   "COURSE": "7",              Degrees
   "HEADING": "511",           Degrees, 511: no data
   "DESTINATION": "CLASS B",   Destination as entered into AIS
   "FLAG": "US",               Country (abbrv)
   "LENGTH": "12",             Meters
   "WIDTH": "4",               Meters
   "ROT": "0",                 Rate of Turn degrees/minute
   "SHIPNAME": "HYPERFISH",    Ship name
   "SHIPTYPE": "2",            Ship type https://help.marinetraffic.com/hc/en-us/articles/205579997-What-is-the-significance-of-the-AIS-SHIPTYPE-number-
   "SHIP_ID": "4143995",       MarineTraffic internal ID
   "INVALID_DIMENSIONS": "0"   0: False, 1: True
   "ELAPSED": "2",             ???
   "L_FORE": "3",              ???
   "W_LEFT": "2",              ???
}
```
