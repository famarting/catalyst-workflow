package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/workflow"
	"github.com/dapr/kit/signals"
	catalyst_workflow "github.com/famarting/catalyst-workflow"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

func HostFromURL(url string) string {
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimSuffix(url, "/")
	url = strings.TrimSuffix(url, ":80")
	url = strings.TrimSuffix(url, ":443")
	return url
}

func newDaprClient(url, token string) (client.Client, error) {

	targetHost := "127.0.0.1:30443"
	config := tls.Config{InsecureSkipVerify: true}
	credentialOption := grpc.WithTransportCredentials(credentials.NewTLS(&config))

	originalHost := HostFromURL(url)

	dialer := func(ctx context.Context, h string) (net.Conn, error) {
		if h == originalHost {
			return net.Dial("tcp4", targetHost)
		}
		return net.Dial("tcp4", h)
	}

	options := []grpc.DialOption{
		grpc.WithContextDialer(dialer),
		credentialOption,
		grpc.WithBlock(),
		// TODO: there's some duplication here from the go-sdk code
		grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			if token != "" {
				ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("dapr-api-token", token))
			}
			return invoker(ctx, method, req, reply, cc, opts...)
		}),
		grpc.WithStreamInterceptor(func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
			if token != "" {
				ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("dapr-api-token", token))
			}
			return streamer(ctx, desc, cc, method, opts...)
		}),
	}

	conn, err := grpc.Dial(originalHost, options...)
	if err != nil {
		return nil, err
	}

	return client.NewClientWithConnection(conn), nil
}

func main() {

	ctx := signals.Context()

	os.Setenv("DAPR_CLIENT_TIMEOUT_SECONDS", "10")

	address := os.Getenv("DAPR_GRPC_ENDPOINT")

	start := time.Now()
	var daprClient client.Client
	var err error
	if os.Getenv("ONEBOXV2_ENV") == "true" {
		daprClient, err = newDaprClient(address, os.Getenv("DAPR_API_TOKEN"))
	} else {
		daprClient, err = client.NewClientWithAddressContext(ctx, address)
	}
	if err != nil {
		panic(fmt.Errorf("error creating connection to '%s': %w", address, err))
	}
	fmt.Println("Time taken to connect to catalyst: " + time.Since(start).String())
	catalyst_workflow.SetDaprClient(daprClient)
	defer daprClient.Close()

	_, err = daprClient.GetMetadata(ctx)
	if err != nil {
		log.Fatal(err)
	}

	w, err := workflow.NewWorker(workflow.WorkerWithDaprClient(daprClient))
	if err != nil {
		log.Fatal(err)
	}

	if err := w.RegisterWorkflow(catalyst_workflow.TestWorkflow); err != nil {
		log.Fatal(err)
	}
	if err := w.RegisterActivity(catalyst_workflow.TestActivity); err != nil {
		log.Fatal(err)
	}

	if err := w.RegisterWorkflow(catalyst_workflow.InfiniteWorkflow); err != nil {
		log.Fatal(err)
	}
	if err := w.RegisterActivity(catalyst_workflow.BoolActivity); err != nil {
		log.Fatal(err)
	}

	if err := w.RegisterWorkflow(catalyst_workflow.LongWorkflow); err != nil {
		log.Fatal(err)
	}
	if err := w.RegisterActivity(catalyst_workflow.NoopActivity); err != nil {
		log.Fatal(err)
	}

	if err := w.RegisterWorkflow(catalyst_workflow.SimpleWorkflow); err != nil {
		log.Fatal(err)
	}

	if err := w.RegisterWorkflow(catalyst_workflow.SlowSimpleWorkflow); err != nil {
		log.Fatal(err)
	}
	if err := w.RegisterActivity(catalyst_workflow.SlowActivity); err != nil {
		log.Fatal(err)
	}

	if err := w.RegisterWorkflow(catalyst_workflow.ParallelWorkflow); err != nil {
		log.Fatal(err)
	}

	// Start workflow runner
	if err := w.Start(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("runner started")

	defer func() {
		// stop workflow runtime
		if err := w.Shutdown(); err != nil {
			log.Fatalf("failed to shutdown runtime: %v", err)
		}
		fmt.Println("workflow worker successfully shutdown")
	}()

	<-ctx.Done()
	fmt.Println("bye")

}
