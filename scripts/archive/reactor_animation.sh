#!/usr/bin/env bash

FILE=$1

while read line
do
    printf "%s\n" "$line"
    sleep 0.03
done < "$FILE"
