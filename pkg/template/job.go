package template

import (
	"context"
	"encoding/json"
	"io"
	"text/template"

	batchv1 "k8s.io/api/batch/v1"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

func NewJobFromTemplate(ctx context.Context, spec batchv1.JobTemplateSpec, payload string) *batchv1.Job {
	log := ctrllog.FromContext(ctx)

	tpl, err := json.Marshal(spec)
	if err != nil {
		log.Error(err, "Error encoding Job template")
		return nil
	}
	t := template.Must(template.New("job").Parse(string(tpl)))

	r, w := io.Pipe()

	if err := t.Execute(w, newEnvironment(payload)); err != nil {
		log.Error(err, "Error executing Job template")
		return nil
	}

	var jobTpl *batchv1.JobTemplateSpec
	if err := json.NewDecoder(r).Decode(&jobTpl); err != nil {
		log.Error(err, "Error decoding Job template")
		return nil
	}

	return &batchv1.Job{
		ObjectMeta: jobTpl.ObjectMeta,
		Spec:       jobTpl.Spec,
	}
}
