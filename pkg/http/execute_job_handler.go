package http

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	log.Info("Creating JobExecution", "name", name, "ns", ns)
	jobExecution := createJobExecution(name, ns, req.Body)

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

func createJobExecution(name, ns string, body io.ReadCloser) *v1alpha1.JobExecution {
	var b bytes.Buffer
	if body != nil {
		defer body.Close()
		io.Copy(&b, body)
	}

	return &v1alpha1.JobExecution{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name + "-",
			Namespace:    ns,
		},
		Spec: v1alpha1.JobExecutionSpec{
			JobTemplateName: name,
			Payload:         b.String(),
		},
	}
}
