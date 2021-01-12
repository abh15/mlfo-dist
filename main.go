package main

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
	"strconv"
	"strings"
	"sync"
	"time"

	pb "github.com/abh15/mlfo-dist/momo"
	"github.com/abh15/mlfo-dist/parser"
	"google.golang.org/grpc"
)

const (
	intentport = ":8000"
	momoport   = ":9000"
)

//Global Variable
//newIntents has structure <intent_target_hostname, intent>. It consists of intents for peers as well as higher nodes
var newIntents = make(map[string]parser.IntentNoExp)
var aggservpresent = false
var globalaggservpresent = false

func main() {

	//Start grpc server for momo on port 9000 in different thread
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		StartServer(momoport)
		wg.Done()
	}()

	//Start REST server for intent  on port 8000
	log.Println("Started REST server on " + intentport)
	http.HandleFunc("/receive", httpReceiveHandler) // Handle the incoming file
	log.Fatal(http.ListenAndServe(intentport, nil))

	wg.Wait()
}

//receiveHandler handles the yaml file sent over REST
func httpReceiveHandler(w http.ResponseWriter, r *http.Request) {

	yamlfile, _, err := r.FormFile("file")
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

		/*
			Step 1: Receive intent over http(:8000) OR over Mo-Mo(:9000)
			Step 2: Resolve local pipelines
			Step 3: Resolve foreign intents (peer + upper nodes)
			Step 4: Send intents over Mo-Mo
			Step 5: Deploy local pipelines
		*/

		pipelines := resolvePipeline(intent)
		resolvePeerIntents(intent)
		resolveUpperIntents(intent, pipelines)
		sendIntents()
		deploylocal(pipelines)

		elapsed := time.Since(start)
		log.Printf("HTTP Intent took %s", elapsed)

		wg2.Done()
	}()
}

//resolvePipeline returns map of pipeline <attributes, values> e.g src,model,sink
func resolvePipeline(in parser.Intent) []map[string]string {
	var pipelines []map[string]string
	var pipeline = make(map[string]string)

	for _, target := range in.Targets {
		if target.Operation == "maximise" && target.Operand == "robots.welding.accuracy" {
			//model selection logic--
			//get robots list from sbi
			//sensors, controller = sbi.GetAssetMetadata(target.Operand)
			//Source data selectin logic---
			pipeline["source"] = "asset1.image.rgb"
			pipeline["model"] = "classifier.randomForest"
			pipeline["sink"] = "robot.arm.controller"
			pipelines = append(pipelines, pipeline)
		}
		if target.Operation == "aggregate" && target.Operand == "model.federated" {
			pipeline["source"] = "edgesrc"
			pipeline["model"] = "federated"
			pipeline["sink"] = "edgetarg"
			pipelines = append(pipelines, pipeline)

		}
	}

	return pipelines
}

//resolvePeerIntents creates intents for peer MLFOs
func resolvePeerIntents(in parser.Intent) {

	//create intent for peer MLFOs
	for _, target := range in.Targets {
		if target.ID == "factory.all" {
			//construct new intents for all factories
			for i := 1; i < int(in.Exp.Eperfog)+1; i++ {
				for j := 2; j < int(in.Exp.Numfog)+2; j++ {
					var edintent parser.IntentNoExp
					var edtargetList []parser.Target
					var edtarget parser.Target
					edintent.IntentID = "edgeintent-edge" + strconv.Itoa(i) + strconv.Itoa(j)
					edtarget.ID = "edge" + strconv.Itoa(i) + "." + strconv.Itoa(j)
					edtarget.Operation = target.Operation
					edtarget.Operand = target.Operand
					edtarget.Constraints = target.Constraints
					edtargetList = append(edtargetList, edtarget)
					edintent.Targets = edtargetList
					newIntents["10.0."+strconv.Itoa(i)+"."+strconv.Itoa(j)] = edintent
				}
			}
			//exception for host on which the intent is received
			delete(newIntents, "10.0.1.2")
		}
	}
}

