package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/workflow"
	"github.com/dapr/kit/signals"
	catalyst_workflow "github.com/famarting/catalyst-workflow"
	"google.golang.org/grpc"
)

func main() {

	ctx := signals.Context()

	address, daprClientOpts := catalyst_workflow.GetGRPCOPTS()

	ctxConn, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	conn, err := grpc.DialContext(
		ctxConn,
		address,
		daprClientOpts...,
	)
	cancel()
	if err != nil {
		panic(err)
	}

	daprClient := client.NewClientWithConnection(conn)
	defer daprClient.Close()

	_, err = daprClient.GetMetadata(ctx)
	if err != nil {
		log.Fatal(err)
	}

	w, err := workflow.NewWorker(workflow.WorkerWithDaprClient(daprClient))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Worker initialized")

	if err := w.RegisterWorkflow(catalyst_workflow.TestWorkflow); err != nil {
		log.Fatal(err)
	}
	fmt.Println("TestWorkflow registered")

	if err := w.RegisterActivity(catalyst_workflow.TestActivity); err != nil {
		log.Fatal(err)
	}
	fmt.Println("TestActivity registered")

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
