package sbi

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

//CheckServer is dummy function to checks if fed server exists
func CheckServer() bool {
	//TODO: add search delay
	if _, err := os.Stat("/fedserv"); err == nil {
		//Server exists
		log.Println("Agg server already present")
		return true
	}
	return false
}

//RegisterServer is dummy function to launch fed server
func RegisterServer() {

	log.Println("Creating agg server...")
	f, _ := os.Create("/fedserv")
	defer f.Close()

	log.Println("...agg Server created")
}

//DeleteFile deletes fedserv/foghit file
func DeleteFile(path string) {
	err := os.Remove(path)
	if err != nil {
		log.Println(err.Error())
	}
}

//CheckFogHit checks if the fog has been hit i.e a intent has been incident
func CheckFogHit() bool {
	if _, err := os.Stat("/foghit"); err == nil {
		log.Println("Fog is hit")
		return true
	}
	return false
}

//RegisterFogHit writes a file to know if a intent has hit the fog
func RegisterFogHit() {
	f, _ := os.Create("/foghit")
	defer f.Close()

}

//StartFedCli sends a http POST reuest to fedcli docker to startfed clients
func StartFedCli(fedcliaddr string, numclipernode string, source string, model string, server string) {

	//e.g curl -X POST 'http://10.0.1.100:5000/cli' -d num=2 -d source=mnist -d model=simple -d server=localhost

	data := url.Values{
		"num":    {numclipernode},
		"source": {source},
		"model":  {model},
		"server": {server},
	}
	resp, err := http.PostForm("http://"+fedcliaddr+"/cli", data)

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

}

func StartFedServ(fedservaddr string) {
	data := url.Values{}
	resp, err := http.PostForm("http://"+fedservaddr+"/serv", data)

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
}

func CheckBandwidth(hostname string) bool {
	if strings.Contains(hostname, "smo") {
		return true
	}
	return false
}

func CheckCompute(hostname string) bool {
	if strings.Contains(hostname, "smo") {
		return true
	}
	return false
}
