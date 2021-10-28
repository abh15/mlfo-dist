package main

/*
	Step 1: Receive intent over http(:8000) OR over Mo-Mo(:9000)
	Step 2: Resolve local pipelines
	Step 3: Resolve intents
	Step 4: Send intents over Mo-Mo
	Step 5: Deploy local pipelines
*/

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"

	//"os/signal"
	"strconv"
	"strings"
	"sync"

	//"syscall"
	"time"

	pb "github.com/abh15/mlfo-dist/momo"
	"github.com/abh15/mlfo-dist/parser"
	"github.com/abh15/mlfo-dist/sbi"
	"google.golang.org/grpc"
)

const (
	//flaskport = ":5000"
	flowerport   = ":6000"
	intentport   = ":8000"
	mlfoport     = ":9000"
	centmlfoaddr = "10.0.0.1" + mlfoport
)

//Global Variable
var mutex = &sync.Mutex{}
var fedservoctet int = 100

func main() {
	// fedservoctet = 100
	// //If our server crashes delete files which indicate created fed servers
	// var gracefulStop = make(chan os.Signal)
	// signal.Notify(gracefulStop, syscall.SIGTERM)
	// signal.Notify(gracefulStop, syscall.SIGINT)
	// go func() {
	// 	sig := <-gracefulStop
	// 	fmt.Printf("caught manual interrupt from user: %+v", sig)
	// 	//sbi.ResetServer()
	// 	// fmt.Println("Wait for 2 second to finish server deletion")
	// 	// time.Sleep(2 * time.Second)
	// 	os.Exit(0)
	// }()

	// log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	// //Start grpc server for momo on port 9000 in different thread
	// wg := new(sync.WaitGroup)
	// wg.Add(1)
	// go func() {
	// 	StartServer(mlfoport)
	// 	wg.Done()
	// }()

	// //Start REST server for intent  on port 8000
	// log.Println("Started REST server on " + intentport)
	// http.HandleFunc("/receive", httpReceiveHandler)       // Handle the incoming intent
	// http.HandleFunc("/cloudreset", httpCloudResetHandler) // Handle the incoming reset msg
	// log.Fatal(http.ListenAndServe(intentport, nil))

	// wg.Wait()
	nodehostname, err := os.Hostname()
	if err != nil {
		log.Println(err.Error())
	}
	if strings.Contains(nodehostname, "fm") {
		time.Sleep(180 * time.Second)
		var fedintent parser.Intent
		var fedtarget parser.Target
		var fedtargetList []parser.Target
		var genIntents []parser.Intent

		fedtarget.ID = "cloud0-001"
		fedtarget.Operation = "aggregate.global"
		fedtarget.Operand = "model.federated"
		fedtarget.Constraints.Modelkind = "model"
		fedtarget.Constraints.Sourcekind = "source"
		fedtarget.Constraints.Avgalgo = "FedAvg"
		fedtarget.Constraints.Fracfit = "0.5"
		fedtarget.Constraints.Minfit = "1"
		fedtarget.Constraints.Minav = "1"
		fedtarget.Constraints.Numround = "20"
		fedtarget.Constraints.Sameserv = "no"

		fedtargetList = append(fedtargetList, fedtarget)
		fedintent.Targets = fedtargetList
		fedintent.IntentID = "fedintent-000"
		genIntents = append(genIntents, fedintent)

		_ = sendIntents(genIntents)

	} else {
		StartServer(mlfoport)
	}
}
func httpCloudResetHandler(w http.ResponseWriter, r *http.Request) {
	sbi.ResetServer()
}

//Step 1:
//receiveHandler handles the yaml file sent over REST
func httpReceiveHandler(w http.ResponseWriter, r *http.Request) {
	var outgoingIntents []parser.Intent
	yamlfile, _, err := r.FormFile("file")
	ipstart, _ := strconv.Atoi(r.FormValue("ipstart"))
	cohortsize, _ := strconv.Atoi(r.FormValue("cohortsize"))
	sameserver := r.FormValue("sameserver")
	avgalgo := r.FormValue("avgalgo")
	fracfit := r.FormValue("fracfit")
	minfit := r.FormValue("minfit")
	minav := r.FormValue("minav")
	numround := r.FormValue("numround")

	//nodehostname, err := os.Hostname()
	if err != nil {
		log.Println(err.Error())
	}

	if err != nil {
		//send error as HTTP response
		fmt.Fprintln(w, err)
		return
	}
	defer yamlfile.Close()

	//copy yamlfile to buffer
	buf := bytes.NewBuffer(nil)
	io.Copy(buf, yamlfile)

	//sent 200 ok as HTTP response
	fmt.Fprintf(w, "OK")
	log.Print("Intent received!!")

	//start intent processing in new thread
	wg2 := new(sync.WaitGroup)
	wg2.Add(1)
	go func() {
		start := time.Now()
		//Process intent further hier
		intent := parser.Parse(buf.Bytes())

		log.Printf("\nThe intent is %v\n", intent)

		//Step 2: Create pipeline configuration for FL clients based on intent
		pipelineconfig := createPipelineConfig(intent)

		//Step 3.1: Check of federated learning in required
		for _, target := range intent.Targets {
			if target.Constraints.Privacylevel == "high" {
				//Step 3.2: Generate Mo-Mo intent based on the user input intent.
				outgoingIntents = generateIntents(pipelineconfig, avgalgo, fracfit, minfit, minav, numround, sameserver)
			}
		}

		//Step 4:
		fedservIP := sendIntents(outgoingIntents)
		pipelineconfig["server"] = fedservIP

		//Step 5: Deploy FL client pipelines according to configuration
		var iplist []string
		for i := ipstart; i < cohortsize+ipstart; i++ {
			ipaddr := "10.0.1." + strconv.Itoa(i)
			iplist = append(iplist, ipaddr)
		}
		_ = deploylocal(pipelineconfig, iplist)

		elapsed := time.Since(start)
		log.Printf("HTTP Intent took %s", elapsed)

		wg2.Done()
	}()
}

