package config


import (
  "github.com/rs/cors"
  "net/http"
)


// SetupCORS configures CORS for the application
func SetupCORS(handler http.Handler) http.Handler {
  c := cors.New(cors.Options{
    AllowedOrigins: []string{"*"},
    AllowedMethods: []string{"GET", "POST", "PUT", "DELETE","OPTIONS"},
    AllowedHeaders: []string{"Content-Type", "Authorization"},
    Debug:          true,
    AllowCredentials: true,
  })


  return c.Handler(handler)
}


