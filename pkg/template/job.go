package template

import (
	"encoding/json"
	"io"
	"text/template"

	batchv1 "k8s.io/api/batch/v1"

	"github.com/Masterminds/sprig/v3"
	"github.com/ivanvc/dispatcher/pkg/api/v1alpha1"
)

func BuildJob(jobTemplateSpec *batchv1.JobTemplateSpec, jobExecution *v1alpha1.JobExecution) (*batchv1.JobTemplateSpec, error) {
	tpl, err := json.Marshal(&jobTemplateSpec)
	if err != nil {
		return nil, err
	}
	t := template.Must(template.New("job").Funcs(sprig.TxtFuncMap()).Parse(string(tpl)))

	r, w := io.Pipe()
	go func() {
		t.Execute(w, newEnvironment(jobExecution))
		w.Close()
	}()

	var jobTpl *batchv1.JobTemplateSpec
	if err := json.NewDecoder(r).Decode(&jobTpl); err != nil {
		return nil, err
	}

	return jobTpl, nil
}
