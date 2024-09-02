package catalyst_workflow

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/dapr/go-sdk/client"
	"github.com/dapr/go-sdk/workflow"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var daprClient client.Client

func SetDaprClient(c client.Client) {
	daprClient = c
}

type MyInput struct {
	ActivityName string
	Quantities   []int
}

type MyActivityResult struct {
	Result int
}

type MyResult struct {
	ID     string
	Result int
}

func TestWorkflow(ctx *workflow.WorkflowContext) (any, error) {
	fmt.Printf("invoking workflow %v , replaying: %v \n", ctx.InstanceID(), ctx.IsReplaying())
	var input MyInput
	if err := ctx.GetInput(&input); err != nil {
		return nil, err
	}

	input.ActivityName = "first"

	var output MyActivityResult
	err := ctx.CallActivity(TestActivity, workflow.ActivityInput(input)).Await(&output)
	retryNum := 0
	for err != nil {
		// infinite retry
		input.ActivityName = "first" + strconv.Itoa(retryNum)
		err = ctx.CallActivity(TestActivity, workflow.ActivityInput(input)).Await(&output)
		retryNum++
	}

	if err := ctx.CallActivity(TestActivity, workflow.ActivityInput(MyInput{
		ActivityName: "second",
		Quantities:   append(input.Quantities, output.Result),
	})).Await(&output); err != nil {
		return nil, err
	}

	return &MyResult{
		ID:     ctx.Name() + "-" + ctx.InstanceID(),
		Result: output.Result,
	}, nil
}

func TestActivity(ctx workflow.ActivityContext) (any, error) {
	var input MyInput
	if err := ctx.GetInput(&input); err != nil {
		return MyActivityResult{}, err
	}
	fmt.Println("invoking activity " + input.ActivityName)

	// if rand.Intn(100) <= 15 {
	// 	return MyActivityResult{}, errors.New("random unexpected error triggered")
	// }

	var r int
	for _, q := range input.Quantities {
		r += q
	}

	ictx, cancel := context.WithTimeout(ctx.Context(), 3*time.Second)
	defer cancel()
	out, err := daprClient.InvokeMethod(ictx, "target", "hello", "GET")
	if err != nil {
		fmt.Println("hello call failed " + err.Error())
		return nil, err
	}
	if out != nil {
		fmt.Println("hello response " + string(out))
	} else {
		fmt.Println("hello call successfull")
	}

	return MyActivityResult{
		Result: r,
	}, nil
}

func GetGRPCOPTS() (string, []grpc.DialOption) {
	port := 443
	hostname := os.Getenv("DAPR_GRPC_ENDPOINT")
	// assume TLS by default
	daprClientOpts := []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(new(tls.Config)))}

	if isOnebox := os.Getenv("LOCAL_ENV"); isOnebox == "true" {
		port = 30011
		originalHostname := hostname
		daprClientOpts = []grpc.DialOption{
			grpc.WithAuthority(originalHostname),
			grpc.WithBlock(),
		}
		daprClientOpts = append(daprClientOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		hostname = "localhost"
	}

	address := fmt.Sprintf("%v:%d", hostname, port)

	apiToken := os.Getenv("DAPR_API_TOKEN")

	daprClientOpts = append(daprClientOpts, grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		token := apiToken
		if token != "" {
			ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("dapr-api-token", token))
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}))
	daprClientOpts = append(daprClientOpts, grpc.WithStreamInterceptor(func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		token := apiToken
		if token != "" {
			ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("dapr-api-token", token))
		}
		return streamer(ctx, desc, cc, method, opts...)
	}))

	return address, daprClientOpts
}
