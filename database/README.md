# Installation

This application stores ship sightings in an SQL database. You need to have one installed and listening on 127.0.0.1:3306. It has been developed and tested with [MySQL Community Server](https://dev.mysql.com/downloads/mysql/).

To easily access the `mysql` binary, add `/usr/local/mysql/bin` to the path.

## SQL statements

If you already have restore files (`ships.sql.gz` and `sightings.sql.gz`) you can skip these steps. The `restore.sh` script will do these for you.

```sql
CREATE DATABASE IF NOT EXISTS ship_ahoy;
```

```sql
CREATE USER IF NOT EXISTS 'ships'@'localhost' IDENTIFIED BY 'ships_password';
GRANT ALL ON ship_ahoy.* TO 'ships';
```

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

## Database Backup / Restore

```sh
mysqldump -u ships -p db_name t1 > dump.sql
mysql -u ships -p db_name < dump.sql
```
