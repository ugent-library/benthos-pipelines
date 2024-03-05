#!/bin/bash

file="/tmp/projects.json"
while read line
do
    a=$( echo -n "$line" | wc -c) # echo -n to prevent counting new line
    if [ "$a" -gt 10000 ]; then
      echo "$line"
    fi
done <"$file"

