package controller

import "net/http"

// LiveHandler отвечает за проверку, жив ли процесс приложения.
func LiveHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"alive"}`))
}
