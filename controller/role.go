package controller

import (
	"encoding/json"
	"net/http"

	"project-tiket/config"
	"project-tiket/model"

	"github.com/gorilla/mux"
	// "github.com/gorilla/mux"
)

func AddRole(w http.ResponseWriter, r *http.Request) {
	var role model.Role
	if err := json.NewDecoder(r.Body).Decode(&role); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `
    INSERT INTO role (role_name,created_at, updated_at)
    VALUES ($1,NOW(), NOW())
    RETURNING role_id`

	var id string
	err := config.DB.QueryRow(query, role.RoleName).Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the newly created ID in the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Role added successfully",
		"id":      id,
	})
}

func GetRole(w http.ResponseWriter, r *http.Request) {
	rows, err := config.DB.Query("SELECT * FROM role")
	if err != nil {
	  http.Error(w, err.Error(), http.StatusInternalServerError)
	  return
	}
  
  
	defer rows.Close()
  
  
	var role []model.Role
	for rows.Next() {
	  var roles model.Role
	  if err := rows.Scan(&roles.RoleID, &roles.RoleName, &roles.CreatedAt, &roles.UpdatedAt); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	  }
	  role = append(role, roles)
	}
  
  
	if err := rows.Err(); err != nil {
	  http.Error(w, err.Error(), http.StatusInternalServerError)
	  return
	}
  
  
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(role)
  }

  func GetRoleByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
	  http.Error(w, "ID not provided", http.StatusBadRequest)
	  return
	}
  
  
	// Query database untuk mendapatkan data lokasi berdasarkan UUID
	var roles model.Role
	err := config.DB.QueryRow(
	  "SELECT role_id, role_name, created_at, updated_at FROM role WHERE role_id = $1",
	  id,
	).Scan(&roles.RoleID, &roles.RoleName, &roles.CreatedAt, &roles.UpdatedAt)
	if err != nil {
	  http.Error(w, "Internal server error: "+err.Error(), http.StatusInternalServerError)
	  return
	}
  
  
	// Berikan respons dalam format JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(roles); err != nil {
	  http.Error(w, "Failed to encode response: "+err.Error(), http.StatusInternalServerError)
	}
  }

  func UpdateRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
	  http.Error(w, "ID not provided", http.StatusBadRequest)
	  return
	}
  
  
	var role model.Role
	if err := json.NewDecoder(r.Body).Decode(&role); err != nil {
	  http.Error(w, err.Error(), http.StatusBadRequest)
	  return
	}
  
  
	query := `
	  UPDATE role
	  SET role_name=$1,updated_at=NOW()
	  WHERE role_id=$2`
  
  
	result, err := config.DB.Exec(query, role.RoleName, idStr)
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
	  "message": "role updated successfully",
	})
  }

  func DeleteRole(w http.ResponseWriter, r *http.Request) {
	// Extract ID from URL
	vars := mux.Vars(r)
	idStr, ok := vars["id"]
	if !ok {
	  http.Error(w, "ID not provided", http.StatusBadRequest)
	  return
	}
  
  
	query := `
	  DELETE FROM role
	  WHERE role_id=$1`
  
  
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
	  "message": "Role deleted successfully",
	})
  }
	
  