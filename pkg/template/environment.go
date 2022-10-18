package template

import (
	"strings"

	"github.com/ivanvc/dispatcher/pkg/api/v1alpha1"
)

type Environment struct {
	UUID      string
	ShortUUID string
	Date      string
	Payload   string
}

func newEnvironmentFromJobExecution(jobExecution *v1alpha1.JobExecution) *Environment {
	return &Environment{
		UUID:      jobExecution.Spec.UUID,
		ShortUUID: strings.Split(string(jobExecution.Spec.UUID), "-")[0],
		Date:      jobExecution.Spec.Timestamp.Format("2006-01-02"),
		Payload:   jobExecution.Spec.Payload,
	}
}

func newEnvironmentFromPVCInstance(pvcInstance *v1alpha1.PersistentVolumeClaimInstance) *Environment {
	return &Environment{
		UUID:      pvcInstance.Spec.UUID,
		ShortUUID: strings.Split(string(pvcInstance.Spec.UUID), "-")[0],
		Date:      pvcInstance.Spec.Timestamp.Format("2006-01-02"),
	}
}
