package main

import (
	"encoding/json"
	"net/http"
)

const dummyCertificateSerial = "1"
const dummyCertificatePEM = `-----BEGIN CERTIFICATE-----
MIIC3DCCAcSgAwIBAgIBATANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQDEwtkdW1t
eS5sb2NhbDAeFw0yNTExMTYyMDQ0MjJaFw0yNjExMTYyMTQ0MjJaMBYxFDASBgNV
BAMTC2R1bW15LmxvY2FsMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA
3AQdtoqg/gsuz1SC7Q4Ej435Ru8jGz6VbtLbQBUzVQ8N+UxYX3OhYhpFgID4XGKw
8Lcq1ZnsgMGW1lTD+V1icwiUxpPQGTHXyj4Y0j8ZNbD581Nl5cdU+1idljk0bXaG
Uv2PJ7IgI70inUXRfIC3iaTODUI9deInCp/OJbxLaUD2xYoc4cTEEcNnhZ6VDICA
X2hT5nfEVSoE1iupQjwHhDEAWJ+1nr6KAmYHUn5imrJOJtS+wOl+qjUD2ytcPCjU
9zKDl6QuftfiQaV7nv9HJQGfN6gypnErk/aZ1FfuJEoxllmjH5yyGEcRAXUBN7pb
o8QmjNhCBy3Y6CxCA3BDYwIDAQABozUwMzAOBgNVHQ8BAf8EBAMCB4AwEwYDVR0l
BAwwCgYIKwYBBQUHAwIwDAYDVR0TAQH/BAIwADANBgkqhkiG9w0BAQsFAAOCAQEA
CgGD5HyPork/2Tlol5jw63fUqFQmT70pDTwk9CFYjcNFloKimCGKI/ZwSl9hwmai
h1MLzJ+XxzMQS9WYAVddZNQ8Odz0URv4RccnyMWdonF/bqC4Roo6Yg1/2kXBX/Ab
Bu0HvVxEl2A3R3hPxxlCHk5E2etUX7ypASSpJdC7suKYfnVrLpBGJvYTAlynjDUV
7GC6lyggtK9eLYrFNGGzIJorlcldgzEMjokE8+lxG44CKxzribD4X+dMO010OizK
D4PQdtDGBYKwgZfRNwbffscpgVL8jIYJYP/TZ35gwYKjRmOWozTAXzT679GtOuGf
VaUbNVhtYshJXaMN0aTS/w==
-----END CERTIFICATE-----
`

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
		_ = json.NewEncoder(w).Encode(map[string]string{"crt": dummyCertificatePEM})
	})

	http.HandleFunc("/certificates/"+dummyCertificateSerial, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(dummyCertificatePEM))
	})

	http.ListenAndServe(":8080", nil)
}
