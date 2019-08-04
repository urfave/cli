#!/usr/bin/env bash
set -e
echo "" > coverage.txt

for f in $(ls *.coverprofile); do
    cat $f >> coverage.txt
    rm $f
done