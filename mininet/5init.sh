#!/bin/bash
docker update --cpus 2 mn.fed.1
docker update --cpus 2 mn.fed.2
docker update --cpus 2 mn.fed.3
docker update --cpus 2 mn.fed.4
docker update --cpus 2 mn.fed.5
docker exec -it mn.cloud.0 ping -c 2 10.0.0.101
docker exec -it mn.cloud.0 ping -c 2 10.0.0.102
docker exec -it mn.cloud.0 ping -c 2 10.0.0.103
docker exec -it mn.cloud.0 ping -c 2 10.0.0.104
docker exec -it mn.cloud.0 ping -c 2 10.0.0.105
docker exec -it -d mn.appserv.0 iperf -s
docker exec -it -d mn.app.1 iperf -c 10.0.0.2 -b 20m -t 2000s