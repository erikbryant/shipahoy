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
