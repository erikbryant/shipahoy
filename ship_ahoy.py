import argparse
import json
import time
import urllib.request

import geocoder
import mysql.connector
import pygame


# Most recent ports of call.
# https://www.vesselfinder.com/api/pro/portcalls/538007561?s
#
# AIS vessel types.
# https://help.marinetraffic.com/hc/en-us/articles/205579997-What-is-the-significance-of-the-AIS-Shiptype-number-
#
# MMSI format.
# https://www.navcen.uscg.gov/index.php?pageName=mtMmsi
# MID registry.
# https://www.itu.int/en/ITU-R/terrestrial/fmd/Pages/mid.aspx
# Shine Micro DB.
# http://www.mmsispace.com/livedisplay.php?mmsiresult=636091798
# http://www.mmsispace.com/common/getdetails_v3.php?mmsi=369083000

# Lat Lon calc
# https://www.movable-type.co.uk/scripts/latlong.html

# TODO:
# Put exception handling around url calls.
#  socket.gaierror: [Errno -2] Name or service not known
#  urllib.error.URLError: <urlopen error [Errno -2] Name or service not known>
# Put exception handling around geo calls.
# $ python3 ship_ahoy.py
#  Status code Unknown from http://ipinfo.io/json: ERROR - HTTPConnectionPool(host='ipinfo.io', port=80): Max retries exceeded with url: /json (Caused by NewConnectionError('<urllib3.connection.HTTPConnection object at 0x7f2df411f8d0>: Failed to establish a new connection: [Errno -2] Name or service not known',))
#  Traceback (most recent call last):
#    File "ship_ahoy.py", line 321, in <module>
#     main()
#    File "ship_ahoy.py", line 306, in main
#      url = "https://www.vesselfinder.com/vesselsonmap?bbox=%f%%2C%f%%2C%f%%2C%f" % bbox(nmiles=500)
#    File "ship_ahoy.py", line 86, in bbox
#      lat = geo.latlng[0]
#  TypeError: 'NoneType' object is not subscriptable
#

# Data Requests
#
# Ships in a region
# 22235849   -- lat * 600000
# 522        -- lon * 600000
# 683        -- Course *10
# 117        -- Speed * 10
# 280        -- ais
# 249857000  -- mmsi
# WAIKIKI    -- Ship name
# 0          -- unknown
#
# MMSI data of a given ship
# {
#  'imo': '9776755',
#  'name': 'WAIKIKI',
#  'type': 'Crude Oil Tanker',
#  't': '1531634521',
#  'sar': False,
#  'dest': 'MALTA',
#  'etastamp': 'Jul 19, 12:00',
#  'ship_speed': 11.7,
#  'ship_course': 68.3,
#  'timestamp': 'Jul 15, 2018 06:02 UTC',
#  '__id': '309251',
#  'pn': '9776755-249857000-6d17e98fedf50ed074675bd8f3396cd5',
#  'vo': 0,
#  'ff': False,
#  'direct_link': '/vessels/WAIKIKI-IMO-9776755-MMSI-249857000',
#  'draught': 8.8,
#  'year': '2017',
#  'gt': '61468',
#  'sizes': '250 x 44 m',
#  'dw': '112829'
# }
#
# Data query to ship_ahoy.ships
# {
#  'mmsi': '374518000',
#  'imo': '8687218',
#  'name': 'DONG HONG HANG 2',
#  'ais': 170,
#  'type': 'Bulk Carrier',
#  'sar': 0,
#  '__id': '0',
#  'vo': 0,
#  'ff': 0,
#  'direct_link': '/vessels/DONG-HONG-HANG-2-IMO-8687218-MMSI-374518000',
#  'draught': 5.0,
#  'year': 2011,
#  'gt': 8465,
#  'sizes': '137 x 20 m',
#  'length': 137,
#  'beam': 20,
#  'dw': 13685,
#  'unknown': 0
# }
#

