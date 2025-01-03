package controller

import (
	"encoding/json"
	"net/http"

	"project-tiket/config"
	"project-tiket/model"

	"github.com/gorilla/mux"
)

func AddLokasi(w http.ResponseWriter, r *http.Request) {
	var lokasi model.Lokasi
	if err := json.NewDecoder(r.Body).Decode(&lokasi); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `
    INSERT INTO lokasi (lokasi, tiket ,created_at, updated_at)
    VALUES ($1, $2,NOW(), NOW())
    RETURNING lokasi_id`

	var id string
	err := config.DB.QueryRow(query, lokasi.Lokasi, lokasi.Tiket).Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the newly created ID in the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Lokasi added successfully",
		"id":      id,
	})
}
func GetLokasi(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query("SELECT * FROM lokasi")
	if err != nil {
	  http.Error(w, err.Error(), http.StatusInternalServerError)
	  return
	}
  
  
	defer rows.Close()
  
  
	var lokasi []model.Lokasi
	for rows.Next() {
	  var lokasis model.Lokasi
	  if err := rows.Scan(&lokasis.LokasiID, &lokasis.Lokasi, &lokasis.Tiket, &lokasis.CreatedAt, &lokasis.UpdatedAt); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	  }
	  lokasi = append(lokasi, lokasis)
	}
  
  
	if err := rows.Err(); err != nil {
	  http.Error(w, err.Error(), http.StatusInternalServerError)
	  return
	}
  
  
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lokasi)
  }

  func GetLokasiByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
	  http.Error(w, "ID not provided", http.StatusBadRequest)
	  return
	}
  
  
	// Query database untuk mendapatkan data lokasi berdasarkan UUID
	var lokasis model.Lokasi
	err := config.DB.QueryRow(
	  "SELECT lokasi_id, lokasi, tiket, created_at, updated_at FROM lokasi WHERE lokasi_id = $1",
	  id,
	).Scan(&lokasis.LokasiID, &lokasis.Lokasi, &lokasis.Tiket, &lokasis.CreatedAt, &lokasis.UpdatedAt)
	if err != nil {
	  http.Error(w, "Internal server error: "+err.Error(), http.StatusInternalServerError)
	  return
	}
  
  
	// Berikan respons dalam format JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(lokasis); err != nil {
	  http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
  }

  func UpdateLokasi(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
	  http.Error(w, "ID not provided", http.StatusBadRequest)
	  return
	}
  
  
	var lokasi model.Lokasi
	if err := json.NewDecoder(r.Body).Decode(&lokasi); err != nil {
	  http.Error(w, err.Error(), http.StatusBadRequest)
	  return
	}
  
  
	query := `
	  UPDATE lokasi
	  SET lokasi=$1, tiket=$2, updated_at=NOW()
	  WHERE lokasi_id=$3`
  
  
	result, err := config.DB.Exec(query, lokasi.Lokasi, lokasi.Tiket, idStr)
	if err != nil {
	  http.Error(w, err.Error(), http.StatusInternalServerError)
	  return
	}
  
  
	rowsAffected, err := result.RowsAffected()
	if err != nil {
	  http.Error(w, err.Error(), http.StatusInternalServerError)
	  return
	}
  
  
	if rowsAffected == 0 {
	  http.Error(w, "No rows were updated", http.StatusNotFound)
	  return
	}
  
  
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
	  "message": "Lokasi updated successfully",
	})
  }
  func DeleteLokasi(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
	  http.Error(w, "ID not provided", http.StatusBadRequest)
	  return
	}
  
  
	query := `
	  DELETE FROM lokasi
	  WHERE lokasi_id=$1`
  
  
	// Execute the SQL statement
	result, err := config.DB.Exec(query, idStr)
	if err != nil {
	  http.Error(w, err.Error(), http.StatusInternalServerError)
	  return
	}
  
  
	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
	  http.Error(w, err.Error(), http.StatusInternalServerError)
	  return
	}
  
  
	if rowsAffected == 0 {
	  http.Error(w, "No rows were deleted", http.StatusNotFound)
	  return
	}
  
  
	// Return the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
	  "message": "Lokasi deleted successfully",
	})
  }
	
  