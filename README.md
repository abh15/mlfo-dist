# mlfo-dist
Distributed version of MLFO based on ITU Y.3172 standard 

## Requirements 

golang v1.14

## Build docker and push to remote 
`sudo bash docker/build.sh`


## Misc commands
### Compile protoc
Everytime the intent structure is changed we need to regenerate the proto files	

`protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative momo/momo.proto`

### Port mapping
Flower: 

		5000 —> REST

	    6000 —> FLWRINTERNAL

MLFO: 

		8000 —> REST

	  	9000 —> MLFO SERVER

### Send test intent
`curl -v -F file=@intent.yaml 'http://localhost:8000/receive'`

### Set verbosity level
export GRPC_GO_LOG_VERBOSITY_LEVEL=99;export GRPC_GO_LOG_SEVERITY_LEVEL=info
