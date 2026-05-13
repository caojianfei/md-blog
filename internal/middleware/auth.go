package middleware

import (
    "net/http"

    authSvc "github.com/cybernote/md-blog/internal/service/auth"
)

func AdminOnly(service *authSvc.Service) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ok, _, err := service.CurrentUser(r)
            if err != nil || !ok {
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusUnauthorized)
                _, _ = w.Write([]byte(`{"code":401,"message":"unauthorized"}`))
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
