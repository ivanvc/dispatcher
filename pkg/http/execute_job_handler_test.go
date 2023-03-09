package http

import (
	"bytes"
	"io/ioutil"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ivanvc/dispatcher/pkg/api/v1alpha1"
)

func TestGetNameAndNamespaceWithAnError(t *testing.T) {
	tt := []string{
		"/execute/",
		"/execute",
		"/execute//",
		"/execute///",
		"/execute/a/b/c",
	}
	for _, tc := range tt {
		if _, _, err := getNameAndNamespace(tc); err == nil {
			t.Errorf("Expecting error with input %q, got nothing", tc)
			return
		}
	}
}

func TestGetNameAndNamespaceWithoutAnError(t *testing.T) {
	tt := [][]string{
		[]string{"/execute/a", "default", "a"},
		[]string{"/execute/a/b", "a", "b"},
	}
	for _, tc := range tt {
		name, ns, err := getNameAndNamespace(tc[0])
		if err != nil {
			t.Error(err)
			return
		}
		if name != tc[2] {
			t.Errorf("Expected name to be %q, got %q", name, tc[2])
		}
		if ns != tc[1] {
			t.Errorf("Expected namespace to be %q, got %q", ns, tc[1])
		}
	}
}

func TestCreateJobExecutionWithoutARequestBody(t *testing.T) {
	jt := &v1alpha1.JobTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels:    map[string]string{"test": "true"},
		},
		Spec: v1alpha1.JobTemplateSpec{},
	}
	je := createJobExecution(jt, nil)

	if je.ObjectMeta.GenerateName != "test-" {
		t.Errorf("Expected JobExecution GenerateName to be %q, got %q", "test-", je.ObjectMeta.GenerateName)
	}
	if je.ObjectMeta.Namespace != "default" {
		t.Errorf("Expected JobExecution GenerateName to be %q, got %q", "default", je.ObjectMeta.Namespace)
	}
	if len(je.ObjectMeta.Labels) != 1 || je.ObjectMeta.Labels["test"] != "true" {
		t.Errorf("Expected JobExecution Labels to be %q, got %q", jt.ObjectMeta.Labels, je.ObjectMeta.Labels)
	}
	if je.Spec.JobTemplateName != "test" {
		t.Errorf("Expected JobExecutionSpec JobTemplateName to be %q, got %q", "test", je.Spec.JobTemplateName)
	}
}

func TestCreateJobExecutionWithARequestBody(t *testing.T) {
	jt := &v1alpha1.JobTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels:    map[string]string{"test": "true"},
		},
		Spec: v1alpha1.JobTemplateSpec{},
	}
	body := ioutil.NopCloser(bytes.NewReader([]byte("testing")))
	je := createJobExecution(jt, body)

	if je.Spec.Payload != "testing" {
		t.Errorf("Expected JobExecutionSpec Payload to be %q, got %q", "testing", je.Spec.Payload)
	}
}
