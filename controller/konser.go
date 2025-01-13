package controller

import (
	"encoding/json"
	"net/http"

	"project-tiket/config"
	"project-tiket/model"

	// "github.com/gorilla/mux"
)

func AddKonser(w http.ResponseWriter, r *http.Request) {
	var konser model.Konser
	if err := json.NewDecoder(r.Body).Decode(&konser); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `
    INSERT INTO konser (user_id, lokasi_id, nama_konser, tanggal_konser, jumlah_tiket, harga, image, jenis_bank, atas_nama, rekening, status, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
    RETURNING konser_id`

	var id string
	err := config.DB.QueryRow(query, 
		konser.UserID, 
		konser.LokasiID,
		konser.NamaKonser,
		konser.TanggalKonser, 
		konser.JumlahTiket, 
		konser.Harga, 
		konser.Image, 
		konser.JenisBank, 
		konser.AtasNama, 
		konser.Rekening, 
		konser.Status, 
	).Scan(&id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the newly created ID in the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Konser added successfully",
		"id":      id,
	})
}
