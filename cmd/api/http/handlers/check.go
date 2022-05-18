package handlers

import (
	"calc/cmd/api/http/handlers/responses"
	"calc/internal/adapters/db"
	"calc/internal/berrors"
	"calc/internal/services/check"
	"context"
	"net/http"
	"os"
)

type checkGroup struct {
	build string
	db    db.DB
}

func newCheckGroup(build string, db db.DB) *checkGroup {
	return &checkGroup{
		build: build,
		db:    db,
	}
}

// Readiness godoc
// @Tags Check
// @Router /check/readiness [get]
// @Summary Check readiness database
// @Description Check readiness database
// @Produce json
// @Success 200 {object} responses.Health
// @Failure 400 {object} berrors.BusinessError
// @Failure 500
func (cg checkGroup) Readiness(_ *http.Request) (interface{}, error) {
	status := "ok"

	if err := cg.db.StatusCheck(context.Background()); err != nil {
		return nil, berrors.WrapWithError(check.ErrDBNotReady, err)
	}

	return responses.Health{
		Status: status,
	}, nil
}

// Liveness godoc
// @Tags Check
// @Router /check/liveness [get]
// @Summary Check liveness of service
// @Description Check liveness of service
// @Produce json
// @Success 200 {object} responses.Info
// @Failure 400 {object} berrors.BusinessError
// @Failure 500
func (cg checkGroup) Liveness(_ *http.Request) (interface{}, error) {
	host, err := os.Hostname()
	if err != nil {
		host = "unavailable"
	}

	info := responses.Info{
		Status:    "up",
		Build:     cg.build,
		Host:      host,
		Pod:       os.Getenv("KUBERNETES_PODNAME"),
		PodIP:     os.Getenv("KUBERNETES_NAMESPACE_POD_IP"),
		Node:      os.Getenv("KUBERNETES_NODENAME"),
		Namespace: os.Getenv("KUBERNETES_NAMESPACE"),
	}

	return info, nil
}
