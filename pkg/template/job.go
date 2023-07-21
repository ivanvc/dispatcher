package template

import (
	batchv1 "k8s.io/api/batch/v1"

	"github.com/ivanvc/dispatcher/pkg/api/v1beta1"
)

func BuildJob(jobTemplateSpec *batchv1.JobTemplateSpec, jobExecution *v1beta1.JobExecution) (*batchv1.JobTemplateSpec, error) {
	tpl := newGenericTemplate(jobTemplateSpec.DeepCopy(), newEnvironment(jobExecution))
	err := tpl.execute()

	return tpl.target.(*batchv1.JobTemplateSpec), err
}
