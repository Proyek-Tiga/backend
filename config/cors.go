package config


import (
  "github.com/rs/cors"
  "net/http"
)


// SetupCORS configures CORS for the application
func SetupCORS(handler http.Handler) http.Handler {
  c := cors.New(cors.Options{
    AllowedOrigins: []string{"http://127.0.0.1:5500","https://proyek-tiga.github.io/"},
    AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders: []string{"Content-Type", "Authorization"},
    Debug:          true,
  })


  return c.Handler(handler)
}


