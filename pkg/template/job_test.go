package template

import (
	"strings"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ivanvc/dispatcher/pkg/api/v1alpha1"
)

func TestBuildJob(t *testing.T) {
	jt := &batchv1.JobTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name:    "pi",
							Image:   "alpine:{{ uuidv4 }}",
							Command: []string{"exit", "1"},
							Env: []corev1.EnvVar{
								corev1.EnvVar{
									Name:  "PAYLOAD",
									Value: `{{ if fromJson .Payload }}{{ .Payload | fromJson | toJson }}{{ else }}{{ .Payload }}{{ end }}`,
								},
							},
						},
					},
				},
			},
		},
	}

	je := &v1alpha1.JobExecution{
		Spec: v1alpha1.JobExecutionSpec{
			JobTemplateName: "test",
			Payload: `{
"hello": "world"
}`,
		},
	}

	job, err := BuildJob(jt, je)
	if err != nil {
		t.Error(err)
		return
	}
	if job.Spec.Template.Spec.Containers[0].Env[0].Value != "{&#34;hello&#34;:&#34;world&#34;}" {
		t.Error("Mismatch in PAYLOAD", job.Spec.Template.Spec.Containers[0].Env[0].Value)
	}
	if len(strings.Split(job.Spec.Template.Spec.Containers[0].Image, ":")[1]) != 37 {
		t.Errorf("Got wrong replacement with Sprig function %q", job.Spec.Template.Spec.Containers[0].Image)
	}
}
