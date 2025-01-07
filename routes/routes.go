package routes

import (
	"project-tiket/controller"

	"github.com/gorilla/mux"
)

// SetupRouter initializes the application's routes
func SetupRouter() *mux.Router {
	router := mux.NewRouter()

	//AUTH
	router.HandleFunc("/auth/login", controller.Login).Methods("POST")
	router.HandleFunc("/auth/regis", controller.Register).Methods("POST")

	//LOKASI
	router.HandleFunc("/api/lokasi", controller.JWTAuth(controller.AddLokasi)).Methods("POST")
	router.HandleFunc("/api/lokasi", controller.GetLokasi).Methods("GET")
	router.HandleFunc("/api/lokasi/{id}", controller.GetLokasiByID).Methods("GET")
	router.HandleFunc("/api/lokasi/{id}", controller.JWTAuth(controller.UpdateLokasi)).Methods("PUT")
	router.HandleFunc("/api/lokasi/{id}", controller.JWTAuth(controller.DeleteLokasi)).Methods("DELETE")

	//ROLE
	router.HandleFunc("/api/role", controller.JWTAuth(controller.AddRole)).Methods("POST")
	router.HandleFunc("/api/role", controller.GetRole).Methods("GET")
	router.HandleFunc("/api/role/{id}", controller.GetRoleByID).Methods("GET")
	router.HandleFunc("/api/role/{id}", controller.JWTAuth(controller.UpdateRole)).Methods("PUT")
	router.HandleFunc("/api/role/{id}", controller.JWTAuth(controller.DeleteRole)).Methods("DELETE")

	//USER
	router.HandleFunc("/api/profile", controller.JWTAuth(controller.GetUser)).Methods("GET")

	//TIKET
	router.HandleFunc("/api/tiket", controller.JWTAuth(controller.AddTiket)).Methods("POST")
	router.HandleFunc("/api/tiket", controller.GetTiket).Methods("GET")
	router.HandleFunc("/api/tiket/{id}", controller.GetTiketByID).Methods("GET")
	router.HandleFunc("/api/tiket/{id}", controller.JWTAuth(controller.UpdateTiket)).Methods("PUT")
	router.HandleFunc("/api/tiket/{id}", controller.JWTAuth(controller.DeleteTiket)).Methods("DELETE")

	// Request
	router.HandleFunc("/api/request", controller.JWTAuth(controller.AddRequestLokasi)).Methods("POST")

	return router
}
