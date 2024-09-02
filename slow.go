package catalyst_workflow

import (
	"fmt"
	"time"

	"github.com/dapr/go-sdk/workflow"
)

func SlowSimpleWorkflow(ctx *workflow.WorkflowContext) (any, error) {
	fmt.Println("start slow simple workflow")
	defer fmt.Println("exit slow simple workflow")

	var out bool
	if err := ctx.CallActivity(SlowActivity).Await(&out); err != nil {
		return nil, err
	}

	return "done", nil
}

func SlowActivity(ctx workflow.ActivityContext) (any, error) {
	fmt.Println("start slow activity")
	defer fmt.Println("exit slow activity")
	time.Sleep(15 * time.Second)
	return true, nil
}
