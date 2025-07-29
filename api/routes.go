// api/routes.go
package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/meghanakudua/print_management_raft3d/raft"
)

func NewRouter(rn *raft.RaftNode) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/join", JoinHandler(rn))

	// Base path
	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/leader", GetLeaderHandler(rn))
		api.Get("/config", GetRaftConfigHandler(rn))

		api.Post("/printers", AddPrinterHandler(rn))
		api.Get("/printers", GetPrintersHandler(rn))

		api.Post("/filaments", AddFilamentHandler(rn))
		api.Get("/filaments", GetFilamentsHandler(rn))

		api.Post("/print_jobs", AddPrintJobHandler(rn))
		api.Get("/print_jobs", GetPrintJobsHandler(rn))

		api.Post("/print_jobs/{jobID}/status", UpdatePrintJobStatusHandler(rn))
	})

	return r
}
