package sbi

import (
	"fmt"
	"net/http"
	"net/url"

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

	_ = numClients
	data := url.Values{
		"server": {fedIP},
		"source": {localsrc},
		"model":  {localmodel},
		"sink":   {localsink},
	}
	fmt.Printf("\n%+v\n", data)

	resp, err := http.PostForm("http://localhost:5000/start", data)

	if err != nil {
		panic(err)
	}
	resp.Body.Close()

	/* var i int32
	for i = 0; i < numClients; i++ {

	} */
}

//StartFedServer starts flwr server
func StartFedServer() string {
	//Kubectl api may be used

	// cmd := exec.Command("python3", "/Users/ab/mlfo-dist/underlay/factory/server.py", "--server_address", "localhost:8080")
	// err := cmd.Start()
	// if err != nil {
	// 	log.Fatalf("cmd.Run() failed with %s\n", err)
	// }
	return "localhost:8080"
}
