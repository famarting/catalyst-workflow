package catalyst_workflow

import (
	"fmt"

	"github.com/dapr/go-sdk/workflow"
	"github.com/microsoft/durabletask-go/task"
)

func ParallelWorkflow(ctx *workflow.WorkflowContext) (any, error) {
	fmt.Println("start parallel workflow")
	defer fmt.Println("exit parallel workflow")

	total := 25
	tasks := []task.Task{}

	for i := 0; i < total; i++ {
		tasks = append(tasks, ctx.CallActivity(NoopActivity))
	}

	for _, t := range tasks {
		var out bool
		if err := t.Await(&out); err != nil {
			panic(err)
		}
	}

	var out bool
	if err := ctx.CallActivity(NoopActivity).Await(&out); err != nil {
		return nil, err
	}

	return "done", nil
}
