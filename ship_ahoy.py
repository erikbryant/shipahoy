import time
import urllib.request
import pygame


# Most recent ports of call.
# https://www.vesselfinder.com/api/pro/portcalls/367673390?s


# alert() prints a message and plays an alert tone.
def alert(message):
    print(message)

    # Play an alert tone.
    if pygame.mixer.music.get_busy():
        return
    pygame.mixer.music.load("train_horn.mp3")
    pygame.mixer.music.play()


# ships_by_region() returns all of the ships in a given region.
def ships_by_region(url):
    headers = {}
    headers['Host'] = 'www.vesselfinder.com'
    headers['Connection'] = 'keep-alive'
    headers['Accept'] = '*/*'
    headers['X-Requested-With'] = 'XMLHttpRequest'
    headers['User-Agent'] = 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.181 Safari/537.36'
    headers['Referer'] = 'https://www.vesselfinder.com/'
    # headers['Accept-Encoding'] = 'gzip'
    headers['Accept-Language'] = 'en-US,en;q=0.9,fr;q=0.8'

    req = urllib.request.Request(url, headers=headers)
    content_raw = b''
    with urllib.request.urlopen(req) as response:
        content_raw += response.read()

    return content_raw.decode('utf-8')


# mmsi_detail() returns the details for a given MMSI.
def mmsi_details(mmsi):
    headers = {}
    headers['Host'] = 'www.vesselfinder.com'
    headers['Connection'] = 'keep-alive'
    headers['Accept'] = '*/*'
    headers['X-Requested-With'] = 'XMLHttpRequest'
    headers['User-Agent'] = 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_4) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/65.0.3325.181 Safari/537.36'
    headers['Referer'] = 'https://www.vesselfinder.com/'
    # headers['Accept-Encoding'] = 'gzip'
    headers['Accept-Language'] = 'en-US,en;q=0.9,fr;q=0.8'

    url = "https://www.vesselfinder.com/clickinfo?mmsi=%s&rn=64229.85898456942&_=1524694015667" % mmsi
    req = urllib.request.Request(url, headers=headers)
    content_raw = b''
    with urllib.request.urlopen(req) as response:
        content_raw += response.read()

    return content_raw.decode('utf-8')


# interesting() determines which ships in a given list are of interest. If any are,
# it signals an alert.
def interesting(ships, headingMin=0, headingMax=359):
    uninteresting_ais = [
        '0',   # Unknown
        '6',   # Passenger
        '36',  # Sailing vessel
        '37',  # Pleasure craft
        '52',  # Tug
        '60',  # Passenger ship
        '69',  # Passenger ship
    ]

    uninteresting_mmsis = [
        '367123640',  # Hawk
    ]

    for ship in ships.split('\n'):
        fields = ship.split('\t')

        # Skip the trailing line with its magic number.
        if len(fields) < 2:
            continue

        heading = int(fields[2]) / 10.0
        speed   = int(fields[3]) / 10.0
        ais     = fields[4]
        mmsi    = fields[5]
        name    = fields[6]

        # Only look at ships that are moving relatively fast.
        if speed < 7:
            # print("Skipping by speed: %s" % ship)
            continue

        # Only look at inbound ships.
        if heading < headingMin or heading > headingMax:
            # print("Skipping by heading: %s" % ship)
            continue

        # Skip 'uninteresting' ships.
        if ais in uninteresting_ais:
            # print("Skipping uninteresting AIS: %s" % ship)
            continue
        if mmsi in uninteresting_mmsis:
            # print("Skipping uninteresting MMSI: %s" % ship)
            continue

        details = mmsi_details(mmsi)
        alert("Ship ahoy!  %s\n%s" % (ship, details))


def main():
    pygame.mixer.init()

    # The part of the bay visible from our apartment.
    visible = "https://www.vesselfinder.com/vesselsonmap?bbox=-122.50628910495868%2C37.7868980951191%2C-122.37668476535909%2C37.87150340418255&zoom=13&mmsi=0&show_names=1&ref=10800.429711736251&pv=6"

    # The part of the bay to the west of our visible area.
    gate = "https://www.vesselfinder.com/vesselsonmap?bbox=-122.56280995055705%2C37.77840105911834%2C-122.47822461224163%2C37.833635454273335&zoom=13.615634506295129&mmsi=0&show_names=1&ref=53552.672128591235&pv=6"

    # The part of the bay to the east of our visible area.
    outbound = "https://www.vesselfinder.com/vesselsonmap?bbox=-122.45043142336208%2C37.79005643280233%2C-122.36597402590112%2C37.94129487900324&zoom=12&mmsi=0&show_names=1&ref=35521.28976544603&pv=6"

    while True:
        # Inbound ships about to enter our visible area.
        ships = ships_by_region(gate)
        interesting(ships=ships, headingMin=10, headingMax=170)

        # Outbound ships about to enter our visible area.
        ships = ships_by_region(outbound)
        interesting(ships=ships, headingMin=225, headingMax=315)

        # Do not spam their interface.
        time.sleep(60)


main()
