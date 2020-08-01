# SQL statements

```sql
CREATE TABLE ships (
    mmsi varchar(20) not null,
    imo varchar(20) not null,
    name varchar(40) not null,
    ais int not null,
    type varchar(128) not null,
    sar boolean not null,
    direct_link varchar(128) not null,
    draught float not null,
    year int not null,
    gt int not null,
    length int not null,
    beam int not null,
    dw int not null,
    flag varchar(20) not null,
    invalidDimensions boolean not null,
    marineTrafficID int not null,
 );

 CREATE UNIQUE INDEX mmsi ON ships ( mmsi );

 DELETE FROM ships;

 ALTER TABLE ships MODIFY mmsi varchar(20);
 ALTER TABLE ships ADD ais int AFTER name;

 UPDATE ships SET length = 0 WHERE length IS NULL;
 ALTER TABLE ships MODIFY length int NOT NULL;

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
