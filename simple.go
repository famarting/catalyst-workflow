package catalyst_workflow

import (
	"fmt"

	"github.com/dapr/go-sdk/workflow"
)

func SimpleWorkflow(ctx *workflow.WorkflowContext) (any, error) {
	fmt.Println("start simple workflow")
	defer fmt.Println("exit simple workflow")

	total := 10
	for i := 0; i < total; i++ {
		var out bool
		if err := ctx.CallActivity(NoopActivity).Await(&out); err != nil {
			return nil, err
		}
	}

	return "done", nil
}
