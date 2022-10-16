package http

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/ivanvc/dispatcher/pkg/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
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

	jobExecution, err := createJobExecution(req)
	if err != nil {
		log.Error(err, "Error creating JobExecution")
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

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

func createJobExecution(req *http.Request) (*v1alpha1.JobExecution, error) {
	name, namespace, err := getNameAndNamespace(req.URL.Path)
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	if req.Body != nil {
		defer req.Body.Close()
		io.Copy(&b, req.Body)
	}

	return &v1alpha1.JobExecution{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name + "-",
			Namespace:    namespace,
		},
		Spec: v1alpha1.JobExecutionSpec{
			JobTemplateName: name,
			Payload:         b.String(),
		},
	}, nil
}