//resolveUpperIntents resolves intents for higher level MLFOs
func resolveUpperIntents(in parser.Intent, pipe []map[string]string) {

	var faintent parser.IntentNoExp
	var fatargetList []parser.Target
	var fatarget1 parser.Target
	var fatarget2 parser.Target

	nodehostname, err := os.Hostname()
	if err != nil {
		log.Println(err.Error())
	}
	nodenum := strings.Split(nodehostname, ".")[1]

	for _, target := range in.Targets {
		if target.Operation == "maximise" {
			if target.Constraints.Privacylevel == "high" {
				if target.Constraints.Latency == "low" {
					//If this two conditions are met, it means a hiearchical fed agg intent is required
					faintent.IntentID = "fedintent-fog" + nodenum
					if strings.Contains(nodehostname, "edge") {
						// Target 1
						fatarget1.ID = "fog" + nodenum
						fatarget1.Operation = "aggregate"
						fatarget1.Operand = "model.federated"
						fatarget1.Constraints.Modelkind = pipe[0]["model"]
						fatarget1.Constraints.Sourcekind = pipe[0]["source"]
						// Target 2
						fatarget2.ID = "fog" + nodenum
						fatarget2.Operation = "aggregate.global"
						fatarget2.Operand = "model.federated"
						fatarget2.Constraints.Modelkind = pipe[0]["model"]
						fatarget2.Constraints.Sourcekind = pipe[0]["source"]
						fatarget2.Constraints.Minaccuracy = 90
						fatargetList = append(fatargetList, fatarget1)
						fatargetList = append(fatargetList, fatarget2)
						faintent.Targets = fatargetList
						newIntents["10.0."+nodenum+".1"] = faintent
					}
				}
			}
		} else if target.Operation == "aggregate.global" {
			if strings.Contains(nodehostname, "fog") {
				fatarget1.ID = "cloud0"
				fatarget1.Operation = "aggregate"
				fatarget1.Operand = "model.federated"
				fatarget1.Constraints.Modelkind = target.Constraints.Modelkind
				fatarget1.Constraints.Sourcekind = target.Constraints.Sourcekind
				fatargetList = append(fatargetList, fatarget1)
				faintent.Targets = fatargetList
				newIntents["10.0.0.1"] = faintent
			}
		} else {
			log.Println("No upper intents for cloud")
		}

	}
}

//sendIntents sends intents over Mo-Mo to all other MLFOs
func sendIntents() {
	var pbIntent *pb.Intent
	if len(newIntents) != 0 {
		log.Printf("Sending the following intents\n %+v\n", newIntents)
		for address, intentmsg := range newIntents {
			intentBytes, err := json.Marshal(intentmsg)
			if err != nil {
				log.Println(err.Error())
			}
			json.Unmarshal(intentBytes, &pbIntent)
			reply := Send(address+momoport, pbIntent) //handle status
			log.Printf("%+v", reply)
		}
	}
}

//deploylocal deploys local pipelines in the local domain
func deploylocal(pipelines []map[string]string) {
	nodehostname, err := os.Hostname()
	if err != nil {
		log.Println(err.Error())
	}
	if strings.Contains(nodehostname, "fog") {
		if !aggservpresent {
			//Simulate fed server creation delay
			time.Sleep(1 * time.Second)
			//set aggregation server present status to true
			aggservpresent = true
		}
	}
	if strings.Contains(nodehostname, "cloud") {
		if !globalaggservpresent {
			//Simulate fed server creation delay
			time.Sleep(1 * time.Second)
			//set global aggregation server present status to true
			globalaggservpresent = true
		}
	}
	log.Println("Deployed local pipelines on " + nodehostname)
	// for _, pipeline := range pipelines {
	// 	if pipeline["model"] == "federated" {
	// 		log.Println("Deploying aggregation server ")
	// 	}
	// }

}

//server is used to implement pb.UnimplementedOrchestrateServer
type server struct {
	pb.UnimplementedOrchestrateServer
}

//Deploy is called when a Mo-Mo message is received on MLFO server
func (s *server) Deploy(ctx context.Context, rcvdIntent *pb.Intent) (*pb.Status, error) {
	start2 := time.Now()
	log.Println(rcvdIntent)
	var intent parser.Intent
	intentbytes, err := json.Marshal(rcvdIntent)
	if err != nil {
		log.Println(err.Error())
	}
	json.Unmarshal(intentbytes, &intent)
	log.Printf("Received the following intent\n%v", intent)

	//here intent can be used as normal struct
	/*
		Step 1: Receive intent over http(:8000) OR over Mo-Mo(:9000)
		Step 2: Resolve local pipelines
		Step 3: Resolve foreign intents (peer + upper nodes)
		Step 4: Send intents over Mo-Mo
		Step 5: Deploy local pipelines
	*/
	pipelines := resolvePipeline(intent)
	resolvePeerIntents(intent)
	resolveUpperIntents(intent, pipelines)
	sendIntents()
	deploylocal(pipelines)

	elapsed2 := time.Since(start2)
	log.Printf("HTTP Intent took %s", elapsed2)
	//return status to client over mo-mo. This may also contain FedIP of created Fed Server
	status := "emptystatus"
	return &pb.Status{Status: status}, nil
}

//StartServer starts MLFO grpc server
func StartServer(port string) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	} else {
		log.Printf("\nStarted listening on %v\n", port)
	}
	s := grpc.NewServer()
	pb.RegisterOrchestrateServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

//Send sends msg over grpc
func Send(address string, message *pb.Intent) string {
	log.Printf("Connecting to %v ......", address)
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewOrchestrateClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	r, err := c.Deploy(ctx, message)
	if err != nil {
		log.Fatalf("could not receive: %v", err)
	}
	return r.GetStatus()

}