# SQL statements
"""
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
    length int,
    beam int,
    dw int,
    unknown int
 );

 CREATE UNIQUE INDEX mmsi ON ships ( mmsi );

 DELETE FROM ships;

 ALTER TABLE ships MODIFY mmsi varchar(20);
 ALTER TABLE ships ADD ais int AFTER name;

 CREATE TABLE sightings (
    mmsi varchar(20),
    ship_course float,
    timestamp int,  # Unix datetime
    lat float,
    lon float,
    my_lat float,
    my_lon float
 );

 # Backup / Restore
 mysqldump -u ships -p db_name t1 > dump.sql
 mysql -u ships -p db_name < dump.sql

"""


ShipsSeen = {}
ExpireSecs = 30

# Invariants about a ship. As far as I know, these do not change
# over the life of the ship (as opposed to course or speed).
KEYS = [
    'mmsi',
    'imo',
    'name',
    'ais',
    'type',
    'sar',
    '__id',
    'vo',
    'ff',
    'direct_link',
    'draught',
    'year',
    'gt',
    'sizes',
    'length',
    'beam',
    'dw',
    'unknown',
]
KEYS_SIGHTING = [
    'mmsi',
    'ship_course',
    'timestamp',
    'lat',
    'lon',
    'my_lat',
    'my_lon',
]


# my_location() returns the geo coordinates of where this
# computer is as a (lat, lon) pair.
def my_location():
    geo = geocoder.ip('me')
    return (geo.latlng[0], geo.latlng[1])


# bbox() returns a bounding box of the circle with center of the
# current location and radius of 'nmiles' nautical miles.
# Returns (longA, latA, longB, latB) Where A is the bottom left
# corner and B is the upper right corner.
def bbox(nmiles, latlon):
    # Convert nautical miles to decimal degrees.
    delta = nmiles * 1.0 / 60.0

    lat = latlon[0]
    lon = latlon[1]

    bbox_latA = lat - delta
    bbox_longA = lon - delta
    bbox_latB = lat + delta
    bbox_longB = lon + delta

    return (bbox_longA, bbox_latA, bbox_longB, bbox_latB)


# alert() prints a message and plays an alert tone.
# Mute if we have already seen this ship.
def alert(mmsi='', ship='', details={}, url=''):
    print("Ship Ahoy!   %s\n%s\n%s\n" % (ship, details, url))

    # Play an alert tone.
    sound_file = "ship_horn.mp3"
    if 'vehicle' in details['type'].lower():
        # Vehicle carriers get their own sound. :-)
        sound_file = "meep.wav"

    # pygame has a 100% CPU usage bug, so only run it while the sound
    # is actually being played. https://github.com/pygame/pygame/issues/331
    pygame.mixer.init()
    pygame.mixer.music.load(sound_file)
    pygame.mixer.music.play()
    while pygame.mixer.music.get_busy():
        time.sleep(1)
    pygame.mixer.quit()


# persist_ship() saves a ship's data to the database.
def persist_ship(details):
    INSERT = "INSERT IGNORE INTO ships ( %s )" % ','.join(KEYS)
    INSERT += " VALUES ( %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s )"

    row = []
    for key in KEYS:
        value = ''
        if key in details:
            value = details[key]
        row.append(value)

    cnx = mysql.connector.connect(user='ships', password='shipspassword', database='ship_ahoy')
    cursor = cnx.cursor()
    cursor.execute(INSERT, row)
    cnx.commit()
    cursor.close()
    cnx.close()


# persist_sighting() saves a ship sighting to the database.
def persist_sighting(mmsi, ship_course, lat, lon):
    INSERT = "INSERT INTO sightings ( %s )" % ','.join(KEYS_SIGHTING)
    INSERT += " VALUES ( %s, %s, %s, %s, %s, %s, %s )"

    details = {}
    details['mmsi'] = mmsi
    details['ship_course'] = ship_course
    details['timestamp'] = int(time.time())
    details['lat'] = lat
    details['lon'] = lon
    latlon = my_location()
    details['my_lat'] = latlon[0]
    details['my_lon'] = latlon[1]

    row = []
    for key in KEYS_SIGHTING:
        value = ''
        if key in details:
            value = details[key]
        row.append(value)

    cnx = mysql.connector.connect(user='ships', password='shipspassword', database='ship_ahoy')
    cursor = cnx.cursor()
    cursor.execute(INSERT, row)
    cnx.commit()
    cursor.close()
    cnx.close()


