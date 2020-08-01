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
