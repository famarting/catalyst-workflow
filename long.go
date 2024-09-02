package catalyst_workflow

import (
	"fmt"
	"time"

	"github.com/dapr/go-sdk/workflow"
)

// dapr-scheduler-server-0.dapr-scheduler-server.root-dapr-system.svc.cluster.local:50006,dapr-scheduler-server-1.dapr-scheduler-server.root-dapr-system.svc.cluster.local:50006,dapr-scheduler-server-2.dapr-scheduler-server.root-dapr-system.svc.cluster.local:50006

func LongWorkflow(ctx *workflow.WorkflowContext) (any, error) {
	fmt.Println("start long workflow")
	defer fmt.Println("exit long workflow")

	total := 200
	for i := 0; i < total; i++ {
		var out bool
		if err := ctx.CallActivity(NoopActivity).Await(&out); err != nil {
			return nil, err
		}
	}

	return "done", nil
}

func NoopActivity(ctx workflow.ActivityContext) (any, error) {
	fmt.Println("start activity")
	defer fmt.Println("exit activity")
	time.Sleep(100 * time.Millisecond)
	return true, nil
}
