package main


import (
  "log"
  "net/http"
  "project-tiket/config"
  "project-tiket/routes"
)


func init() {
  // Inisialisasi database saat server dimulai
  config.InitDB()
}


func Handler(w http.ResponseWriter, r *http.Request) {
  // Ambil router dari file router.go
  appRouter := routes.SetupRouter()


  // Ambil CORS handler dari file cors.go
  corsHandler := config.SetupCORS(appRouter)


  // Jalankan HTTP handler
  corsHandler.ServeHTTP(w, r)
}


func main() {
  config.InitDB()


  http.HandleFunc("/", Handler)
  log.Println("Starting server on :5000")
  log.Fatal(http.ListenAndServe(":5000", nil))
}