//Step 2:
//createPipelineConfig returns map of pipeline <attributes, values> e.g src,model,sink for a target in the intent
func createPipelineConfig(in parser.Intent) map[string]string {
	var pipeline = make(map[string]string)

	for _, target := range in.Targets {
		//Logic: Welding accuracy can be improved by using 'handmnist' data set with 'complex' model and applying it to robot controller
		if target.Operation == "maximise" && target.Operand == "robots.welding.accuracy" {

			pipeline["source"] = "handmnist"
			pipeline["model"] = "complex"
			pipeline["sink"] = "robot.controller"
		}

		//Logic: Drilling accuracy can be improved by using 'fashionmnist' data set with 'complex' model and applying it to robot controller
		if target.Operation == "maximise" && target.Operand == "robots.drilling.accuracy" {

			pipeline["source"] = "fashionmnist"
			pipeline["model"] = "complex"
			pipeline["sink"] = "robot.controller"
		}
		//Logic: Drilling accuracy can be improved by using 'fashionmnist' data set with 'complex' model and applying it to robot controller
		if target.Operation == "maximise" && target.Operand == "robots.cutting.accuracy" {

			pipeline["source"] = "cifar"
			pipeline["model"] = "mobilenet"
			pipeline["sink"] = "robot.controller"
		}
		//Logic: For fed agg server create pipeline for fed agg
		if target.Operation == "aggregate.global" && target.Operand == "model.federated" {
			pipeline["source"] = target.Constraints.Sourcekind
			pipeline["model"] = target.Constraints.Modelkind
			pipeline["avgalgo"] = target.Constraints.Avgalgo
			pipeline["fracfit"] = target.Constraints.Fracfit
			pipeline["minfit"] = target.Constraints.Minfit
			pipeline["minav"] = target.Constraints.Minav
			pipeline["numround"] = target.Constraints.Numround
			pipeline["sameserver"] = target.Constraints.Sameserv
		}
	}

	return pipeline
}

//Step 3:
//generateIntents resolves intents for higher level MLFOs. If mlfo is enabled it will generate only one intent.
//If disabled it will generate intents equal to total number of FL clients
func generateIntents(pipelineconfig map[string]string, avgalgo string, fracfit string, minfit string, minav string, numround string, sameserver string) []parser.Intent {

	var fedintent parser.Intent
	var fedtarget parser.Target
	var fedtargetList []parser.Target
	var genIntents []parser.Intent

	fedtarget.ID = "cloud0-001"
	fedtarget.Operation = "aggregate.global"
	fedtarget.Operand = "model.federated"
	fedtarget.Constraints.Modelkind = pipelineconfig["model"]
	fedtarget.Constraints.Sourcekind = pipelineconfig["source"]
	fedtarget.Constraints.Avgalgo = avgalgo
	fedtarget.Constraints.Fracfit = fracfit
	fedtarget.Constraints.Minfit = minfit
	fedtarget.Constraints.Minav = minav
	fedtarget.Constraints.Numround = numround
	fedtarget.Constraints.Sameserv = sameserver

	fedtargetList = append(fedtargetList, fedtarget)
	fedintent.Targets = fedtargetList
	fedintent.IntentID = "fedintent-000"
	genIntents = append(genIntents, fedintent)

	return genIntents
}

//Step 4:
//sendIntents sends intents over Mo-Mo to all other MLFOs
func sendIntents(outIntents []parser.Intent) string {
	var waitgroup sync.WaitGroup
	var reply string
	log.Printf("gnerateIntents intent is ----------------------------->%+v\n", outIntents[0])
	//Convert intent struc to pb intent struc
	if len(outIntents) != 0 {
		waitgroup.Add(len(outIntents))
		for _, intent := range outIntents {
			go func(intent parser.Intent) {
				var pbIntent *pb.Intent
				intentBytes, err := json.Marshal(intent)
				if err != nil {
					log.Println(err.Error())
				}
				json.Unmarshal(intentBytes, &pbIntent)
				log.Printf("Sending this intent to-- %+v\n%+v\n", centmlfoaddr, pbIntent)
				reply = Send(centmlfoaddr, pbIntent) //handle reply
				log.Printf("Reply is %+v", reply)
				waitgroup.Done()
			}(intent)
		}
		waitgroup.Wait()
	}
	return reply //we assume all of the intents reply with same IP, for a given model-sourcekind
}

