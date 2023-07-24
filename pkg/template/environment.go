package template

import "github.com/ivanvc/dispatcher/pkg/api/v1beta1"

type Environment struct {
	Name    string
	Payload string
}

func newEnvironment(jobExecution *v1beta1.JobExecution) *Environment {
	return &Environment{
		Name:    jobExecution.ObjectMeta.Name,
		Payload: jobExecution.Spec.Payload,
	}
}
