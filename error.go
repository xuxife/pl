package pl

import (
	"fmt"
	"strings"
)

// ErrFlow indicates the error happens in passing Dependee's Output to Depender's Input.
//
//	Input(func(ctx context.Context, i *I) error {
//		return err
//	})
type ErrFlow struct {
	Err  error
	From StepReader
}

func (e *ErrFlow) Error() string {
	return fmt.Sprintf("ErrFlow(From %s [%s]): %s", e.From, e.From.GetStatus(), e.Err.Error())
}

// ErrWorkflow contains all errors of Steps in a Workflow.
type ErrWorkflow map[StepReader]error

func (e ErrWorkflow) Error() string {
	builder := new(strings.Builder)
	for reporter, err := range e {
		if err != nil {
			builder.WriteString(fmt.Sprintf(
				"%s [%s]: %s\n",
				reporter.String(), reporter.GetStatus().String(), err.Error(),
			))
		}
	}
	return builder.String()
}

func (e ErrWorkflow) IsNil() bool {
	for _, err := range e {
		if err != nil {
			return false
		}
	}
	return true
}

var ErrWorkflowIsRunning = fmt.Errorf("Workflow is running, please wait for it terminated")
var ErrWorkflowHasRun = fmt.Errorf("Workflow has run, check result error via Err(), or reset the Workflow via Reset()")

// Only when the Step status is not StepStautsPending when Workflow starts to run.
type ErrUnexpectStepInitStatus []StepReader

func (e ErrUnexpectStepInitStatus) Error() string {
	builder := new(strings.Builder)
	builder.WriteString("Unexpect Step initial status:")
	for _, j := range e {
		builder.WriteRune('\n')
		builder.WriteString(fmt.Sprintf(
			"%s [%s]",
			j, j.GetStatus(),
		))
	}
	return builder.String()
}

// There is a cycle-dependency in your Workflow!!!
type ErrCycleDependency map[StepReader][]StepReader

func (e ErrCycleDependency) Error() string {
	builder := new(strings.Builder)
	builder.WriteString("Cycle Dependency Error:")
	for j, deps := range e {
		depsStr := []string{}
		for _, dep := range deps {
			depsStr = append(depsStr, dep.String())
		}
		builder.WriteRune('\n')
		builder.WriteString(fmt.Sprintf(
			"%s: [%s]",
			j, strings.Join(depsStr, ", "),
		))
	}
	return builder.String()
}

// catchPanicAsError catches panic from f and return it as error.
// recoverFunc => func(recover()) (error)
func catchPanicAsError(f func() error, extractErrs ...func(any) error) error {
	var returnErr error
	func(err *error) {
		defer func() {
			if r := recover(); r != nil {
				for _, extract := range extractErrs {
					if xerr := extract(r); xerr != nil {
						*err = xerr
						return
					}
				}
				// otherwise, return the panic as error
				*err = fmt.Errorf("%s", r)
			}
		}()
		*err = f()
	}(&returnErr)
	return returnErr
}
