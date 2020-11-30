package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	pb "github.com/abh15/mlfo-dist/momo"
	"github.com/abh15/mlfo-dist/parser"
	"github.com/abh15/mlfo-dist/sbi"
	"google.golang.org/grpc"
)

var myhostname string

func main() {

	// Usage: go run main.go -s=<serverip:serverport> -i=<intent>
	serveraddr := flag.String("s", "localhost:8000", "MLFO server will run on this addr:port default is localhost:8000")
	yamlpath := flag.String("i", "", "Intent YAML file full path")
	hostname := flag.String("h", "edge1", "Hostname of this node")
	flag.Parse()

	myhostname = *hostname
	//Start server in different thread
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		StartServer(*serveraddr)
		wg.Done()
	}()

	//parse intent and call respective function
	if *yamlpath != "" {
		intent := parser.Parse(*yamlpath)
		if intent.DistIntent {
			switch intent.Type {
			case "splitNN":
				SplitNN(intent)
			case "federated":
				Federated(intent)
			default:
				fmt.Println("Distributed type in intent not supported")
			}
		} else {
			LocalDeploy(intent)
		}
	}

	wg.Wait()
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
		switch intent.Type {
		case "splitNN":
			SplitNN(intent)
		case "federated":
			Federated(intent)
		default:
			fmt.Println("Distributed type in intent not supported")
		}
	} else {
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
		fmt.Printf("\nStarted listening on %v\n\n", port)
	}
	s := grpc.NewServer()
	pb.RegisterOrchestrateServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

//LocalDeploy handles Local pipeline deployment
func LocalDeploy(local parser.Intent) string {
	fmt.Printf("\n%+v\n", local)
	fmt.Println("\nExecuting local")
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

// Federated handles federated dist. intents
func Federated(in parser.Intent) {
	var mo *pb.Model
	var pipelet = make(map[string]*pb.Pipeline)
	mobytes, _ := json.Marshal(in.Models[0]) //assuming single fed model for all nodes
	json.Unmarshal(mobytes, &mo)
	serv := in.Servers[0].Server //assuming single fed serer
	LocalIntent := parser.Intent{}

	for i := 0; i < len(in.Sources); i++ {
		//create Local fed pipeline
		if in.Sources[i].ID == myhostname {
			LocalIntent.Type = "federated"
			LocalIntent.Sources = []parser.Source{in.Sources[i]}
			LocalIntent.Sinks = []parser.Sink{in.Sinks[i]}
			LocalIntent.Models = []parser.Model{in.Models[0]}
			//create pipeline for other edges
		} else {
			var so *pb.Source
			var si *pb.Sink
			sobytes, _ := json.Marshal(in.Sources[i])
			json.Unmarshal(sobytes, &so)
			sibytes, _ := json.Marshal(in.Sinks[i])
			json.Unmarshal(sibytes, &si)
			//construct momo msg for other edges
			pipelet[in.Sources[i].ID] = &pb.Pipeline{DistIntent: true, Type: "federated",
				Servers: []*pb.Server{{Server: serv}}, Sources: []*pb.Source{so},
				Models: []*pb.Model{mo}, Sinks: []*pb.Sink{si}}
		}
	}

	//create  pipeline for federated server
	pipelet[serv] = &pb.Pipeline{DistIntent: false, Type: "federated", Sources: []*pb.Source{{ID: myhostname}},
		Models: []*pb.Model{{ID: "FedAvg"}}, Sinks: []*pb.Sink{{ID: myhostname}}}

	status := Send(serv, pipelet[serv])

	fmt.Println(status)
	LocalIntent.Servers = []parser.Server{parser.Server{status}}
	LocalDeploy(LocalIntent)

	/* //send to target
	for k, v := range pipelet {
		fmt.Printf("%+v\n", k)
		fmt.Printf("%+v\n", v)
		status := Send(k, v)
		if k == serv {
			LocalIntent.Servers = []parser.Server{parser.Server{status}}
			LocalDeploy(LocalIntent)
		}
		fmt.Println("This is the status")
		fmt.Println(status)
	} */

}

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
} */

//SplitNN handles splitNN dist. intents
func SplitNN(in parser.Intent) {
	_ = in
	/* 	var pipelet = make(map[string]*pb.Pipeline)
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
	   	} */
}
