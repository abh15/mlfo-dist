# mlfo-dist
Distributed version of MLFO based on ITU Y.3172 standard 
## Requirements 
Kubernetes v1.19.4

Docker v19.03.13

Minikube v1.15.1

kubectl v1.19.4

helm v3.4.1

go v1.14

[abh15/flower](https://github.com/abh15/flower)

## Usage
1. Clone this repo

2. `kubectl create clusterrolebinding default-edit --clusterrole=edit --serviceaccount=default:default`

3. `helm install edge helmchart/fedml` 

4. `kubectl port-forward <mlfo-0 pod name> 8000:8000`

5. `kubectl logs -f <pods to monitor>`

6. `cd intents; curl -v -F file=@fedIntent.yaml 'http://localhost:8000/receive'`

7. TBD: add script to remove all fedserv pod/svc/deployments




## Misc commands

### Delete helm chart

`helm delete edge`

### Compile protoc

`protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative momo/momo.proto`

### Build docker and push
`bash docker/build.sh`

### http test commands
`curl -X POST -F 'maxcli=2' http://fedserv:5000/launchserv`

`curl -X POST -F 'server=fedserv-6a3f7aee6b:6000' -F 'source=oran' -F 'model=MNIST' -F 'sink=robot' -F 'num=2' http://localhost:5000/startcli`


### Port mapping
Flower: 

		5000 —> REST

	    6000 —> FLWRSERVER

MLFO: 

		8000 —> REST

	  	9000 —> MLFO SERVER

