package template

import (
	"fmt"
	"strings"

	"github.com/ivanvc/dispatcher/pkg/api/v1alpha1"
)

type Environment struct {
	UUID      string
	ShortUUID string
	Date      string
	Payload   string
}

func newEnvironment(jobExecution *v1alpha1.JobExecution) *Environment {
	escapedPayload := fmt.Sprintf("%q", jobExecution.Spec.Payload)
	return &Environment{
		UUID:      string(jobExecution.ObjectMeta.UID),
		ShortUUID: strings.Split(string(jobExecution.ObjectMeta.UID), "-")[0],
		Date:      jobExecution.ObjectMeta.CreationTimestamp.Format("20060102"),
		Payload:   escapedPayload[1 : len(escapedPayload)-1],
	}
}

func newEnvironmentFromPVCInstance(pvcInstance *v1alpha1.PersistentVolumeClaimInstance) *Environment {
	return &Environment{
		UUID:      pvcInstance.Spec.UUID,
		ShortUUID: strings.Split(string(pvcInstance.Spec.UUID), "-")[0],
		Date:      pvcInstance.Spec.Timestamp.Format("2006-01-02"),
	}
}
