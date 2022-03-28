#!/bin/zsh -u

# TODO: Conditionally create database, user

echo "Creating ship_ahoy database and 'ships' user. Enter root password..."
mysql --user=root --password --execute="CREATE DATABASE IF NOT EXISTS ship_ahoy; CREATE USER IF NOT EXISTS 'ships'@'localhost' IDENTIFIED BY 'ships_password'; GRANT ALL ON ship_ahoy.* TO 'ships'@'localhost';"

for table in *.sql.gz; do
  gzcat ${table} | mysql --user=ships --password=ships_password ship_ahoy
done
