package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/abh15/mlfo-dist/parser"

	pb "github.com/abh15/mlfo-dist/momo"

	"google.golang.org/grpc"
)

const (
	port    = ":8000"
	address = "localhost:8000"
)

func main() {

	if os.Args[1] == "client" {
		StartClient(os.Args[2])
	} else {
		StartServer()
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
func StartClient(intentpath string) {

	intent := parser.Parse(intentpath)

	for k, v := range intent.Sources {
		fmt.Printf("%+v\t", k)
		fmt.Printf("%+v\n", v)
	}

	// conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	// if err != nil {
	// 	log.Fatalf("did not connect: %v", err)
	// }
	// defer conn.Close()
	// c := pb.NewOrchestrateClient(conn)

	// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// defer cancel()

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

	// finalmsg := &pb.Pipeline{Src: sources, Model: model, Sink: sinks}

	// r, err := c.Deploy(ctx, finalmsg)
	// if err != nil {
	// 	log.Fatalf("could not greet: %v", err)
	// }
	// log.Printf("Greeting: %s", r.GetStatus())

}
