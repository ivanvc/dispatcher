package template

import (
	"encoding/json"
	"io"
	"text/template"

	batchv1 "k8s.io/api/batch/v1"

	"github.com/ivanvc/dispatcher/pkg/api/v1alpha1"
)

func BuildJob(jobTemplateSpec *batchv1.JobTemplateSpec, jobExecution *v1alpha1.JobExecution) (*bachv1.JobTemplateSpec, error) {
	tpl, err := json.Marshal(&spec)
	if err != nil {
		return nil, err
	}
	t := template.Must(template.New("job").Parse(string(tpl)))

	r, w := io.Pipe()
	if err := t.Execute(w, newEnvironment(jobExecution)); err != nil {
		return nil, err
	}

	var jobTpl *batchv1.JobTemplateSpec
	if err := json.NewDecoder(r).Decode(&jobTpl); err != nil {
		return nil, err
	}

	return jobTpl, nil
}
