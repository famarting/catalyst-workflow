package catalyst_workflow

import (
	"fmt"

	"github.com/dapr/go-sdk/workflow"
)

func InfiniteWorkflow(ctx *workflow.WorkflowContext) (any, error) {
	fmt.Println("start infinite workflow")
	defer fmt.Println("exit infinite workflow")

	out := true
	for out {
		if err := ctx.CallActivity(BoolActivity).Await(&out); err != nil {
			return nil, err
		}
	}

	return "done", nil
}

func BoolActivity(ctx workflow.ActivityContext) (any, error) {
	fmt.Println("start activity")
	defer fmt.Println("exit activity")
	return true, nil
}
