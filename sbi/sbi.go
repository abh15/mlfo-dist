package sbi

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	flaskport = ":5000"
)

//CheckServer is dummy function to checks if fed server exists
func CheckServer(serverid string) bool {
	//TODO: add search delay
	if _, err := os.Stat("/" + serverid); err == nil {
		//Server exists
		log.Printf("Agg server of type %v already present", serverid)
		return true
	}
	return false
}

//RegisterServer is dummy function to launch fed server
func RegisterServer(serverid string) {

	log.Printf("Creating agg server %v..", serverid)
	f, _ := os.Create("/" + serverid)
	defer f.Close()

	log.Println("...agg Server created")
}

//DeleteFile deletes fedserv/foghit file
func ResetServer() {
	err := os.Remove("/mnistsimple")
	if err != nil {
		log.Println(err.Error())
	}
	err = os.Remove("/cifarmobilenet")
	if err != nil {
		log.Println(err.Error())
	}

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
	resp, err := http.PostForm("http://"+fedcliaddr+flaskport+"/cli", data)

	log.Printf("Sent :\n%v\n", data)

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

}

func StartFedServ(fedservaddr string) {
	data := url.Values{}
	resp, err := http.PostForm("http://"+fedservaddr+flaskport+"/serv", data)

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
