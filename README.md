# mlfo-dist
Distributed version of MLFO based on ITU Y.3172 standard 

## Usage
### Compile protoc

`protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative momo/momo.proto`

### MLFO Server

`go run main.go -s=localhost:8001`

### MLFO Client

`go run main.go -h=localhost:8000 -i=/Users/ab/mlfo-dist/test/fedIntent.yaml`



python3 /Users/ab/mlfo-dist/underlay/factory/client.py --server=localhost:8080 --source=MNIST --model=keras --sink=robot.controller &

python3 /Users/ab/mlfo-dist/underlay/factory/server.py --server_address=localhost:8080 &
