package http

import (
	"net/http"
	"strings"

	"github.com/ivanvc/dispatcher/pkg/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

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
	jobExecution := e.createJobExecution(req)

	log.Info("Creating JobExecution")
	if err := e.Create(ctx, jobExecution); err != nil {
		log.Error(err, "Error creating JobExecution")
		w.WriteHeader(http.StatusNotAcceptable)
	}
	w.WriteHeader(http.StatusCreated)
}

func (e *executeJobHandler) createJobExecution(req *http.Request) *v1alpha1.JobExecution {
	name := strings.TrimPrefix(req.URL.Path, "/execute/")
	body := parseBody(req)
	if body == nil {
		return nil
	}

	return &v1alpha1.JobExecution{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name + "-",
			Namespace:    body.getNamespace(),
		},
		Spec: v1alpha1.JobExecutionSpec{
			JobTemplateName: name,
			Args:            body.getArgs(),
		},
	}
}
