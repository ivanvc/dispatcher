package template

import (
	"encoding/json"
	"io"
	"text/template"

	"github.com/ivanvc/dispatcher/pkg/api/v1alpha1"
)

func BuildPersistentVolumeClaim(pvcTemplateSpec *v1alpha1.PersistentVolumeClaimTemplateSpec, jobExecution *v1alpha1.JobExecution) (*v1alpha1.PersistentVolumeClaimTemplateSpec, error) {
	tpl, err := json.Marshal(&pvcTemplateSpec)
	if err != nil {
		return nil, err
	}
	t := template.Must(template.New("pvc").Parse(string(tpl)))

	r, w := io.Pipe()
	if err := t.Execute(w, newEnvironment(jobExecution)); err != nil {
		return nil, err
	}

	var pvcTpl *v1alpha1.PersistentVolumeClaimTemplateSpec
	if err := json.NewDecoder(r).Decode(&pvcTpl); err != nil {
		return nil, err
	}

	return pvcTpl, nil
}
