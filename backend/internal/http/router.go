package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	custommw "gerador-contratos/backend/internal/http/middleware"
	"gerador-contratos/backend/internal/http/handlers"
)

type RouterDependencies struct {
	AuthHandler      *handlers.AuthHandler
	ContractsHandler *handlers.ContractsHandler
	BrokersHandler   *handlers.BrokersHandler
	ClausesHandler   *handlers.ClausesHandler
	AuthValidator    custommw.AuthValidator
}

func NewRouter(deps RouterDependencies) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(custommw.Recover)
	r.Use(custommw.CORS)

	requireAuth := custommw.RequireAuth(deps.AuthValidator)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api/v1", func(api chi.Router) {
		api.Route("/auth", func(ar chi.Router) {
			deps.AuthHandler.RegisterRoutes(ar, requireAuth)
		})
		api.Route("/contracts", func(cr chi.Router) {
			deps.ContractsHandler.RegisterRoutes(cr, requireAuth)
		})
		api.Route("/brokers", func(br chi.Router) {
			deps.BrokersHandler.RegisterRoutes(br, requireAuth)
		})
		api.Route("/clauses", func(cl chi.Router) {
			deps.ClausesHandler.RegisterRoutes(cl, requireAuth)
		})
	})

	return r
}

