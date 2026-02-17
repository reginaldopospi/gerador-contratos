package middleware

import (
	"log"
	"net/http"

	"gerador-contratos/backend/internal/http/httpx"
	"gerador-contratos/backend/internal/modules/common"
)

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic recovered: %v", rec)
				httpx.WriteError(w, common.NewInternal("panic", "internal server error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
