package taskDerive

import (
	"context"

	"github.com/v2fly/v2ray-core/v5/common/task"
)

func RunTryAll(ctx context.Context, tasks ...func() error) []error {
	errors := make([]error, len(tasks))
	wrappedTasks := make([]func() error, len(tasks))
	for i, currentTask := range tasks {
		index := i
		wrappedTasks[i] = func() error {
			err := currentTask()
			errors[index] = err
			if err != nil {
				return nil
			}
			return newError()
		}
	}
	runErr := task.Run(ctx, wrappedTasks...)
	if runErr != nil {
		return nil
	}
	return errors
}
