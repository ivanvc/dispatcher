package template

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ivanvc/dispatcher/pkg/api/v1beta1"
)

var jobTemplateSpec = &batchv1.JobTemplateSpec{
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

func TestBuildJobWithAJSON(t *testing.T) {
	je := &v1beta1.JobExecution{
		Spec: v1beta1.JobExecutionSpec{
			JobTemplateName: "test",
			Payload: `{
"hello": "world"
}`,
		},
	}

	job, err := BuildJob(jobTemplateSpec, je)
	if err != nil {
		t.Error(err)
		return
	}
	if job.Spec.Template.Spec.Containers[0].Env[0].Value != `{"hello":"world"}` {
		t.Error("Mismatch in PAYLOAD", job.Spec.Template.Spec.Containers[0].Env[0].Value)
	}
	if reflect.DeepEqual(job, jobTemplateSpec) {
		t.Errorf("Input JobTemplate shouldn't be modified")
	}
	if len(strings.Split(job.Spec.Template.Spec.Containers[0].Image, ":")[1]) != 36 {
		t.Errorf("Got wrong replacement with Sprig function %q", job.Spec.Template.Spec.Containers[0].Image)
	}
}

func TestBuildJobWithoutAJSON(t *testing.T) {
	je := &v1beta1.JobExecution{
		Spec: v1beta1.JobExecutionSpec{
			JobTemplateName: "test",
			Payload: `
"hello": "world"
`,
		},
	}

	job, err := BuildJob(jobTemplateSpec, je)
	if err != nil {
		t.Error(err)
		return
	}
	if job.Spec.Template.Spec.Containers[0].Env[0].Value != "\n\"hello\": \"world\"\n" {
		t.Error("Mismatch in PAYLOAD", job.Spec.Template.Spec.Containers[0].Env[0].Value)
	}
}

func TestBuildJobWithAnError(t *testing.T) {
	jt := &batchv1.JobTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Name: `{{fail "expected error"}}`,
		},
	}

	if _, err := BuildJob(jt, &v1beta1.JobExecution{}); err == nil {
		t.Error("Expecting error, got nothing")
	}
}

func TestExecuteTemplateWithAJobTemplate(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/job_template.json")
	if err != nil {
		t.Fatal(err)
	}

	var jt v1beta1.JobTemplate
	if err := json.Unmarshal(b, &jt); err != nil {
		t.Fatal(err)
	}

	tpl := newGenericTemplate(&jt, &Environment{"Name", `{"date":"2022-12-12"}`})
	if err := tpl.execute(); err != nil {
		t.Error(err)
	}
}
