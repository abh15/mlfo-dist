# mlfo-dist
Distributed version of MLFO based on ITU Y.3172 standard 
## Requirements 
Containernet v3.1

go v1.14

[abh15/flower](https://github.com/abh15/flower)

## Usage
1. Clone this repo (intentdriven branch)

2. Copy the mininet/dyntopo.py to containernet/examples directory

3. `sudo python3 examples/dyntopo.py <fognum> <numedges per fog>` 

4. From another terminal : `curl -v -F file=@intent.yaml 'http://localhost:8000/receive'`

5. To start a new experiment , reset containers : `curl -X POST 'http://localhost:7000/cloudreset' -d numfog=2`


## Misc commands
### Compile protoc

`protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative momo/momo.proto`

### Build docker and push
`bash docker/build.sh`





### Port mapping
Flower: 

		5000 —> REST

	    6000 —> FLWRSERVER

MLFO: 

		8000 —> REST

	  	9000 —> MLFO SERVER



export GRPC_GO_LOG_VERBOSITY_LEVEL=99;export GRPC_GO_LOG_SEVERITY_LEVEL=info
