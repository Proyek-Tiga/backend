package controller

import (
	"encoding/json"
	"net/http"

	"project-tiket/config"
	"project-tiket/model"
	"github.com/gorilla/mux"
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
        `INSERT INTO request_lokasi (user_id, nama_lokasi, kapasitas, status, created_at, updated_at) 
         VALUES ($1, $2, $3, 'pending', NOW(), NOW()) RETURNING request_id`,
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
            "INSERT INTO tiket (nama_tiket, harga, jumlah_tiket, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW())",
            tiket.NamaTiket, tiket.Harga, tiket.JumlahTiket,
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

func UpdateRequestStatus(w http.ResponseWriter, r *http.Request) {
    id := mux.Vars(r)["id"]
    var status struct {
        Status string `json:"status"`
    }

    // Decode JSON body untuk mendapatkan status baru
    if err := json.NewDecoder(r.Body).Decode(&status); err != nil {
        http.Error(w, "Invalid request payload", http.StatusBadRequest)
        return
    }

    // Pastikan status yang diberikan valid
    if status.Status != "pending" && status.Status != "approved" && status.Status != "rejected" {
        http.Error(w, "Invalid status", http.StatusBadRequest)
        return
    }

    // Mulai transaksi untuk memastikan atomicity
    tx, err := config.DB.Begin()
    if err != nil {
        http.Error(w, "Failed to start transaction: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Update status request
    query := `UPDATE request_lokasi SET status = $1 WHERE request_id = $2`
    _, err = tx.Exec(query, status.Status, id)
    if err != nil {
        tx.Rollback()
        http.Error(w, "Failed to update status: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Jika status yang diberikan adalah "approved", tambahkan lokasi ke tabel lokasii
    if status.Status == "approved" {
        // Ambil data request lokasi
        var request model.RequestLokasi
        err := tx.QueryRow(`SELECT user_id, nama_lokasi, kapasitas FROM request_lokasi WHERE request_id = $1`, id).Scan(&request.UserID, &request.NamaLokasi, &request.Kapasitas)
        if err != nil {
            tx.Rollback()
            http.Error(w, "Failed to get request data: "+err.Error(), http.StatusInternalServerError)
            return
        }

        // Insert lokasi ke tabel lokasii
        _, err = tx.Exec(
            `INSERT INTO lokasii (user_id, nama_lokasi, kapasitas) VALUES ($1, $2, $3)`,
            request.UserID, request.NamaLokasi, request.Kapasitas,
        )
        if err != nil {
            tx.Rollback()
            http.Error(w, "Failed to insert location: "+err.Error(), http.StatusInternalServerError)
            return
        }
    }

    // Commit transaksi jika semuanya berjalan lancar
    if err := tx.Commit(); err != nil {
        http.Error(w, "Failed to commit transaction: "+err.Error(), http.StatusInternalServerError)
        return
    }

    // Send success response
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"message": "Request status updated successfully"})
}

