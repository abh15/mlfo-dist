package sbi

import (
	"log"
	"os/exec"

	"github.com/abh15/mlfo-dist/parser"
)

//ResolveRequirements talks with the underlay to match requirements with available resource
func ResolveRequirements(s string, r parser.Requirements) string {
	var resourceID string
	_ = r
	switch s {
	case "source":
		resourceID = "robot.imagedata"
	case "model":
		resourceID = "keras.imagerec"
	case "sink":
		resourceID = "robot.armOptimiser"
	}
	return resourceID
}

//StartFedClients starts fed clients using flwr
func StartFedClients(localsrc string, localmodel string, localsink string, fedIP string, numClients int32) {

	cmd := exec.Command("python3 /Users/ab/mlfo-dist/underlay/factory/client.py" +
		"--server=" + fedIP + "--source" + localsrc +
		"--model" + localmodel + "--sink" + localsink)

	var i int32
	for i = 0; i < numClients; i++ {
		err := cmd.Run()
		if err != nil {
			log.Fatalf("cmd.Run() failed with %s\n", err)
		}
	}
}

//StartFedServer starts flwr server
func StartFedServer() string {
	cmd := exec.Command("python3 /Users/ab/mlfo-dist/underlay/factory/server.py --server_address=localhost:8080")
	err := cmd.Run()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}
	return "localhost:8080"
}
