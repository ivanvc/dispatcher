package template

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Environment struct {
	UUID, ShortUUID, Date, Payload string
}

func newEnvironment(payload string) *Environment {
	id := uuid.New().String()
	shortID := strings.Split(string(id), "-")[0]
	return &Environment{
		UUID:      id,
		ShortUUID: shortID,
		Date:      time.Now().Format("2006-01-02"),
		Payload:   payload,
	}
}
