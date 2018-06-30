import argparse
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

# TODO:
# Put exception handling around url calls.
#  socket.gaierror: [Errno -2] Name or service not known
#  urllib.error.URLError: <urlopen error [Errno -2] Name or service not known>
# Write the detected ships to a database so we can keep stats on them.

# SQL statements
"""
 CREATE TABLE ships (
    mmsi varchar(255),
    imo varchar(255),
    name varchar(255),
    type varchar(255),
    -- t unixtimestamp,
    sar boolean,
    -- dest varchar(255),
    -- etastamp 'Jun 21, 07:30',
    -- ship_speed float,
    -- ship_course float,
    -- timestamp 'Jun 27, 2018 17:48 UTC',
    __id varchar(255),
    -- pn varchar(255),
    vo int,
    ff boolean,
    direct_link varchar(255),
    draught float,
    year varchar(255),
    gt varchar(255),
    sizes varchar(255),
    dw varchar(255)
 );

 CREATE UNIQUE INDEX mmsi ON ships ( mmsi );

 DELETE FROM ships;
"""


ShipsSeen = {}
ExpireSecs = 60

# The bounding box for the area visible from our apartment.
Visible_latA  = 37.8052
Visible_latB  = 37.8613
Visible_longA = -122.48
Visible_longB = -122.4092


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


keys = [
    'mmsi',
    'imo',
    'name',
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
    'dw',
]


# persist() saves a ship sighting to the database.
def persist(mmsi='', ship='', details={}, url='', now=time.time()):
    INSERT = """
    INSERT IGNORE INTO ships (
       mmsi, imo, name, type, sar, __id, vo, ff, direct_link, draught, year, gt, sizes, dw
    )
    VALUES(
       %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s
    )
    """

    details['mmsi'] = mmsi

    row = []
    for key in keys:
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


# lookup() checks to see if a ship is already in the database.
def lookup(mmsi):
    SELECT = """
    SELECT * FROM ships WHERE mmsi = '%s'
    """ % mmsi

    details = None

    cnx = mysql.connector.connect(user='root', password='password', database='ship_ahoy')
    cursor = cnx.cursor()
    cursor.execute(SELECT)

    row = cursor.fetchone()
    if row is not None:
        details = {}
        for k in range(len(row)):
            details[keys[k]] = ''
            if row[k] is not None:
                details[keys[k]] = row[k]

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
def interesting(ships, headingMin=0, headingMax=360, visible=False):
    uninteresting_ais = [
        '0',   # Unknown
        '6',   # Passenger
        '31',  # Tug
        '36',  # Sailing vessel
        '37',  # Pleasure craft
        '52',  # Tug
        '60',  # Passenger ship
        '69',  # Passenger ship
    ]

    uninteresting_mmsis = [
        '367123640',  # Hawk
        '367389640',  # Oski
        '366990520',  # Del Norte
        '367566960',  # F/V Pioneer
    ]

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
        ais     = fields[4]
        mmsi    = fields[5]
        name    = fields[6]

        # Only look at ships that are moving.
        if speed < 4:
            continue

        # Only look at ships headed the direction of interest.
        if course < headingMin or course > headingMax:
            continue

        # Only look at ships visible from our apartment?
        if visible and not visible_from_apt(lat1, long1):
            continue

        # Skip 'uninteresting' ships.
        if ais in uninteresting_ais:
            continue
        if mmsi in uninteresting_mmsis:
            continue

        mmsi_url = "https://www.vesselfinder.com/clickinfo?mmsi=%s&rn=64229.85898456942&_=1524694015667" % mmsi
        details = lookup(mmsi)
        if details is None:
            details = web_request(url=mmsi_url, use_json=True)
        url = "https://www.vesselfinder.com/?mmsi=%s&zoom=13" % mmsi
        persist(mmsi, ship, details, url)
        alert(mmsi, ship, details, url)


def main():
    # (longA, latA, longB, latB)
    # A is the bottom left corner and B is the upper right corner.
    new_orleans = (-90.60, 29.54, -89.77, 30.08)
    greater_bay_area = (-123.5,37.2, -121.5, 38.4)
    # The bay visible from our apartment.
    visible = (Visible_longA, Visible_latA, Visible_longB, Visible_latB)
    # The bay to the west of our apartment's visible area.
    gate = (-122.56280995055705, 37.77840105911834, -122.47822461224163, 37.833635454273335)
    # The bay to the east of our apartment's visible area.
    outbound = (-122.45043142336208, 37.79005643280233, -122.36597402590112, 37.94129487900324)

    visible = False
    url = "https://www.vesselfinder.com/vesselsonmap?bbox=%f%%2C%f%%2C%f%%2C%f" % greater_bay_area
    url += "&zoom=12&mmsi=0&show_names=1&ref=35521.28976544603&pv=6"

    parser = argparse.ArgumentParser(description='Ship Ahoy')
    parser.add_argument('--snapshot', help='exit after one scan pass', action='store_true')
    args = parser.parse_args()

    # Initialize the sound system.
    pygame.mixer.init()

    while True:
        print("Scanning...")
        ships = web_request(url=url)
        interesting(ships=ships, visible=visible)

        if args.snapshot:
            break

        # Do not spam their web service.
        time.sleep(ExpireSecs)


main()
