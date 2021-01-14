#!/bin/bash
for containerId in $(docker ps -q)
    do
        sudo docker exec $containerId /bin/bash -c "pkill /app/mlfo" &

    done