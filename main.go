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
	"sync"
	"time"

	pb "github.com/abh15/mlfo-dist/momo"
	"github.com/abh15/mlfo-dist/parser"
	"github.com/abh15/mlfo-dist/sbi"
	"google.golang.org/grpc"
)

func main() {

	//Start grpc server for momo on port 9000 in different thread
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		StartServer(":9000")
		wg.Done()
	}()

	//Start REST server for intent  on port 8000
	log.Print("Started REST server on :8000")
	http.HandleFunc("/receive", receiveHandler) // Handle the incoming file
	log.Fatal(http.ListenAndServe(":8000", nil))

	wg.Wait()
}

//receiveHandler handles the yaml file sent over REST
func receiveHandler(w http.ResponseWriter, r *http.Request) {

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
		intent := parser.Parse(buf.Bytes())
		if intent.DistIntent {
			//If intent is distributed handle it using Distributed()
			log.Println("Starting Distributed")
			Distributed(intent)
		} else {
			//If intent is not distributed deploy it locally using LocalDeploy()
			log.Println("Starting LocalDeploy")
			LocalDeploy(intent)
		}
		wg2.Done()
	}()
}

//server is used to implement pb.UnimplementedOrchestrateServer
type server struct {
	pb.UnimplementedOrchestrateServer
}

//Deploy is called when a message is received on MLFO server
func (s *server) Deploy(ctx context.Context, rcvdIntent *pb.Pipeline) (*pb.Status, error) {

	var intent parser.Intent
	bytes, err := json.Marshal(rcvdIntent)
	if err != nil {
		log.Println(err.Error())
	}
	json.Unmarshal(bytes, &intent)
	status := ""

	if intent.DistIntent {
		//If intent is distributed handle it using Distributed()
		Distributed(intent)
		status = "Received distributed intent"
		log.Println("Received Distributed intent on MLFO server")
	} else {
		//If intent is not distributed, deploy it locally using LocalDeploy()
		status = LocalDeploy(intent)
		log.Println("Received Local intent on MLFO server")
	}
	//return status to client over mo-mo. This may also contain FedIP of created Fed Server
	return &pb.Status{Status: status}, nil
}

//StartServer starts MLFO grpc server
func StartServer(port string) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	} else {
		log.Printf("\nStarted listening on %v\n\n", port)
	}
	s := grpc.NewServer()
	pb.RegisterOrchestrateServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

//LocalDeploy executes local pipeline orchestration
func LocalDeploy(intent parser.Intent) string {
	/* Local pipelets can be of various types-
	 Type 1: Local pipelet which is part of some distributed pipeline e.g federated, splitNN, Model chain etc.
	 Type 2: Local pipelet which not part of any distributed pipeline.
			   It is vanilla local pipline as described in ITU Y.3172
	Note: intent.Sources[0].Req.Num is overloaded.
			In fedavg case it min of clients required to start sampling.
			In Local case it is num of fed clients i.e robots to be started
	*/
	//Deploy Type 1 pipelet
	if intent.Type == "federated" {
		if intent.Models[0].ID == "FedAvg" {
			/* Local pipelet is on Fed server node. Match the requirements, and create a new server if required,
			   else use the existing server. Return IP of the server to the fed client so it can join.
			*/
			//match for model, src type
			isPresent, serverIP := sbi.MatchServer(intent.Models[0].Req.Kind, intent.Sources[0].Req.Kind)
			if isPresent {
				return serverIP
			}
			//StartFedServer starts a new fed server and return serviceIP for the server
			return sbi.StartFedServer(intent.Models[0].Req.Kind, intent.Sources[0].Req.Kind, intent.Sources[0].Req.Num)

		} else {
			/* Local pipelet is on edge node. Resolve requirements and deploy fed clients according to intent and
			IP received from server
			*/
			localsrc := sbi.ResolveRequirements("source", intent.Sources[0].Req)
			localmodel := sbi.ResolveRequirements("model", intent.Models[0].Req)
			localsink := sbi.ResolveRequirements("sink", intent.Sinks[0].Req)
			numClients := intent.Sources[0].Req.Num
			fedIP := intent.Servers[0].Server
			sbi.StartFedClients(localsrc, localmodel, localsink, fedIP, numClients)
			log.Printf("%v number of federated clients deployed", numClients)
			return "Fed clients deployed"
		}
	}
	//DeployType2Pipelet()
	return "Deployed type 2 pipelet"

}

//Send sends msg over grpc
func Send(address string, message *pb.Pipeline) string {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewOrchestrateClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.Deploy(ctx, message)
	if err != nil {
		log.Fatalf("could not receive: %v", err)
	}
	return r.GetStatus()

}

