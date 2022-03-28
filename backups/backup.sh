#!/bin/zsh -u

for table in ships sightings; do
  mysqldump --user=ships --password=ships_password ship_ahoy --no-tablespaces ${table} > ${table}.sql
  gzip --best --force ${table}.sql
done
