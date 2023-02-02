package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/spiffe/go-spiffe/v2/spiffegrpc/grpccredentials"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/svid/x509svid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const socketPath = "unix:///spire/agent.sock"

const tdServerString = "spiffe://accuknox.com/knoxgrpc"
const tdClientString = "spiffe://accuknox.com/feeder-client"

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	name := strings.Split(in.GetName(), " ")
	return &pb.HelloReply{Message: "Hello " + name[len(name)-1]}, nil
}

func main() {

	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func customSVIDPicker(svid []*x509svid.SVID) *x509svid.SVID {

	svids, err := workloadapi.FetchX509SVIDs(context.Background(), workloadapi.WithAddr(socketPath))

	if err != nil {
		return &x509svid.SVID{}
	}
	for _, svid := range svids {
		if svid.ID.String() == tdServerString {
			return svid
		}
	}

	return svid[0]
}

func run(ctx context.Context) error {
	// Create a `workloadapi.X509Source`, it will connect to Workload API using provided socket path
	// If socket path is not defined using `workloadapi.SourceOption`, value from environment variable `SPIFFE_ENDPOINT_SOCKET` is used.

	fmt.Println("Starting Execution")

	source, err := workloadapi.NewX509Source(ctx, workloadapi.WithDefaultX509SVIDPicker(customSVIDPicker), workloadapi.WithClientOptions(workloadapi.WithAddr(socketPath)))
	if err != nil {
		return fmt.Errorf("unable to create X509Source: %w", err)
	}

	fmt.Println(source.GetX509SVID())

	clientID := spiffeid.RequireFromString(os.Getenv("SPIRE_CLIENT_ID"))

	fmt.Printf("clientID: %v\n", clientID)

	// Create a server with credentials that do mTLS and verify that the presented certificate has SPIFFE ID `spiffe://example.org/client`
	s := grpc.NewServer(grpc.Creds(
		grpccredentials.MTLSServerCredentials(source, source, tlsconfig.AuthorizeID(clientID)),
	))
	fmt.Println("Starting to listen on 0.0.0.0:50051")
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		return fmt.Errorf("error creating listener: %w", err)
	}

	pb.RegisterGreeterServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}
	return nil
}
