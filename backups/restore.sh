#!/bin/bash -u

gunzip $1

SQL=$( echo $1 | sed "s/[.]gz$//1")

mysql -u ships -p shipahoy < ${SQL}