// Distributed handles distributed intents
func Distributed(in parser.Intent) {
	/* A federated pipeline can span over multiple edges and cloud domains. Each domain has its own MLFO.
	The section of the federated pipeline which resides within a domain is called pipelet.
	All the pipelets add up to make a multi-domain federated pipeline.
	*/
	log.Println("Distributed Initiated")
	var pipelet = make(map[string]*pb.Pipeline)
	var model *pb.Model
	var localsrckind string
	var totalnumofclients int32
	totalnumofclients = 0

	/* Structs defined in parser.go have one to one correspondance to structs
	defined in momo.pb.go which themselves are generated from the momo.proto.
	We use json marshal/unmarshal for converting to/from parser structs and pb strcuts
	*/
	modelBytes, err := json.Marshal(in.Models[0]) //in.Models[0] because we assume single fed model for all nodes
	if err != nil {
		log.Println(err.Error())
	}
	json.Unmarshal(modelBytes, &model)
	myhostname, err := os.Hostname() //Get hostname of this node
	if err != nil {
		log.Println(err.Error())
	}
	cloudMlfoIP := in.Servers[0].Server //IP of the single fed serer
	LocalPipelet := parser.Intent{}

	/* We decompose the intent in various steps:
	   Step 1: Create intent struct for local pipelet
	   Step 2: Create intent/momo struct for piplets of other edges
	   Step 3: Create intent/momo struct for federated server
	   Step 4: Send the struct to cloudmlfo, get fedserverip from response and LocalDeploy() using that IP
	   Step 5: Send the struct to other edge mlfos
	*/
	for i := 0; i < len(in.Sources); i++ {
		totalnumofclients = totalnumofclients + in.Sources[i].Req.Num
		//Step 1: If source ID is same as hostname create pipelet struct sourceN-model-sinkN to deploy locally
		if in.Sources[i].ID == myhostname {
			LocalPipelet.Type = "federated"
			LocalPipelet.Sources = []parser.Source{in.Sources[i]}
			LocalPipelet.Sinks = []parser.Sink{in.Sinks[i]}
			LocalPipelet.Models = []parser.Model{in.Models[0]}
			log.Println("LocalPipelet Created")
			localsrckind = in.Sources[i].Req.Kind

		} else {
			//Step 2: Creates a map where <pipeletTargetEdgeIP:pipeletStruct>
			var source *pb.Source
			var sink *pb.Sink
			sourceBytes, err := json.Marshal(in.Sources[i])
			if err != nil {
				log.Println(err.Error())
			}
			json.Unmarshal(sourceBytes, &source)
			sinkBytes, err := json.Marshal(in.Sinks[i])
			if err != nil {
				log.Println(err.Error())
			}
			json.Unmarshal(sinkBytes, &sink)
			pipelet[in.Sources[i].ID] = &pb.Pipeline{DistIntent: true, Type: "federated",
				Servers: []*pb.Server{{Server: cloudMlfoIP}}, Sources: []*pb.Source{source},
				Models: []*pb.Model{model}, Sinks: []*pb.Sink{sink}}
			log.Println("Pipelet Created")
		}
	}

	//Step 3: Create piplet struct for fed server
	//Send source kind and model kind for matching against existing servers
	localmodelkind := sbi.ResolveRequirements("model", in.Models[0].Req)
	modelreq := &pb.Requirements{Kind: localmodelkind}
	srcreq := &pb.Requirements{Kind: localsrckind, Num: totalnumofclients}
	fedservpipelet := &pb.Pipeline{
		DistIntent: false,
		Type:       "federated",
		Sources:    []*pb.Source{{ID: myhostname, Req: srcreq}},
		Models:     []*pb.Model{{ID: "FedAvg", Req: modelreq}},
		Sinks:      []*pb.Sink{{ID: myhostname}}}

	//Step 4: Send piplet struct to cloudMLFO
	status := Send(cloudMlfoIP, fedservpipelet)
	if status == "" {
		log.Println("Empty response from cloud MLFO")
	}
	LocalPipelet.Servers = []parser.Server{{Server: status}}
	log.Printf("Received response %v from fed server", status)
	log.Printf("Deploying local pipelet now")
	LocalDeploy(LocalPipelet)

	//Step 5: Send pipelet structs over Mo-Mo(gRPC) to other edge MLFOs
	for k, v := range pipelet {
		status := Send(k, v)
		if status == "" {
			log.Println("Empty response from edge MLFO")
		} else {
			log.Printf("Received response %v from edge", status)
		}
	}
}

//================================
/* //GetModelSegments describes logic of which node will host which model segment
func GetModelSegments(num int, node0 string, nodes []parser.Server) ([]string, []string) {
	var segments = make([]string, num+1)
	var locations = make([]string, num+1)

	segments[0] = "model.segment.0"
	locations[0] = node0
	for i := 1; i < num+1; i++ {
		locations[i] = nodes[i-1].Server
		segments[i] = "model.segment." + strconv.Itoa(i)
	}
	return segments, locations
}

/*
//SplitNN handles splitNN dist. intents
func SplitNN(in parser.Intent) {
	_ = in
	 	var pipelet = make(map[string]*pb.Pipeline)
	   	LocalPipelet := parser.Intent{}
	   	//Get model segments and their locations
	   	segments, locations := GetModelSegments(len(in.Servers), in.Sources[0].ID, in.Servers)
	   	//Deploy local intent
	   	LocalPipelet.Sources = []parser.Source{parser.Source{ID: locations[0]}}
	   	LocalPipelet.Models = []parser.Model{parser.Model{ID: segments[0]}}
	   	LocalPipelet.Sinks = []parser.Sink{parser.Sink{ID: locations[1]}}
	   	LocalDeploy(LocalPipelet)

	   	//Prepate pipelet msgs
	   	for i := 1; i < len(segments); i++ {
	   		//for the last segment sink should be origin(thishost)
	   		if i == len(segments)-1 {
	   			pipelet[in.Servers[i-1].Server] = &pb.Pipeline{DistIntent: false, Sources: []*pb.Source{{ID: locations[i]}},
	   				Models: []*pb.Model{{ID: segments[i]}}, Sinks: []*pb.Sink{{ID: locations[0]}}}
	   		} else {
	   			pipelet[in.Servers[i-1].Server] = &pb.Pipeline{DistIntent: false, Sources: []*pb.Source{{ID: locations[i]}},
	   				Models: []*pb.Model{{ID: segments[i]}}, Sinks: []*pb.Sink{{ID: locations[i+1]}}}
	   		}
	   	}
	   	// This CANNOT be used for testing
	   	//Send the pipelet msgs =================(assuming all other MLFOs are running)===================
	   	for k, v := range pipelet {
	   		status := Send(k, v)
	   		fmt.Printf("%+v", status)
	   	}
}*/
