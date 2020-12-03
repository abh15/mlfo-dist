# mlfo-dist
Distributed version of MLFO based on ITU Y.3172 standard 

## Usage
### Compile protoc

`protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative momo/momo.proto`

### MLFO Server

`go run main.go -s=localhost:8001`

### MLFO Client

`go run main.go -h=localhost:8000 -i=/Users/ab/mlfo-dist/test/fedIntent.yaml`

### Build docker
`docker build . -f docker/Dockerfile -t mlfo:latest`

### Load intent in MLFO
`curl -v -F file=@fedIntent.yaml http://localhost:8000/receive`

### Enable creation of resources on kube cluster
`kubectl create clusterrolebinding default-edit --clusterrole=edit --serviceaccount=default:default`


http://localhost:5000/start?server=localhost:8080&source=oran.du&model=MNIST&sink=robot.one


kubectl create clusterrolebinding default-edit --clusterrole=edit --serviceaccount=default:default

Flower: 5000 —> REST
	    6000 —> FLWRSERVER

MLFO: 8000 —> REST
	  9000 —> MLFO SERVER