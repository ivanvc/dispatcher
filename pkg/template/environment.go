package template

import (
	"fmt"
	"strings"

	"github.com/ivanvc/dispatcher/pkg/api/v1alpha1"
)

type Environment struct {
	Date      string
	Name      string
	Payload   string
	ShortUUID string
	UUID      string
}

func newEnvironment(jobExecution *v1alpha1.JobExecution) *Environment {
	escapedPayload := fmt.Sprintf("%q", jobExecution.Spec.Payload)
	return &Environment{
		Date:      jobExecution.ObjectMeta.CreationTimestamp.Format("20060102"),
		Name:      jobExecution.ObjectMeta.Name,
		Payload:   escapedPayload[1 : len(escapedPayload)-1],
		ShortUUID: strings.Split(string(jobExecution.ObjectMeta.UID), "-")[0],
		UUID:      string(jobExecution.ObjectMeta.UID),
	}
}
