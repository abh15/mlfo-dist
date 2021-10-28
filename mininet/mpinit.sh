#!/bin/bash
docker update --cpus 1 mn.cloud.0
docker exec -it -d mn.appserv.0 iperf -s
docker exec -it -d mn.app.1 iperf -c 10.0.0.2 -b 100m -t 2000