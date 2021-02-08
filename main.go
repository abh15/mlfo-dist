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
	intentport = ":8000"
	momoport   = ":9000"
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
	http.HandleFunc("/receive", httpReceiveHandler)       // Handle the incoming intent
	http.HandleFunc("/cloudreset", httpCloudResetHandler) // Handle the incoming reset msg
	http.HandleFunc("/fogreset", httpFogResetHandler)     // Handle the incoming internal reset msg
	log.Fatal(http.ListenAndServe(intentport, nil))

	wg.Wait()
}

//httpCloudResetHandler handles the reset request from external user, resets cloud node and triggers fogreset
func httpCloudResetHandler(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		log.Println(err.Error())
		fmt.Fprintf(w, "Bad Request")
	}
	if sbi.CheckServer() {
		sbi.DeleteFile("/fedserv")
	}
	log.Println(r.Form)

	//send fog reset parallely to all fogs
	var waitgroup sync.WaitGroup

	numfog, _ := strconv.Atoi(r.Form.Get("numfog"))
	waitgroup.Add(numfog)

	for i := 1; i <= numfog; i++ {
		go func(i int) {
			//create numnode number of reset reqests for all fog nodes
			resp, err := http.Get("http://10.0." + strconv.Itoa(i) + ".1:8000/fogreset")
			if err != nil {
				log.Println(err)
			}
			defer resp.Body.Close()
			waitgroup.Done()
		}(i)
	}
	waitgroup.Wait()

	fmt.Fprintf(w, "OK")

}

//httpFogResetHandler deletes foghit/fedserv files to reset the state of the experiment
func httpFogResetHandler(w http.ResponseWriter, _ *http.Request) {
	if sbi.CheckServer() {
		sbi.DeleteFile("/fedserv")
	}
	if sbi.CheckFogHit() {
		sbi.DeleteFile("/foghit")
	}
	fmt.Fprintf(w, "OK")

}

//receiveHandler handles the yaml file sent over REST
func httpReceiveHandler(w http.ResponseWriter, r *http.Request) {
	var outgoingIntents = make(map[string]parser.IntentNoExp)
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
		outgoingIntents = resolvePeerIntents(intent, outgoingIntents)
		outgoingIntents = resolveUpperIntents(intent, pipelines, outgoingIntents)
		sendIntents(outgoingIntents)
		_ = deploylocal(pipelines)

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
func resolvePeerIntents(in parser.Intent, newIntents map[string]parser.IntentNoExp) map[string]parser.IntentNoExp {

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
	return newIntents
}

//resolveUpperIntents resolves intents for higher level MLFOs
func resolveUpperIntents(in parser.Intent, pipe []map[string]string, newIntents map[string]parser.IntentNoExp) map[string]parser.IntentNoExp {

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
					faintent.IntentID = "fedintent-fog-" + nodehostname
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
				mutex.Lock()
				if sbi.CheckFogHit() == false {
					//if server does not exist create one
					fatarget1.ID = "cloud0"
					fatarget1.Operation = "aggregate"
					fatarget1.Operand = "model.federated"
					fatarget1.Constraints.Modelkind = target.Constraints.Modelkind
					fatarget1.Constraints.Sourcekind = target.Constraints.Sourcekind
					fatargetList = append(fatargetList, fatarget1)
					faintent.Targets = fatargetList
					newIntents["10.0.0.1"] = faintent
					sbi.RegisterFogHit()
				}
				mutex.Unlock()

			}
		} else {
			log.Println("No upper intents for cloud")
		}

	}
	return newIntents
}

//sendIntents sends intents over Mo-Mo to all other MLFOs
func sendIntents(outIntents map[string]parser.IntentNoExp) {
	var waitgroup sync.WaitGroup
	var outpbIntents = make(map[string]*pb.Intent)
	//Convert intent struc to pb intent struc
	if len(outIntents) != 0 {
		for address, intentmsg := range outIntents {
			var pbIntent *pb.Intent
			intentBytes, err := json.Marshal(intentmsg)
			if err != nil {
				log.Println(err.Error())
			}
			json.Unmarshal(intentBytes, &pbIntent)
			outpbIntents[address+momoport] = pbIntent
		}
		waitgroup.Add(len(outpbIntents))
		for sockadd, pbmsg := range outpbIntents {
			go func(sockadd string, pbmsg *pb.Intent) {
				log.Printf("Sending this intent to-- %+v\n%+v\n", sockadd, pbmsg)
				reply := Send(sockadd, pbmsg) //handle status
				log.Printf("Reply is %+v", reply)
				waitgroup.Done()
			}(sockadd, pbmsg)
		}
		waitgroup.Wait()
	}
}

//deploylocal deploys local pipelines in the local domain
func deploylocal(pipelines []map[string]string) string {
	nodehostname, err := os.Hostname()
	if err != nil {
		log.Println(err.Error())
	}
	if strings.Contains(nodehostname, "cloud") {
		mutex.Lock()
		if sbi.CheckServer() == false {
			//if server does not exist create one
			sbi.LaunchServer()
		}
		mutex.Unlock()
	}
	if strings.Contains(nodehostname, "fog") {
		mutex.Lock()
		if sbi.CheckServer() == false {
			//if server does not exist create one
			sbi.LaunchServer()
		}
		mutex.Unlock()
	}
	if strings.Contains(nodehostname, "edge") {

		log.Println("Local edge pipeline deployed")
		// mutex.Lock()
		// sbi.CreateFedMLCient(edgedelay)
		// mutex.Unlock()
	}
	return nodehostname
}

//Deploy is called when a Mo-Mo message is received on MLFO server
func (s *server) Deploy(ctx context.Context, rcvdIntent *pb.Intent) (*pb.Status, error) {
	start2 := time.Now()

	//newIntents has structure <intent_target_hostname, intent>. It consists of intents for peers as well as higher nodes
	var outgoingIntents = make(map[string]parser.IntentNoExp)
	var intent parser.Intent
	intentbytes, err := json.Marshal(rcvdIntent)
	if err != nil {
		log.Println(err.Error())
	}
	json.Unmarshal(intentbytes, &intent)
	log.Printf("Received the following intent\n%v\n", intent)

	//here intent can be used as normal struct
	/*
		Step 1: Receive intent over http(:8000) OR over Mo-Mo(:9000)
		Step 2: Resolve local pipelines
		Step 3: Resolve foreign intents (peer + upper nodes)
		Step 4: Send intents over Mo-Mo
		Step 5: Deploy local pipelines
	*/

	pipelines := resolvePipeline(intent)
	outgoingIntents = resolvePeerIntents(intent, outgoingIntents)
	outgoingIntents = resolveUpperIntents(intent, pipelines, outgoingIntents)
	sendIntents(outgoingIntents)
	status := deploylocal(pipelines)

	elapsed2 := time.Since(start2)
	log.Printf("MoMo Intent took %s", elapsed2)
	reply := status + "replied"
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
