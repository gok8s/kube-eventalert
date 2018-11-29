package api


import (
"net/http"
"io"
)

func Healthz(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "OK")
}

