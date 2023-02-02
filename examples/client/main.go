package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/spiffe/go-spiffe/v2/spiffegrpc/grpccredentials"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/svid/x509svid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

// Workload API socket path
const socketPath = "unix:///spire/agent.sock"

const tdClientString = "spiffe://accuknox.com/feeder-client"

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
		if svid.ID.String() == tdClientString {
			return svid
		}
	}

	return svid[0]
}

func run(ctx context.Context) error {
	// Create a `workloadapi.X509Source`, it will connect to Workload API using provided socket path
	// If socket path is not defined using `workloadapi.SourceOption`, value from environment variable `SPIFFE_ENDPOINT_SOCKET` is used.
	source, err := workloadapi.NewX509Source(ctx, workloadapi.WithDefaultX509SVIDPicker(customSVIDPicker), workloadapi.WithClientOptions(workloadapi.WithAddr(socketPath)))
	if err != nil {
		return fmt.Errorf("unable to create X509Source: %w", err)
	}

	defer source.Close()

	fmt.Println(source.GetX509SVID())

	serverID := spiffeid.RequireFromString(os.Getenv("SPIRE_SERVER_ID"))

	fmt.Printf("serverID: %v\n", serverID)

	// Dial the server with credentials that do mTLS and verify that presented certificate has SPIFFE ID `spiffe://example.org/server`

	serverAddr := os.Getenv("SERVER_ADDR")

	conn, err := grpc.DialContext(ctx, serverAddr, grpc.WithTransportCredentials(
		grpccredentials.MTLSClientCredentials(source, source, tlsconfig.AuthorizeID(serverID)),
	))
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}

	client := pb.NewGreeterClient(conn)
	reply, err := client.SayHello(ctx, &pb.HelloRequest{Name: "Hi I'm Feeder"})
	if err != nil {
		return fmt.Errorf("failed issuing RPC to server: %w", err)
	}

	log.Print(reply.Message)
	return nil
}
