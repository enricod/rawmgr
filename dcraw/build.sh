#!/bin/bash
rm -f dcraw
gcc -o dcraw -O4 dcraw.c -lm -DNODEPS

./dcraw ~/Pictures/raws/Canon_001.CR2 > OUT
