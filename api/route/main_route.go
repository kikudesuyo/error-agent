// SPDX-License-Identifier: Apache-2.0
package route

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kikudesuyo/error-agent/api/handler/v1"
	"github.com/kikudesuyo/error-agent/api/library"
	"github.com/kikudesuyo/error-agent/api/middleware"
	"github.com/kikudesuyo/error-agent/api/repository"
	"github.com/kikudesuyo/error-agent/api/service"
	"gorm.io/gorm"
)

func NewMainRoute(db *gorm.DB, ingestToken string) http.Handler {
	router := chi.NewRouter()
	errorEventRepository := repository.NewErrorEventPostgresRepository(db)
	errorEventService := service.NewErrorEventService(errorEventRepository)
	errorEventHandler := v1.NewErrorEventHandler(errorEventService)
	router.Get("/healthz", func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})
	router.Get("/readyz", func(writer http.ResponseWriter, _ *http.Request) {
		if err := library.PingDatabase(db); err != nil {
			http.Error(writer, "not ready", http.StatusServiceUnavailable)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
	router.With(func(next http.Handler) http.Handler {
		return middleware.IngestAuth(ingestToken, next)
	}).Post("/v1/error-events", errorEventHandler.HandlePostErrorEvent)
	return router
}
