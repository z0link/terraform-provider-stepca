package main

import (
	"encoding/json"
	"net/http"
)

func main() {
	http.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("1.2.3"))
	})

	http.HandleFunc("/sign", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var req struct {
			CSR string `json:"csr"`
			OTT string `json:"ott"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		_ = json.NewEncoder(w).Encode(map[string]string{"crt": "dummy-certificate"})
	})

	http.ListenAndServe(":8080", nil)
}
