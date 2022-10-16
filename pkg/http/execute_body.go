package http

import (
	"encoding/json"
	"net/http"

	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

const defaultNamespace = "default"

type executeBody struct {
	Namespace string      `json:"namespace"`
	Args      interface{} `json:"args"`
}

func parseBody(req *http.Request) *executeBody {
	var e *executeBody
	if err := json.NewDecoder(req.Body).Decode(&e); err != nil {
		log := ctrllog.FromContext(req.Context())
		log.Error(err, "Error parsing execution body")
		return nil
	}
	return e
}

func (e *executeBody) getNamespace() string {
	if len(e.Namespace) == 0 {
		return defaultNamespace
	}
	return e.Namespace
}

func (e *executeBody) getArgs() string {
	a, _ := json.Marshal(e.Args)
	return string(a)
}