# update() updates a given mmsi's row in the database.
def update(mmsi, col, value):
    UPDATE = "UPDATE ships SET %s = '%s' WHERE mmsi = '%s'" % (col, value, mmsi)

    cnx = mysql.connector.connect(user='ships', password='shipspassword', database='ship_ahoy')
    cursor = cnx.cursor()
    cursor.execute(UPDATE)
    cnx.commit()
    cursor.close()
    cnx.close()


# lookup() checks to see if a ship is already in the database.
def lookup(mmsi):
    SELECT = "SELECT * FROM ships WHERE mmsi = '%s'" % mmsi

    details = None

    cnx = mysql.connector.connect(user='ships', password='shipspassword', database='ship_ahoy')
    cursor = cnx.cursor()
    cursor.execute(SELECT)

    row = cursor.fetchone()
    if row is not None:
        details = {}
        for k in range(len(row)):
            details[KEYS[k]] = ''
            if row[k] is not None:
                details[KEYS[k]] = row[k]

    cnx.commit()
    cursor.close()
    cnx.close()

    return details


# lookup() checks to see if a ship is already in the database.
def sizes_by_mmsi():
    SELECT = "select mmsi from ships where length is NULL and sizes != 'N/A'"

    cnx = mysql.connector.connect(user='ships', password='shipspassword', database='ship_ahoy')
    cursor = cnx.cursor()
    cursor.execute(SELECT)

    mmsi = []
    row = cursor.fetchone()
    while row is not None:
        mmsi.append(row[0])
        row = cursor.fetchone()

    cnx.commit()
    cursor.close()
    cnx.close()

    return mmsi


# web_request() makes a web request for a given URL.
def web_request(url='', use_json=False):
    headers = {}
    headers['Host'] = 'www.vesselfinder.com'
    headers['Connection'] = 'keep-alive'
    headers['Accept'] = '*/*'
    headers['X-Requested-With'] = 'XMLHttpRequest'
    headers['User-Agent'] = 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.181 Safari/537.36'
    headers['Referer'] = 'https://www.vesselfinder.com/'
    headers['Accept-Language'] = 'en-US,en;q=0.9,fr;q=0.8'

    req = urllib.request.Request(url, headers=headers)
    content = ''
    try:
        response = urllib.request.urlopen(req)
        if use_json:
            content = json.load(response)
        else:
            content = response.read().decode('utf-8')
    except Exception as err:  # Let the caller retry if they care.
        print(err)

    return content


# visible_from_apt() returns a bool indicating whether the ship is visible
# from our apartment window.
def visible_from_apt(lat1, long1):
    # The bounding box for the area visible from our apartment.
    Visible_latA = 37.8052
    Visible_latB = 37.8613
    Visible_longA = -122.48
    Visible_longB = -122.4092

    # Note that A is the bottom left corner and B is the upper
    # right corner, so we need to work out C and D which are the
    # upper left and lower right corners.
    latC = Visible_latB
    latD = Visible_latA
    longC = Visible_longA
    longD = Visible_longB

    # Is the ship within the bounding box of our visible area?
    if lat1 < latD or lat1 > latC:
        return False
    if long1 < longC or long1 > longD:
        return False

    # Is the ship within our visible triangle (the bottom left
    # triangle of the bounding box)? It is if the latitude is
    # less than the latitude of the box's diagonal.
    # x == longitude, y == latitude
    m = (latC - latD) / (longC - longD)
    b = latC - m*longC
    y = m*long1 + b
    if lat1 > y:
        return False

    return True


