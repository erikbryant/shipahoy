#!/bin/bash

for t in ships sightings; do
    mysqldump -u root -p ship_ahoy ${t} > ${t}.sql
    gzip --best --force ${t}.sql
done
