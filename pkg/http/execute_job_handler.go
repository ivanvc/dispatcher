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
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	"github.com/ivanvc/dispatcher/pkg/api/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	jobRequestsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "job_requests_total",
		Help: "The total number of requests to dispatch jobs",
	})
	jobRequestsFailuresTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "job_requests_failures_total",
		Help: "The total number of failed dispatch job requests",
	})
	jobRequestsNotFoundFailuresTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "job_requests_not_found_failures_total",
		Help: "The total number of not found dispatch job requests",
	})
	jobRequestsSuccessTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "job_requests_success_total",
		Help: "The total number of success dispatch job requests",
	})
)

func init() {
	metrics.Registry.MustRegister(
		jobRequestsTotal,
		jobRequestsFailuresTotal,
		jobRequestsNotFoundFailuresTotal,
		jobRequestsSuccessTotal,
	)
}

type executeJobHandler struct {
	*Server
}

func (e *executeJobHandler) registerHandler() {
	http.HandleFunc("/execute/", e.handle)
}

func (e *executeJobHandler) handle(w http.ResponseWriter, req *http.Request) {
	jobRequestsTotal.Inc()
	if req.Method != http.MethodPost && req.Method != http.MethodPut {
		jobRequestsFailuresTotal.Inc()
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	ctx := req.Context()
	log := ctrllog.FromContext(ctx)

	name, ns, err := getNameAndNamespace(req.URL.Path, e.defaultNamespace)
	if err != nil {
		jobRequestsFailuresTotal.Inc()
		log.Error(err, "Error getting name and namespace")
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	jt, err := e.getJobTemplate(ns, name, ctx)
	if err != nil {
		jobRequestsFailuresTotal.Inc()
		jobRequestsNotFoundFailuresTotal.Inc()
		log.Error(err, "JobTemplate doesn't exist", "name", name, "namespace", ns)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	log.Info("Creating JobExecution", "name", name, "namespace", ns)
	jobExecution := createJobExecution(jt, req.Body)
	log.Debug("JobExecution payload", "jobExecution", jobExecution)

	if err := e.Create(ctx, jobExecution); err != nil {
		jobRequestsFailuresTotal.Inc()
		log.Error(err, "Error creating JobExecution")
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	jobRequestsSuccessTotal.Inc()
	w.WriteHeader(http.StatusCreated)
}

func getNameAndNamespace(path, defaultNamespace string) (name, namespace string, err error) {
	if n := strings.Split(strings.TrimPrefix(path, "/execute/"), "/"); len(n[0]) == 0 || len(n) > 2 {
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

func createJobExecution(jobTemplate *v1alpha1.JobTemplate, body io.ReadCloser) *v1alpha1.JobExecution {
	var b bytes.Buffer
	if body != nil {
		io.Copy(&b, body)
	}

	return &v1alpha1.JobExecution{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: jobTemplate.ObjectMeta.Name + "-",
			Namespace:    jobTemplate.ObjectMeta.Namespace,
			Labels:       jobTemplate.ObjectMeta.Labels,
		},
		Spec: v1alpha1.JobExecutionSpec{
			JobTemplateName: jobTemplate.ObjectMeta.Name,
			Payload:         b.String(),
		},
	}
}

func (e *executeJobHandler) getJobTemplate(namespace, name string, ctx context.Context) (*v1alpha1.JobTemplate, error) {
	jt := new(v1alpha1.JobTemplate)
	err := e.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, jt)
	return jt, err
}
