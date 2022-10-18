package template

import (
	"encoding/json"
	"io"
	"text/template"

	corev1 "k8s.io/api/core/v1"

	"github.com/ivanvc/dispatcher/pkg/api/v1alpha1"
)

func BuildPersistentVolumeClaim(pvcTemplateSpec *corev1.PersistentVolumeClaimSpec, pvcInstance *v1alpha1.PersistentVolumeClaimInstance) (*corev1.PersistentVolumeClaimSpec, error) {
	tpl, err := json.Marshal(&pvcTemplateSpec)
	if err != nil {
		return nil, err
	}
	t := template.Must(template.New("pvc").Parse(string(tpl)))

	r, w := io.Pipe()
	if err := t.Execute(w, newEnvironmentFromPVCInstance(pvcInstance)); err != nil {
		return nil, err
	}

	var pvcTpl *corev1.PersistentVolumeClaimSpec
	if err := json.NewDecoder(r).Decode(&pvcTpl); err != nil {
		return nil, err
	}

	return pvcTpl, nil
}
