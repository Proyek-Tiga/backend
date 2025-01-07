package controller

import (
	"encoding/json"
	"net/http"

	"project-tiket/config"
	"project-tiket/model"
	// "github.com/gorilla/mux"
)

func AddRequestLokasi(w http.ResponseWriter, r *http.Request) {
    var request model.RequestLokasi

    // Decode JSON body
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    // Begin transaction
    tx, err := config.DB.Begin()
    if err != nil {
        http.Error(w, "Failed to start transaction: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Insert request lokasi and get UUID
    var requestID string
    err = tx.QueryRow(
        `INSERT INTO request_lokasi (user_id, nama_lokasi, kapasitas, status) VALUES ($1, $2, $3, 'pending') RETURNING request_id`,
        request.UserID, request.NamaLokasi, request.Kapasitas,
    ).Scan(&requestID)
    if err != nil {
        tx.Rollback()
        http.Error(w, "Failed to insert request lokasi: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Insert each tiket
    for _, tiket := range request.Tiket {
        _, err := tx.Exec(
            "INSERT INTO tiket (request_id, nama_tiket, harga) VALUES ($1, $2, $3)",
            requestID, tiket.NamaTiket, tiket.Harga,
        )
        if err != nil {
            tx.Rollback()
            http.Error(w, "Failed to insert tiket: "+err.Error(), http.StatusInternalServerError)
            return
        }
    }

    // Commit transaction
    if err := tx.Commit(); err != nil {
        http.Error(w, "Failed to commit transaction: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Send success response
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"message": "Request lokasi dan tiket berhasil ditambahkan"})
}
