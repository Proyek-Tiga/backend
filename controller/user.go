package controller

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"project-tiket/config"
	"project-tiket/model"

	"github.com/gorilla/mux"
)

func GetUsersByRole(w http.ResponseWriter, r *http.Request) {
	// Ambil role_name dari query parameter
	roleName := r.URL.Query().Get("role_name")
	if roleName == "" {
		http.Error(w, "role_name is required", http.StatusBadRequest)
		return
	}

	// Query ke database dengan PostgreSQL-style placeholder
	query := `
		SELECT u.user_id, u.name, u.email
		FROM users u
		JOIN role r ON u.role_id = r.role_id
		WHERE r.role_name = $1`
	rows, err := config.DB.Query(query, roleName)
	if err != nil {
		http.Error(w, "Error querying database: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Parsing hasil query
	var users []model.UserResponse
	for rows.Next() {
		var user model.UserResponse
		if err := rows.Scan(&user.UserID, &user.Name, &user.Email); err != nil {
			http.Error(w, "Error scanning data: "+err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	// Kirim response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Perbaikan query dengan PostgreSQL-style placeholder
	query := `
		INSERT INTO users (user_id, role_id, name, email, password, created_at, updated_at) 
		VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW(), NOW())`
	_, err := config.DB.Exec(query, user.RoleID, user.Name, user.Email, user.Password)
	if err != nil {
		http.Error(w, "Error inserting user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Berikan respons sukses
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})
}

func GetUserById(w http.ResponseWriter, r *http.Request) {
	// Ambil ID dari parameter URL
	vars := mux.Vars(r)
	id := vars["id"]

	// Query ke database
	query := `
		SELECT user_id, role_id, name, email, created_at, updated_at
		FROM users
		WHERE user_id = $1`
	row := config.DB.QueryRow(query, id)

	// Deklarasi variabel untuk menampung data user
	var user model.User
	err := row.Scan(&user.UserID, &user.RoleID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error querying database: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Kirim response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var user model.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	query := `UPDATE users SET role_id = $1, name = $2, email = $3, updated_at = NOW() WHERE user_id = $4`
	_, err := config.DB.Exec(query, user.RoleID, user.Name, user.Email, id)
	if err != nil {
		http.Error(w, "Error updating user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User updated successfully"})
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	query := `DELETE FROM users WHERE user_id = $1`
	_, err := config.DB.Exec(query, id)
	if err != nil {
		http.Error(w, "Error deleting user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "User deleted successfully"})
}
