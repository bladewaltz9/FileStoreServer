package handler

import "net/http"

// HTTPInterceptor: intercept http requests to check token
func HTTPInterceptor(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		username := r.Form.Get("username")
		token := r.Form.Get("token")

		if !IsTokenValid(username, token) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		h(w, r)
	})
}
