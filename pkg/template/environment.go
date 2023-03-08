package template

import (
	"strings"

	"github.com/ivanvc/dispatcher/pkg/api/v1alpha1"
)

type Environment struct {
	Name    string
	Payload string
}

func newEnvironment(jobExecution *v1alpha1.JobExecution) *Environment {
	escapedPayload := fmt.Sprintf("%q", jobExecution.Spec.Payload)
	return &Environment{
		Name:    jobExecution.ObjectMeta.Name,
		Payload: escapedPayload[1 : len(escapedPayload)-1],
	}
}
