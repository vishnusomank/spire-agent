package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/vishnusomank/go-spiffe/v2/spiffegrpc/grpccredentials"
	"github.com/vishnusomank/go-spiffe/v2/spiffeid"
	"github.com/vishnusomank/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/vishnusomank/go-spiffe/v2/svid/x509svid"
	"github.com/vishnusomank/go-spiffe/v2/workloadapi"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

//const socketPath = "unix:///tmp/spire-agent/public/api.sock"

const socketPath = "tcp://127.0.0.1:9098"
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

	meta := map[string]string{
		"sa_token": "eyJhbGciOiJSUzI1NiIsImtpZCI6ImZkMTRmODAzYWQwMWQ4Y2RmZWQ3M2NlZDY4YzRlMDBhMjlkMzllZmQifQ.eyJhdWQiOlsiaHR0cHM6Ly9rdWJlcm5ldGVzLmRlZmF1bHQuc3ZjIl0sImV4cCI6MTcwODQyODQ2NCwiaWF0IjoxNjc2ODkyNDY0LCJpc3MiOiJodHRwczovL29pZGMuZWtzLnVzLWVhc3QtMi5hbWF6b25hd3MuY29tL2lkL0NBOTE1QTNCRkJEOTI2MzY4NUEzQzJBMTc4QUQwOEY2Iiwia3ViZXJuZXRlcy5pbyI6eyJuYW1lc3BhY2UiOiJhY2N1a25veC1kZXYtY2x1c3Rlci1tZ210IiwicG9kIjp7Im5hbWUiOiJjbHVzdGVyLW1hbmFnZW1lbnQtc2VydmljZS04OTY4YjVmYjYtNTVjNGYiLCJ1aWQiOiI1MmFhNDMyMS1mY2ZiLTRiZjgtYWMzNC1iODNmYTYyZGFkMDAifSwic2VydmljZWFjY291bnQiOnsibmFtZSI6ImNsdXN0ZXItbWdtdCIsInVpZCI6ImM4OGM4ZDRiLWZjOTktNDc0Zi1hMTAwLWNlNWQwOTE1OWNlMyJ9LCJ3YXJuYWZ0ZXIiOjE2NzY4OTYwNzF9LCJuYmYiOjE2NzY4OTI0NjQsInN1YiI6InN5c3RlbTpzZXJ2aWNlYWNjb3VudDphY2N1a25veC1kZXYtY2x1c3Rlci1tZ210OmNsdXN0ZXItbWdtdCJ9.BDqzF_R3hWq63r9APRJ7d339tZWZQo-xOoMacMX3gZAnmTl-b-OBsbnRaF9LNN3tz5rqssdaXKMStptMcGcUCtWUQOe-Uc2jyjeaSPAip3J845VULR_GDrvQKPXmkmRux5v7zKXgDUbzQjGk37494lYCu1wiqHkq7GG58nw-qrb6_DQ1XI5-fwAM35cmxdatm1nJAREgBdzCxkoQIwcOa-sCWtjdwlgEK8b8XEHlQawZSh4nU_GWzjwzT_kR9UrWpz97whyzlnSdbRZhws-2wTipBRD_nXv7r8Z-Y_Ko2ZGZ0Wb32mznZf_PHacOR4c9EE2RVh5K5Q5dDsV0tkvJtw",
	}
	svids, err := workloadapi.FetchX509SVIDs(context.Background(), meta, workloadapi.WithAddr(socketPath))

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