//Step 5:
//deploylocal deploys local pipelines in the local domain
func deploylocal(pipeline map[string]string, addrlist []string) string {
	var waitgroup sync.WaitGroup
	var aggserverip string = ""
	var fedservip string = ""
	nodehostname, err := os.Hostname()
	if err != nil {
		log.Println(err.Error())
	}
	if strings.Contains(nodehostname, "cloud") {
		mutex.Lock()
		// //Check if agg server of this type exists. If not then create one
		// if pipeline["source"] == "mnist" && pipeline["model"] == "simple" {
		// 	if sbi.CheckServer(pipeline["source"]+pipeline["model"]) == false {
		// 		sbi.RegisterServer(pipeline["source"] + pipeline["model"])
		// 		sbi.StartFedServ("10.0.0.101")
		// 	}
		// 	aggserverip = "10.0.0.101" + flowerport
		// }
		// //Check if agg server of this type exists. If not then create one
		// if pipeline["source"] == "cifar" && pipeline["model"] == "mobilenet" {
		// 	if sbi.CheckServer(pipeline["source"]+pipeline["model"]) == false {
		// 		sbi.RegisterServer(pipeline["source"] + pipeline["model"])
		// 		sbi.StartFedServ("10.0.0.102")
		// 	}
		// 	aggserverip = "10.0.0.102" + flowerport
		// }

		//If it does not use same server continue normal operation
		//else use the same server as previous, do not increment and do not send sbi msg. Return the old aggservip
		if strings.Contains(pipeline["sameserver"], "no") {
			//for reset
			if pipeline["sameserver"] == "nor" {
				fedservoctet = 100
			}
			fedservoctet = fedservoctet + 1
			fedservip = "10.0.0." + strconv.Itoa(fedservoctet)
			sbi.StartFedServ(fedservip, pipeline["avgalgo"], pipeline["fracfit"], pipeline["minfit"], pipeline["minav"], pipeline["numround"])
			aggserverip = fedservip + flowerport
		} else {
			fedservip = "10.0.0." + strconv.Itoa(fedservoctet)
			aggserverip = fedservip + flowerport
		}
		mutex.Unlock()
	}

	//deploy pipelines using campus mlfo by triggering ht FL clients
	if strings.Contains(nodehostname, "mo") {
		//nodenum := strings.Split(nodehostname, ".")[1]
		numclipercohort := len(addrlist)
		log.Printf("Addrlist ist %+v", addrlist)
		waitgroup.Add(numclipercohort)
		for i := 0; i < numclipercohort; i++ {
			go func(i int) {
				sbi.StartFedCli(addrlist[i], pipeline["source"], pipeline["model"], pipeline["server"], numclipercohort, i) //Might need to change to just i
				waitgroup.Done()
			}(i)
		}
		waitgroup.Wait()
	}
	return aggserverip
}

//Deploy is called when a Mo-Mo message is received on MLFO server
func (s *server) Deploy(ctx context.Context, rcvdIntent *pb.Intent) (*pb.Status, error) {
	// start2 := time.Now()

	// var intent parser.Intent
	// intentbytes, err := json.Marshal(rcvdIntent)
	// if err != nil {
	// 	log.Println(err.Error())
	// }
	// json.Unmarshal(intentbytes, &intent)
	// log.Printf("Received the following intent\n%v\n", intent)

	// /*
	// 	Step 1: Receive intent over http(:8000) OR over Mo-Mo(:9000)
	// 	Step 2: Deploy fed agg pipelines
	// 	Step 3: Send fed agg server reply back to client
	// */

	// pipelineconfig := createPipelineConfig(intent)

	// var dummy []string

	// status := deploylocal(pipelineconfig, dummy)

	// elapsed2 := time.Since(start2)
	// log.Printf("MoMo Intent took %s", elapsed2)
	// reply := status
	// return &pb.Status{Status: reply}, nil

	_ = rcvdIntent
	return &pb.Status{Status: "0.0.0.0"}, nil
}

//server is used to implement pb.UnimplementedOrchestrateServer
type server struct {
	pb.UnimplementedOrchestrateServer
}

//StartServer starts MLFO grpc server
func StartServer(port string) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Printf("failed to listen: %v", err)
	} else {
		log.Printf("\nStarted listening on %v\n", port)
	}
	s := grpc.NewServer()
	pb.RegisterOrchestrateServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Printf("failed to serve: %v", err)
	}
}

//Send sends msg over grpc
func Send(address string, message *pb.Intent) string {
	log.Printf("Connecting to %v ......", address)
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Printf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewOrchestrateClient(conn)

	// ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	// defer cancel()

	// r, err := c.Deploy(ctx, message)

	r, err := c.Deploy(context.Background(), message)
	if err != nil {
		log.Printf("could not receive: %v", err)
	}
	return r.GetStatus()

}
