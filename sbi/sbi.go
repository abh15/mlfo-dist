package sbi

import (
	"log"
	"os"
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

//LaunchServer is dummy function to launch fed server
func LaunchServer() {

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

// //CreateFedMLCient simulates creating local FL client pipeline
// func CreateFedMLCient(delay string) {
// 	t, err := strconv.Atoi(delay)
// 	if err != nil {
// 		log.Println(err.Error())
// 	}
// 	time.Sleep(time.Duration(t) * time.Second)
// }
