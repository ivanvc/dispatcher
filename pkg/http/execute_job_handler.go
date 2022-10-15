package http

import (
	"net/http"
	"strings"

	"github.com/ivanvc/dispatcher/pkg/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	namespace = "default"
)

type executeJobHandler struct {
	client.Client
}

func (e *executeJobHandler) registerHandler() {
	http.HandleFunc("/execute/", e.handle)
}

func (e *executeJobHandler) handle(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	log := ctrllog.FromContext(ctx)
	name := strings.TrimPrefix(req.URL.Path, "/execute/")
	jobExecution := &v1alpha1.JobExecution{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name + "-",
			Namespace:    namespace,
		},
		Spec: v1alpha1.JobExecutionSpec{
			JobTemplateName: name,
		},
	}
	log.Info("Creating JobExecution", "jobTemplateName", name)
	if err := e.Create(ctx, jobExecution); err != nil {
		log.Error(err, "Error creating JobExecution", "jobTemplateName", name)
		w.WriteHeader(http.StatusNotAcceptable)
	}
	w.WriteHeader(http.StatusCreated)
}
