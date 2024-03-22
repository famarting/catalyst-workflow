package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/workflow"
	catalyst_workflow "github.com/famarting/catalyst-workflow"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

func main() {

	address, daprClientOpts := catalyst_workflow.GetGRPCOPTS()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	conn, err := grpc.DialContext(
		ctx,
		address,
		daprClientOpts...,
	)
	cancel()
	if err != nil {
		panic(err)
	}

	ctx = context.Background()
	daprClient := client.NewClientWithConnection(conn)
	defer daprClient.Close()

	_, err = daprClient.GetMetadata(ctx)
	if err != nil {
		log.Fatal(err)
	}

	wfClient, err := workflow.NewClient(workflow.WithDaprClient(daprClient))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("staring workflow")
	wfID, err := wfClient.ScheduleNewWorkflow(context.Background(), "TestWorkflow", workflow.WithInstanceID(uuid.NewString()), workflow.WithInput(catalyst_workflow.MyInput{
		Quantities: []int{1, 2, 3, 4},
	}))
	if err != nil {
		log.Fatal(err)
	}

	res, err := wfClient.WaitForWorkflowCompletion(context.Background(), wfID)
	if err != nil {
		log.Fatal(err)
	}

	resS, err := json.Marshal(res)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(resS))

}
