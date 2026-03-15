package api

import (
	"log/slog"
	"net/http"

	"atthub/internal/attachment"
	"atthub/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(service *attachment.Service, cfg config.Config, logger *slog.Logger) http.Handler {
	handler := &Handler{
		service:        service,
		logger:         logger,
		maxUploadBytes: cfg.MaxUploadBytes,
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", handler.healthz)
	r.Get("/", handler.webApp)
	r.Get("/web", handler.webApp)
	r.Get("/web/attachments", handler.webApp)
	r.Get("/f/{publicID}", handler.openAttachmentByPublicID)

	r.Route("/api/v1", func(apiRouter chi.Router) {
		apiRouter.Post("/attachments/import", handler.importAttachment)
		apiRouter.Get("/attachments", handler.searchAttachments)
		apiRouter.Get("/attachments/{id}", handler.getAttachment)
		apiRouter.Get("/attachments/public/{publicID}", handler.getAttachmentByPublicID)
		apiRouter.Patch("/attachments/{id}", handler.patchAttachment)
		apiRouter.Delete("/attachments/{id}", handler.deleteAttachment)
	})

	return r
}
