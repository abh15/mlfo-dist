package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/abh15/mlfo-dist/parser"

	pb "github.com/abh15/mlfo-dist/momo"

	"google.golang.org/grpc"
)

const (
	port = ":8000"
	//address    = "localhost:8000"
	myhostname = "edge1"
)

var sourcesOut = make(map[string]*pb.Source)
var sinksOut = make(map[string]*pb.Sink)

func main() {
	if os.Args[1] == "client" {
		//StartClient()
		intent := parser.Parse(os.Args[2])
		srcMap, model, sinkMap := Federated(intent)
		for k, v := range srcMap {
			//send to k
			finalmsg := &pb.Pipeline{Src: v, Model: model, Sink: sinkMap[k]}
			status := Send(k, finalmsg)
			fmt.Printf("Send to :\t %+v", k)
			fmt.Printf("%+v", status)
		}
	} else {
		StartServer(os.Args[1])
	}

	// wg := new(sync.WaitGroup)
	// wg.Add(1)

	// go func() {
	// 	StartServer()
	// 	wg.Done()
	// }()

	// StartClient()

	// wg.Wait()

}

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedOrchestrateServer
}

// part of server
func (s *server) Deploy(ctx context.Context, mintent *pb.Pipeline) (*pb.Status, error) {
	msg := mintent.GetModel()

	fmt.Println(msg)

	// log.Printf("Received: %v", msg)
	// fmt.Println(mintent.GetModel())

	return &pb.Status{Status: "deployed successfully"}, nil
}

//StartServer ...
func StartServer(portt string) {
	lis, err := net.Listen("tcp", portt)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterOrchestrateServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

//Split handles SPlitNN
func Split() {}

//LocalDeploy handles Local pipeline deployment
func LocalDeploy() {}

//DelEmptyValues deletes empty values (and their keys) from map
func DelEmptyValues(inputmap map[string]string) map[string]string {
	for k, v := range inputmap {
		if v == "" {
			delete(inputmap, k)
		}
	}
	return inputmap
}

//ReqStrucToMap takes in struc and converts to map of string:string
func ReqStrucToMap(s parser.Requirements) map[string]string {
	m := make(map[string]string)
	j, _ := json.Marshal(s)
	json.Unmarshal(j, &m)
	return m
}

//Federated handles federated meta-intents
func Federated(in parser.Intent) (map[string]*pb.Source, *pb.Model, map[string]*pb.Sink) {
	//Step1: create protobuf msgs for all remote edge nodes and deploy local pipeline.
	//In IF case we prepare protobuf for all edges, while in ELSE we prepare protobuf for Fed server
	for _, v := range in.Sources {
		if v.ID != myhostname {
			reqmap := DelEmptyValues(ReqStrucToMap(v.Req))
			src := &pb.Source{Requirements: reqmap}
			sourcesOut[v.ID] = src
		} else {
			fedsrc := &pb.Source{ID: myhostname}
			sourcesOut[in.Location[0].Server] = fedsrc
			//
			LocalDeploy()
		}
	}
	for _, v := range in.Sinks {
		if v.ID != myhostname {
			reqmap := DelEmptyValues(ReqStrucToMap(v.Req))
			snk := &pb.Sink{Requirements: reqmap}
			sinksOut[v.ID] = snk
		} else {
			fedsink := &pb.Sink{ID: myhostname}
			sinksOut[in.Location[0].Server] = fedsink
			//
			LocalDeploy()
		}
	}

	mdlmap := DelEmptyValues(ReqStrucToMap(in.Models["model"].Req))
	mdl := &pb.Model{Requirements: mdlmap}
	return sourcesOut, mdl, sinksOut

}

//Send ...
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

	//log.Printf("Greeting: %s", r.GetStatus())
	// m := make(map[string]string)

	// m["a"] = "A"
	// m["b"] = "B"

	// sources := []*pb.Source{
	// 	{SourceID: "cu-up", Requirements: m},
	// 	{SourceID: "du", Requirements: m},
	// 	{SourceID: "cu-cp", Requirements: m},
	// }

	// model := &pb.Model{
	// 	ModelID:      "mit.splitNN",
	// 	Constraints:  m,
	// 	Requirements: m,
	// }

	// sinks := []*pb.Sink{
	// 	{SinkID: "app1.sink"},
	// 	{SinkID: "app2.sink", Requirements: m},
	// 	{SinkID: "app3.sink", Requirements: m},
	// }
	// if intent.DistIntent {
	// 	switch intent.Type {
	// 	case "federated":
	// 		Federated(intent)
	// 	case "splitNN":
	// 		Split()
	// 	}
	// } else {
	// 	LocalDeploy()
	// }

}
