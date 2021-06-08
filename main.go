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
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	pb "github.com/abh15/mlfo-dist/momo"
	"github.com/abh15/mlfo-dist/parser"
	"github.com/abh15/mlfo-dist/sbi"
	"google.golang.org/grpc"
)

const (
	intentport   = ":8000"
	momoport     = ":9000"
	flowerport   = ":5000"
	centmlfoaddr = "10.0.0.1"
)

//Global Variable
var mutex = &sync.Mutex{}

func main() {
	//Handle graceful exit
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	go func() {
		sig := <-gracefulStop
		fmt.Printf("caught sig: %+v", sig) //add if check hier
		if sbi.CheckServer() {
			sbi.DeleteFile("/fedserv")
		}
		if sbi.CheckFogHit() {
			sbi.DeleteFile("/foghit")
		}

		// fmt.Println("Wait for 2 second to finish server deletion")
		// time.Sleep(2 * time.Second)
		os.Exit(0)
	}()

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	//Start grpc server for momo on port 9000 in different thread
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		StartServer(momoport)
		wg.Done()
	}()

	//Start REST server for intent  on port 8000
	log.Println("Started REST server on " + intentport)
	http.HandleFunc("/receive", httpReceiveHandler) // Handle the incoming intent
	// http.HandleFunc("/cloudreset", httpCloudResetHandler) // Handle the incoming reset msg
	// http.HandleFunc("/fogreset", httpFogResetHandler)     // Handle the incoming internal reset msg
	log.Fatal(http.ListenAndServe(intentport, nil))

	wg.Wait()
}

//Step 1:
//receiveHandler handles the yaml file sent over REST
func httpReceiveHandler(w http.ResponseWriter, r *http.Request) {
	var outgoingIntents []parser.Intent
	yamlfile, _, err := r.FormFile("file")
	nodesperedge, err := strconv.Atoi(r.FormValue("nodesperedge"))
	clipernode, err := strconv.Atoi(r.FormValue("clipernode"))
	mlfostatus := r.FormValue("mlfostatus")
	flstatus := r.FormValue("flstatus")
	hierflstatus := r.FormValue("hierflstatus")
	nodehostname, err := os.Hostname()
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
				//Step 3.2: Generate Mo-Mo intents based on the user input intent.
				if mlfostatus == "enabled" {
					outgoingIntents = generateIntents(pipelineconfig, (nodesperedge * clipernode), true)
				} else {
					outgoingIntents = generateIntents(pipelineconfig, (nodesperedge * clipernode), false)
				}
			}
		}

		//Step 3.2: Check if local aggregation is required. Checks if the node is connected to satellite gateway. This should be done by checking compute and link BW for this edge.
		pipelineconfig["hierarchical"] = "false" //By default local agg is disabled
		if hierflstatus == "enabled" {
			if sbi.CheckBandwidth(nodehostname) && sbi.CheckCompute(nodehostname) {
				pipelineconfig["hierarchical"] = "true"
			}
		}

		//Step 4:
		fedservIP := sendIntents(outgoingIntents)
		pipelineconfig["server"] = fedservIP
		pipelineconfig["numclipernode"] = strconv.Itoa(clipernode)
		pipelineconfig["nodesperedge"] = strconv.Itoa(nodesperedge)

		if flstatus == "enabled" {
			//Step 5: Deploy FL client pipelines according to configuration
			_ = deploylocal(pipelineconfig)
		}

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
		//Logic: Welding accuracy can be improved by using 'mnist' data set with 'simple' model and applying it to robot controller
		if target.Operation == "maximise" && target.Operand == "robots.welding.accuracy" {
			pipeline["source"] = "mnist"
			pipeline["model"] = "simple"
			pipeline["sink"] = "robot.controller"
		}

		//Logic: Drilling accuracy can be improved by using 'cifar' data set with 'mobilenet' model and applying it to robot controller
		if target.Operation == "maximise" && target.Operand == "robots.drilling.accuracy" {

			pipeline["source"] = "cifar"
			pipeline["model"] = "mobilenet"
			pipeline["sink"] = "robot.controller"
		}
		//Logic: For fed agg server create pipeline for fed agg
		if target.Operation == "aggregate.global" && target.Operand == "model.federated" {
			pipeline["source"] = target.Constraints.Sourcekind
			pipeline["model"] = target.Constraints.Modelkind
		}
	}

	return pipeline
}

