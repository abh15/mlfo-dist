package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	pb "github.com/abh15/mlfo-dist/momo"
	"github.com/abh15/mlfo-dist/parser"

	"google.golang.org/grpc"
)

const (
	port    = ":8000"
	address = "localhost:8000"
)

func main() {
	// start := time.Now()

	// if os.Args[1] == "server" {
	// 	StartServer()
	// } else if os.Args[1] == "client" {
	// 	StartClient()
	// } else {
	// 	sources, models, sinks = parser.Parse(os.Args[1])
	// }
	sources, models, sinks, isDist := parser.Parse(os.Args[1])
	fmt.Println(isDist)
	for _, src := range sources.(map[interface{}]interface{}) {
		//for every source in the sources list
		for k, v := range src.(map[interface{}]interface{}) {
			fmt.Printf("%v : ", k)
			fmt.Printf("%v\n\n", v)
		}
	}
	for _, model := range models.(map[interface{}]interface{}) {
		//for every model in models list
		for k, v := range model.(map[interface{}]interface{}) {
			fmt.Printf("%v : ", k)
			fmt.Printf("%v\n\n", v)
		}
	}
	for _, sink := range sinks.(map[interface{}]interface{}) {
		//for every sink in sinks list
		for k, v := range sink.(map[interface{}]interface{}) {
			fmt.Printf("%v : ", k)
			fmt.Printf("%v\n\n", v)
		}
	}
	// duration := time.Since(start)
	// fmt.Println(duration)

}

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedOrchestrateServer
}

// part of server
func (s *server) Deploy(ctx context.Context, mintent *pb.Pipeline) (*pb.Status, error) {
	msg := mintent.GetSrc()
	for _, v := range msg {

		fmt.Println(v.GetSourceID(), v.GetRequirements())
	}

	// log.Printf("Received: %v", msg)
	fmt.Println(mintent.GetModel())

	return &pb.Status{Status: "this is the current status"}, nil
}

//StartServer ist ein
func StartServer() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterOrchestrateServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

//StartClient ist ein
func StartClient() {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewOrchestrateClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	m := make(map[string]string)

	m["a"] = "A"
	m["b"] = "B"

	sources := []*pb.Source{
		{SourceID: "cu-up"},
		{SourceID: "du", Requirements: m},
		{SourceID: "cu-cp", Requirements: m},
	}

	model := &pb.Model{
		ModelID:      "mit.splitNN",
		Constraints:  m,
		Requirements: m,
	}

	sinks := []*pb.Sink{
		{SinkID: "app1.sink"},
		{SinkID: "app2.sink", Requirements: m},
		{SinkID: "app3.sink", Requirements: m},
	}

	finalmsg := &pb.Pipeline{Src: sources, Model: model, Sink: sinks}

	r, err := c.Deploy(ctx, finalmsg)
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetStatus())

}
