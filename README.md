# mlfo-dist
Distributed version of MLFO based on ITU Y.3172 standard 
## Requirements 
Containernet v3.1

<!-- go v1.14

[abh15/flower](https://github.com/abh15/flower) -->

## Start ONOS

`cd onos`

Terminal 1
`bazel run onos-local -- clean debug`

Terminal 2
`tools/test/bin/onos localhost`

onos> `app activate org.onosproject.openflow`

onos> `app activate org.onosproject.fwd`

`ctl^ + D`


## Usage
1. Clone this repo (evaluation branch). Start ONOS (see above).

2. Copy the mininet/cli50.py to containernet/examples directory

3. `cd containernet`

	`sudo python3 examples/cli50.py <num of cohorts>` 

	`sudo python3 examples/cli50.py 5`

4. In another terminal from the remote machine (Not the HHI system) we send intents to all MLFO nodes. Note the cohortdistr or intentdistr variables in the file before sending intent
`cd intents`

`python3 sendintent.py`

5. Open grafana dashboard at:
`http://10.66.2.142:3000/d/ouQy_xDnk/daszweite`

## Build docker and push to remote 
`sudo bash docker/build.sh`

## Reset federated learning server on cloud
`curl -X 'http://10.66.2.142:8999/cloudreset'`

## **************IMPORTANT**************
`sudo docker update --cpus 1 mn.fog.1`


## Misc commands
### Compile protoc

`protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative momo/momo.proto`

### Reset mininet
`sudo mn -c`

### Port mapping
Flower: 

		5000 —> REST

	    6000 —> FLWRINTERNAL

MLFO: 

		8000 —> REST

	  	9000 —> MLFO SERVER

### Send test intent
`curl -v -F file=@intent.yaml 'http://localhost:8000/receive'`

export GRPC_GO_LOG_VERBOSITY_LEVEL=99;export GRPC_GO_LOG_SEVERITY_LEVEL=info

sudo docker exec -it mn.fog.1 iperf -s

edge.1.2 iperf -c fog.1 -p 5001 -t 5


sudo docker update --cpus 2 mn.fog.1

curl -X POST 'http://10.0.1.100:5000/cli' -d num=2 -d source=mnist -d model=simple -d server=localhost

curl -X POST 'http://10.0.1.100:5000/cli' -d num=2 -d source=cifar -d model=mobilenet -d server=localhost


curl -X POST http://10.0.1.100:5000/serv


sudo docker exec -it mn.cloud.0 iperf -s

sudo docker exec -it mn.smo.1 /bin/bash

iperf -c 10.0.0.1 -p 5001 -t 5