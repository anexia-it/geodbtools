#!/bin/bash

set -eu

URLS=("https://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz" "https://geolite.maxmind.com/download/geoip/database/GeoLite2-Country.tar.gz" "https://geolite.maxmind.com/download/geoip/database/GeoLite2-ASN.tar.gz")

cd $(dirname $(readlink -e $0))

echo ">>> downloading files..."
for url in ${URLS[*]}
do
    filename=$(echo ${url##*/})
    directory_prefix=$(echo ${filename} | cut -d'.' -f1)_
    echo ">>> downloading ${filename} ..."
    rm -rf ./${filename} ./${directory_prefix}*
    wget -c -t0 -O ./${filename} ${url}
    echo ">>> extracting ${filename} ..."
    tar zxf ${filename}
    echo ">>> moving *.mmdb to $(dirname $(readlink -e $0)) ..."
    mv ./${directory_prefix}*/*.mmdb .
    echo ">>> cleaning up ..."
    rm -rf ./${filename} ./${directory_prefix}*
done