//Step 3:
//generateIntents resolves intents for higher level MLFOs. If mlfo is enabled it will generate only one intent.
//If disabled it will generate intents equal to total number of FL clients
func generateIntents(pipelineconfig map[string]string, totalflclients int, mlfoenabled bool) []parser.Intent {

	var fedintent parser.Intent
	var fedtarget parser.Target
	var fedtargetList []parser.Target
	var genIntents []parser.Intent

	fedtarget.ID = "cloud0-001"
	fedtarget.Operation = "aggregate.global"
	fedtarget.Operand = "model.federated"
	fedtarget.Constraints.Modelkind = pipelineconfig["model"]
	fedtarget.Constraints.Sourcekind = pipelineconfig["source"]
	fedtargetList = append(fedtargetList, fedtarget)
	fedintent.Targets = fedtargetList

	if mlfoenabled {
		fedintent.IntentID = "fedintent-000"
		genIntents = append(genIntents, fedintent)
	} else {
		for i := 0; i <= totalflclients; i++ {
			fedintent.IntentID = "fedintent-" + strconv.Itoa(i)
			genIntents = append(genIntents, fedintent)
		}
	}

	return genIntents
}

//Step 4:
//sendIntents sends intents over Mo-Mo to all other MLFOs
func sendIntents(outIntents []parser.Intent) string {
	var waitgroup sync.WaitGroup
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
				reply := Send(centmlfoaddr, pbIntent) //handle reply
				intent.FedServerIP = reply
				log.Printf("Reply is %+v", reply)
				waitgroup.Done()
			}(intent)
		}
		waitgroup.Wait()
	}
	return outIntents[0].FedServerIP //we assume all of the intents reply with same IP, for a given model-sourcekind
}

//Step 5:
//deploylocal deploys local pipelines in the local domain
func deploylocal(pipeline map[string]string) string {
	var waitgroup sync.WaitGroup
	var aggserverip string
	aggserverip = ""
	nodehostname, err := os.Hostname()
	if err != nil {
		log.Println(err.Error())
	}
	if strings.Contains(nodehostname, "cloud") {
		mutex.Lock()
		if sbi.CheckServer() == false {
			//if server does not exist create one
			sbi.RegisterServer()
			if pipeline["source"] == "mnist" && pipeline["model"] == "simple" {
				sbi.StartFedServ("10.0.0.101:5000")
				aggserverip = "10.0.0.101:5000"
			}
			if pipeline["source"] == "cifar" && pipeline["model"] == "mobilenet" {
				sbi.StartFedServ("10.0.0.102:5000")
				aggserverip = "10.0.0.102:5000"
			}
		}
		mutex.Unlock()
	}

	//deploy pipelines using campus mlfo by triggering ht FL clients
	if strings.Contains(nodehostname, "mo") {
		nodenum := strings.Split(nodehostname, ".")[1]
		//If hierarchical is true start local
		if pipeline["hierarchical"] == "true" {
			endpoint := "http://10.0." + nodenum + "100:5000/cli"
			sbi.StartFedCli(endpoint, "1", pipeline["source"], pipeline["model"], pipeline["server"])

		} else {
			numberofnodes, _ := strconv.Atoi(pipeline["nodesperedge"])
			waitgroup.Add(numberofnodes)
			for i := 1; i <= numberofnodes; i++ {
				go func(i int) {
					endpoint := "http://" + "10.0." + nodenum + strconv.Itoa(i+10) + ":5000/cli"
					sbi.StartFedCli(endpoint, pipeline["numclipernode"], pipeline["source"], pipeline["model"], pipeline["server"])
					waitgroup.Done()
				}(i)
			}
			waitgroup.Wait()
		}
	}

	return aggserverip
}

//Deploy is called when a Mo-Mo message is received on MLFO server
func (s *server) Deploy(ctx context.Context, rcvdIntent *pb.Intent) (*pb.Status, error) {
	start2 := time.Now()

	var intent parser.Intent
	intentbytes, err := json.Marshal(rcvdIntent)
	if err != nil {
		log.Println(err.Error())
	}
	json.Unmarshal(intentbytes, &intent)
	log.Printf("Received the following intent\n%v\n", intent)

	/*
		Step 1: Receive intent over http(:8000) OR over Mo-Mo(:9000)
		Step 2: Deploy fed agg pipelines
		Step 3: Send fed agg server reply back to client
	*/

	pipelineconfig := createPipelineConfig(intent)

	status := deploylocal(pipelineconfig)

	elapsed2 := time.Since(start2)
	log.Printf("MoMo Intent took %s", elapsed2)
	reply := status
	return &pb.Status{Status: reply}, nil
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