# interesting() determines which ships in a given list are of interest.
# If any are, it signals an alert.
def interesting(ships):
    uninteresting_ais = [
        0,   # Unknown
        6,   # Passenger
        31,  # Tug
        36,  # Sailing vessel
        37,  # Pleasure craft
        52,  # Tug
        60,  # Passenger ship
        69,  # Passenger ship
    ]

    uninteresting_mmsi = [
        '367123640',  # Hawk
        '367389640',  # Oski
        '366990520',  # Del Norte
        '367566960',  # F/V Pioneer
        '367469070',  # Sunset Hornblower
        '338234637',  # HEWESCRAFT 220 OP
    ]

    throttle = 0

    for ship in ships.split('\n'):
        fields = ship.split('\t')

        # Skip the trailing line with its magic number.
        if len(fields) < 2:
            continue

        # https://api.vesselfinder.com/docs/response-ais.html
        lat = int(fields[0]) / 600000.0
        lon = int(fields[1]) / 600000.0
        ship_course = int(fields[2]) / 10.0
        speed = int(fields[3]) / 10.0  # SOG
        ais = int(fields[4])
        mmsi = fields[5]
        name = fields[6]
        unknown = int(fields[7])

        url = "https://www.vesselfinder.com/?mmsi=%s&zoom=13" % mmsi
        details = lookup(mmsi)
        if details is None:
            throttle += 1
            if throttle >= 50000:
                continue
            print("Found new ship: %s %s" % (mmsi, name))
            mmsi_url = "https://www.vesselfinder.com/clickinfo?mmsi=%s&rn=64229.85898456942&_=1524694015667" % mmsi
            details = web_request(url=mmsi_url, use_json=True)
            if not type(details) == type({}):
                print("Skipping... /", details, "/")
                continue
            length = 0
            beam = 0
            sizes = details['sizes'].split(' ')
            if len(sizes) == 4 and sizes[1] == 'x' and sizes[3] == 'm':
                length = int(sizes[0])
                beam = int(sizes[2])
            details['mmsi'] = mmsi
            details['ais'] = ais
            details['length'] = length
            details['beam'] = beam
            details['unknown'] = unknown
            persist_ship(details=details)
        else:
            if details['ais'] != ais:
                update(mmsi=mmsi, col='ais', value=ais)
            if details['unknown'] != unknown:
                update(mmsi=mmsi, col='unknown', value=unknown)

        # Only alert for ships that are moving.
        if speed < 4:
            continue

        # Skip 'uninteresting' ships.
        if ais in uninteresting_ais or mmsi in uninteresting_mmsi:
            continue

        # Only alert for ships visible from our apartment.
        if not visible_from_apt(lat, lon):
            continue

        now = time.time()
        last_seen = ShipsSeen.get(mmsi, 0)
        mute = now - last_seen <= 5*ExpireSecs
        ShipsSeen[mmsi] = now
        if mute:
            continue

        # We have passed all the tests! Save and alert.
        persist_sighting(mmsi, ship_course, lat, lon)
        alert(mmsi, ship, details, url)


def main():
    parser = argparse.ArgumentParser(description='Ship Ahoy')
    parser.add_argument('--snapshot', help='exit after one scan pass', action='store_true')
    args = parser.parse_args()

    boxes = []
    # (longA, latA, longB, latB) Where A is the bottom left
    # corner and B is the upper right corner.
    # boxes.append((-193, -16, -36, 71))  # North America
    # boxes.append((0, -16, 160, 62))  # Europe, SE Asia

    my_box = bbox(nmiles=100, latlon=my_location())
    step = 10
    for lat in range(-80, 80, step):
        boxes.append(my_box)
        for lon in range(-180, 180, step):
            boxes.append((lon, lat, lon+step, lat+step))

    while True:
        for box in boxes:
            print("Scanning", box)

            url = "https://www.vesselfinder.com/vesselsonmap?bbox=%f%%2C%f%%2C%f%%2C%f" % box
            url += "&zoom=12&mmsi=0&show_names=1&ref=35521.28976544603&pv=6"

            ships = web_request(url=url)
            interesting(ships=ships)

            if args.snapshot:
                return

            # Do not spam their web service.
            time.sleep(ExpireSecs)


main()
