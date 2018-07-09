import argparse
import geocoder
import json
import time
import urllib.request
import pygame

import mysql.connector


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
# Write the detected ships to a database so we can keep stats on them.

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
    -- pn varchar(255),
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
 ALTER TABLE ships MODIFY imo varchar(20);
 ALTER TABLE ships MODIFY name varchar(128);
 ALTER TABLE ships ADD ais int AFTER name;
 ALTER TABLE ships MODIFY type varchar(128);
 ALTER TABLE ships MODIFY __id varchar(20);
 ALTER TABLE ships MODIFY direct_link varchar(128);
 ALTER TABLE ships MODIFY sizes varchar(50);
 ALTER TABLE ships ADD unknown int AFTER dw;
 ALTER TABLE ships MODIFY year int;
 ALTER TABLE ships MODIFY gt int;
 ALTER TABLE ships MODIFY dw int;

 TODO:
  backfill ais
  backfill length
  backfill beam
  backfill unknown

"""


ShipsSeen = {}
ExpireSecs = 90
# Invariants about a ship. As far as I know, these do not change
# over the life of the ship (as opposed to course, speed, etc.).
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


# bbox() returns a bounding box of the circle with center of the
# current location and radius of 'nmiles' nautical miles.
# Returns (longA, latA, longB, latB) Where A is the bottom left
# corner and B is the upper right corner.
def bbox(nmiles=15):
    # Convert nautical miles to decimal degrees.
    delta = nmiles * 1.0 / 60.0

    geo = geocoder.ip('me')
    lat = geo.latlng[0]
    lng = geo.latlng[1]

    bbox_latA = lat - delta
    bbox_longA = lng - delta
    bbox_latB = lat + delta
    bbox_longB = lng + delta

    return (bbox_longA, bbox_latA, bbox_longB, bbox_latB)


# alert() prints a message and plays an alert tone.
# Mute if we have already seen this ship.
def alert(mmsi='', ship='', details={}, url=''):
    now = time.time()
    last_seen = ShipsSeen.get(mmsi, 0)
    mute = now - last_seen <= ExpireSecs
    expire = now + 2*ExpireSecs
    ShipsSeen[mmsi] = expire
    if mute:
        return

    print("Ship Ahoy!   %s\n%s\n%s\n" % (ship, details, url))

    # Play an alert tone.
    if 'vehicle' in details['type'].lower():
        # Vehicle carriers get their own sound. :-)
        pygame.mixer.music.load("meep.wav")
    else:
        # Play the generic ship horn, unless something
        # is already playing.
        if pygame.mixer.music.get_busy():
            return
        pygame.mixer.music.load("ship_horn.mp3")
    pygame.mixer.music.play()


# persist() saves a ship sighting to the database.
def persist(details):
    INSERT = "INSERT IGNORE INTO ships ( %s )" % ','.join(KEYS)
    INSERT += " VALUES ( %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s )"

    row = []
    for key in KEYS:
        value = ''
        if key in details:
            value = details[key]
        row.append(value)

    cnx = mysql.connector.connect(user='root', password='password', database='ship_ahoy')
    cursor = cnx.cursor()
    cursor.execute(INSERT, row)
    cnx.commit()
    cursor.close()
    cnx.close()


# update() updates a given mmsi's row in the database.
def update(mmsi, col, value):
    UPDATE = "UPDATE ships SET %s = '%s' WHERE mmsi = '%s'" % (col, value, mmsi)

    cnx = mysql.connector.connect(user='root', password='password', database='ship_ahoy')
    cursor = cnx.cursor()
    cursor.execute(UPDATE)
    cnx.commit()
    cursor.close()
    cnx.close()


# lookup() checks to see if a ship is already in the database.
def lookup(mmsi):
    SELECT = "SELECT * FROM ships WHERE mmsi = '%s'" % mmsi

    details = None

    cnx = mysql.connector.connect(user='root', password='password', database='ship_ahoy')
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


# web_request() makes a web request.
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
    except urllib.error.URLError as err:
        print(err)

    return content


# visible_from_apt() returns a bool indicating whether the ship is visible
# from our apartment window.
def visible_from_apt(lat1, long1):
    # The bounding box for the area visible from our apartment.
    Visible_latA  = 37.8052
    Visible_latB  = 37.8613
    Visible_longA = -122.48
    Visible_longB = -122.4092

    # Note that A is the bottom left corner and B is the upper
    # right corner, so we need to work out C and D which are the
    # upper left and lower right corners.
    latC  = Visible_latB
    latD  = Visible_latA
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
    ]

    throttle = 0

    for ship in ships.split('\n'):
        fields = ship.split('\t')

        # Skip the trailing line with its magic number.
        if len(fields) < 2:
            continue

        # https://api.vesselfinder.com/docs/response-ais.html
        lat1    = int(fields[0]) / 600000.0
        long1   = int(fields[1]) / 600000.0
        course  = int(fields[2]) / 10.0
        speed   = int(fields[3]) / 10.0  # SOG
        ais     = int(fields[4])
        mmsi    = fields[5]
        name    = fields[6]
        unknown = int(fields[7])

        url = "https://www.vesselfinder.com/?mmsi=%s&zoom=13" % mmsi
        details = lookup(mmsi)
        if details is None:
            throttle += 1
            if throttle >= 20:
                continue
            print("Found new ship: %s %s" % (mmsi, name))
            mmsi_url = "https://www.vesselfinder.com/clickinfo?mmsi=%s&rn=64229.85898456942&_=1524694015667" % mmsi
            details = web_request(url=mmsi_url, use_json=True)
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
            persist(details=details)
        else:
            if details['ais'] != ais:
                print(mmsi, 'ais', ais, details['ais'])
                update(mmsi=mmsi, col='ais', value=ais)
            if details['unknown'] != unknown:
                print(mmsi, 'unknown', unknown, details['unknown'])
                update(mmsi=mmsi, col='unknown', value=unknown)
            sizes = details['sizes'].split(' ')
            if len(sizes) == 4 and sizes[1] == 'x' and sizes[3] == 'm':
                length = int(sizes[0])
                beam = int(sizes[2])
                if details['length'] != length:
                    print(mmsi, 'length', length, details['length'])
                    update(mmsi=mmsi, col='length', value=length)
                if details['beam'] != beam:
                    print(mmsi, 'beam', beam, details['beam'])
                    update(mmsi=mmsi, col='beam', value=beam)

        # Only alert for ships that are moving.
        if speed < 4:
            continue

        # Skip 'uninteresting' ships.
        if ais in uninteresting_ais or mmsi in uninteresting_mmsi:
            continue

        # Only alert for ships visible from our apartment.
        if not visible_from_apt(lat1, long1):
            continue

        alert(mmsi, ship, details, url)


def main():
    parser = argparse.ArgumentParser(description='Ship Ahoy')
    parser.add_argument('--snapshot', help='exit after one scan pass', action='store_true')
    args = parser.parse_args()

    # Initialize the sound system.
    pygame.mixer.init()

    box = (-193, -16, -36, 71)  # North America
    box = bbox(nmiles=150)
    url = "https://www.vesselfinder.com/vesselsonmap?bbox=%f%%2C%f%%2C%f%%2C%f" % box
    url += "&zoom=12&mmsi=0&show_names=1&ref=35521.28976544603&pv=6"

    while True:
        print("Scanning...")
        ships = web_request(url=url)
        interesting(ships=ships)

        if args.snapshot:
            break

        # Do not spam their web service.
        time.sleep(ExpireSecs)


main()
