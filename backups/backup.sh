#!/bin/bash

for t in ships sightings; do
    mysqldump -u ships -p ship_ahoy ${t} > ${t}.sql
    gzip --best --force ${t}.sql
done
