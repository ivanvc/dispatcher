package http

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/ivanvc/dispatcher/pkg/api/v1alpha1"
)

const defaultNamespace = "default"

type executeJobHandler struct {
	*Server
}

func (e *executeJobHandler) registerHandler() {
	http.HandleFunc("/execute/", e.handle)
}

func (e *executeJobHandler) handle(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost && req.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx := req.Context()
	log := ctrllog.FromContext(ctx)

	name, ns, err := getNameAndNamespace(req.URL.Path)
	if err != nil {
		log.Error(err, "Error getting name and namespace")
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	jt, err := e.getJobTemplate(ns, name, ctx)
	if err != nil {
		log.Error(err, "JobTemplate doesn't exist", "name", name, "namespace", ns)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	log.Info("Creating JobExecution", "name", name, "namespace", ns)
	jobExecution := createJobExecution(ns, name, jt, req.Body)

	if err := e.Create(ctx, jobExecution); err != nil {
		log.Error(err, "Error creating JobExecution")
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func getNameAndNamespace(path string) (name, namespace string, err error) {
	if n := strings.Split(strings.TrimPrefix(path, "/execute/"), "/"); len(n) == 0 {
		return "", "", errors.New("Empty job name")
	} else if len(n) > 1 {
		namespace = n[0]
		name = n[1]
	} else {
		namespace = defaultNamespace
		name = n[0]
	}
	return
}

func createJobExecution(namespace, name string, jobTemplate *v1alpha1.JobTemplate, body io.ReadCloser) *v1alpha1.JobExecution {
	var b bytes.Buffer
	if body != nil {
		defer body.Close()
		io.Copy(&b, body)
	}

	return &v1alpha1.JobExecution{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name + "-",
			Namespace:    namespace,
			Labels:       jobTemplate.ObjectMeta.Labels,
		},
		Spec: v1alpha1.JobExecutionSpec{
			JobTemplateName: name,
			Payload:         b.String(),
		},
	}
}

func (e *executeJobHandler) getJobTemplate(namespace, name string, ctx context.Context) (*v1alpha1.JobTemplate, error) {
	jt := new(v1alpha1.JobTemplate)
	err := e.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, jt)
	return jt, err
}
