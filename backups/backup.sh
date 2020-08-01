#!/bin/bash

for t in ships sightings; do
    mysqldump -u ships -p ship_ahoy --no-tablespaces ${t} > ${t}.sql
    gzip --best --force ${t}.sql
done
