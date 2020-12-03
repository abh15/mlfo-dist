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
			//If intent is distributed handle it using Federated()
			log.Println("Starting Federated")
			Federated(intent)
		} else {
			//If intent is not distributed deploy it locally using LocalDeploy()
			log.Println("Starting LocalDeploy")
			LocalDeploy(intent)
		}
		wg2.Done()
	}()
}

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedOrchestrateServer
}

//Deploy is called when a message is received on server side
func (s *server) Deploy(ctx context.Context, mintent *pb.Pipeline) (*pb.Status, error) {

	// fmt.Printf("%+v\n", mintent)
	var intent parser.Intent
	bytes, _ := json.Marshal(mintent)
	json.Unmarshal(bytes, &intent)
	status := "Received distributed pipeline"

	if intent.DistIntent {
		//If intent is distributed handle it using Federated()
		Federated(intent)
	} else {
		//If intent is not distributed deploy it locally using LocalDeploy()
		status = LocalDeploy(intent)
	}
	//========================intelligently return server status===================
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
func LocalDeploy(local parser.Intent) string {
	log.Println("Starting local deployment of intent")
	var localOutcome string

	if local.Type == "federated" {
		if local.Models[0].ID == "FedAvg" {
			_ = sbi.StartFedServer()
			localOutcome = "localhost:8080"

		} else {
			//start fed client local pipeline
			localsrc := sbi.ResolveRequirements("source", local.Sources[0].Req)
			localmodel := sbi.ResolveRequirements("model", local.Models[0].Req)
			localsink := sbi.ResolveRequirements("sink", local.Sinks[0].Req)
			numClients := local.Sources[0].Req.Num
			fedIP := local.Servers[0].Server
			sbi.StartFedClients(localsrc, localmodel, localsink, fedIP, numClients)
			localOutcome = "Started local clients"
		}
	}

	return localOutcome
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
		log.Fatalf("could not greet: %v", err)
	}
	return r.GetStatus()

}

// Federated handles federated intents
func Federated(in parser.Intent) {
	log.Println("Federated Initiated")
	/* A federated pipeline can span over multiple edges and cloud domains. Each domain has its own MLFO.
	The section of the federated pipeline which resides within a domain is called pipelet.
	All the pipelets add up to make a multi-domain federated pipeline.
	*/

	var pipelet = make(map[string]*pb.Pipeline)
	var mo *pb.Model

	/* Structs defined in parser.go have one to one correspondance to structs
	defined in momo.pb.go which themselves are generated from the momo.proto.
	We use json marshal/unmarshal for converting to/from parser structs and pb strcuts
	*/
	mobytes, err := json.Marshal(in.Models[0]) //in.Models[0] because we assume single fed model for all nodes
	if err != nil {
		log.Println(err.Error())
	}
	json.Unmarshal(mobytes, &mo)
	serv := in.Servers[0].Server //Servers[0] assuming single fed serer
	LocalIntent := parser.Intent{}
	myhostname, err := os.Hostname() //Get hostname of this node
	if err != nil {
		log.Println(err.Error())
	}

	/* We decompose the intent in various steps:
	   Step 1: Create intent struct for local pipelet
	   Step 2: Create intent/momo struct for piplets of other edges
	   Step 3: Create intent/momo struct for federated server
	   Step 4: Send the struct to other edges
	   Step 5: Send the struct to federated server, receive the fedserv IP,
				set ip in the struct and pass it to LocalDeploy()
	*/
	for i := 0; i < len(in.Sources); i++ {
		//Step 1: If source ID is same as hostname create local pipeline of sourceN-model-sinkN
		if in.Sources[i].ID == myhostname {
			LocalIntent.Type = "federated"
			LocalIntent.Sources = []parser.Source{in.Sources[i]}
			LocalIntent.Sinks = []parser.Sink{in.Sinks[i]}
			LocalIntent.Models = []parser.Model{in.Models[0]}
			log.Println("LocalIntent Created")
		} else {
			//Step 2: Creates pipelet map where <targetEdgeIP:intent_struct>
			var so *pb.Source
			var si *pb.Sink
			sobytes, err := json.Marshal(in.Sources[i])
			if err != nil {
				log.Println(err.Error())
			}
			json.Unmarshal(sobytes, &so)
			sibytes, err := json.Marshal(in.Sinks[i])
			if err != nil {
				log.Println(err.Error())
			}
			json.Unmarshal(sibytes, &si)
			pipelet[in.Sources[i].ID] = &pb.Pipeline{DistIntent: true, Type: "federated",
				Servers: []*pb.Server{{Server: serv}}, Sources: []*pb.Source{so},
				Models: []*pb.Model{mo}, Sinks: []*pb.Sink{si}}
			log.Println("Pipelet Created")
		}
	}

	//Step 3:
	pipelet[serv] = &pb.Pipeline{DistIntent: false, Type: "federated", Sources: []*pb.Source{{ID: myhostname}},
		Models: []*pb.Model{{ID: "FedAvg"}}, Sinks: []*pb.Sink{{ID: myhostname}}}

	//Step 4:
	for k, v := range pipelet {
		fmt.Printf("%+v\n", k)
		fmt.Printf("%+v\n", v)
		status := Send(k, v)
		if k == serv {
			LocalIntent.Servers = []parser.Server{{Server: status}}
			LocalDeploy(LocalIntent)
		}
		fmt.Println("This is the status")
		fmt.Println(status)
	}

	//Step 5:
	status := Send(serv, pipelet[serv])
	log.Printf("Request sent to Fed server: %v", serv)
	log.Printf("Received response %v", status)
	LocalIntent.Servers = []parser.Server{{Server: status}}
	LocalDeploy(LocalIntent)

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
	   	LocalIntent := parser.Intent{}
	   	//Get model segments and their locations
	   	segments, locations := GetModelSegments(len(in.Servers), in.Sources[0].ID, in.Servers)
	   	//Deploy local intent
	   	LocalIntent.Sources = []parser.Source{parser.Source{ID: locations[0]}}
	   	LocalIntent.Models = []parser.Model{parser.Model{ID: segments[0]}}
	   	LocalIntent.Sinks = []parser.Sink{parser.Sink{ID: locations[1]}}
	   	LocalDeploy(LocalIntent)

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
