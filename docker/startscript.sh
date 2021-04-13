#!/bin/bash
for containerId in $(docker ps -q)
    do
        sudo docker exec -d $containerId /app/mlfo $1 $2 $3 &
    done
