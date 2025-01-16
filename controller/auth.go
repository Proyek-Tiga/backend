package controller

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"project-tiket/config"
	"project-tiket/model"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)


var jwtKey = []byte("your_secret_key")


func Register(w http.ResponseWriter, r *http.Request) {
  var user model.User


  // Decode request body
  err := json.NewDecoder(r.Body).Decode(&user)
  if err != nil {
    http.Error(w, "Invalid request payload", http.StatusBadRequest)
    return
  }


  // Periksa apakah email sudah ada di database
  var existingUserID string
  err = config.DB.QueryRow("SELECT user_id FROM users WHERE email = $1", user.Email).Scan(&existingUserID)
  if err != nil && err != sql.ErrNoRows {
    http.Error(w, "Internal server error", http.StatusInternalServerError)
    return
  }
  if existingUserID != "" {
    http.Error(w, "Email already exists", http.StatusBadRequest)
    return
  }


  // Hash password
  hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
  if err != nil {
    http.Error(w, "Failed to hash password", http.StatusInternalServerError)
    return
  }


  // Simpan pengguna baru ke database
  _, err = config.DB.Exec(
    `INSERT INTO users (user_id, role_id, name, email, password, created_at, updated_at)
         VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW(), NOW())`,
    user.RoleID, user.Name, user.Email, hashedPassword,
  )
  if err != nil {
    http.Error(w, "Failed to insert user into database", http.StatusInternalServerError)
    return
  }


  // Berikan respon sukses
  w.Header().Set("Content-Type", "application/json")
  response := map[string]interface{}{
    "message": "User registered successfully",
  }
  err = json.NewEncoder(w).Encode(response)
  if err != nil {
    log.Printf("Error encoding response: %v", err)
    http.Error(w, "Internal server error", http.StatusInternalServerError)
  }
}

func Login(w http.ResponseWriter, r *http.Request) {
  var user model.User
  var hashedPassword string
  var roleID string


  // Decode request body
  err := json.NewDecoder(r.Body).Decode(&user)
  if err != nil {
      http.Error(w, "Invalid request payload", http.StatusBadRequest)
      return
  }


  // Query database untuk mendapatkan user_id, email, hashed password, dan role_id
  err = config.DB.QueryRow(
      "SELECT user_id, email, password, role_id FROM users WHERE email=$1",
      user.Email,
  ).Scan(&user.UserID, &user.Email, &hashedPassword, &roleID)
  if err != nil {
      if err == sql.ErrNoRows {
          http.Error(w, "User not found", http.StatusUnauthorized)
          return
      }
      http.Error(w, "Internal server error", http.StatusInternalServerError)
      return
  }


  // Bandingkan password mentah dengan hashed password dari database
  err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(user.Password))
  if err != nil {
      http.Error(w, "Invalid password", http.StatusUnauthorized)
      return
  }


  // Generate JWT token
  expirationTime := time.Now().Add(60 * time.Minute)
  claims := &Claims{
      Email: user.Email,
      UserID: user.UserID,
      StandardClaims: jwt.StandardClaims{
          ExpiresAt: expirationTime.Unix(),
      },
  }


  token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
  tokenString, err := token.SignedString(jwtKey)
  if err != nil {
      http.Error(w, "Internal server error", http.StatusInternalServerError)
      return
  }


  // Return response dengan token dan role_id
  w.Header().Set("Content-Type", "application/json")
  response := map[string]interface{}{
      "message": "Login successful",
      "token":   tokenString,
      "role_id": roleID,
  }
  err = json.NewEncoder(w).Encode(response)
  if err != nil {
      log.Printf("Error encoding response: %v", err)
      http.Error(w, "Internal server error", http.StatusInternalServerError)
  }
}



type Claims struct {
  Email string `json:"email"`
  UserID string `json:"user_id"`
  jwt.StandardClaims
}

func ValidateToken(tokenString string) (bool, error) {
  token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    return jwtKey, nil
  })


  if err != nil {
    return false, err
  }


  return token.Valid, nil
}


func JWTAuth(next http.HandlerFunc) http.HandlerFunc {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    bearerToken := r.Header.Get("Authorization")
    sttArr := strings.Split(bearerToken, " ")
    if len(sttArr) == 2 {
      isValid, _ := ValidateToken(sttArr[1])
      if isValid {
        next.ServeHTTP(w, r)
      } else {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
      }
    } else {
      http.Error(w, "Unauthorized", http.StatusUnauthorized)
    }
  })
}

// GetUserByToken retrieves the logged-in user's data
func GetUser(w http.ResponseWriter, r *http.Request) {
  // Get token from Authorization header
  authHeader := r.Header.Get("Authorization")
  if authHeader == "" {
    http.Error(w, "Authorization header is missing", http.StatusUnauthorized)
    return
  }


  // Extract token from "Bearer <token>"
  tokenString := strings.TrimPrefix(authHeader, "Bearer ")


  // Parse and validate token
  claims := &Claims{}
  token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
    return jwtKey, nil
  })
  if err != nil || !token.Valid {
    http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
    return
  }


  // Get user data from database
  var user model.User
  err = config.DB.QueryRow(
    "SELECT user_id, name, email, role_id, created_at FROM users WHERE email = $1",
    claims.Email,
  ).Scan(&user.UserID, &user.Name, &user.Email, &user.RoleID, &user.CreatedAt)
  if err != nil {
    if err == sql.ErrNoRows {
      http.Error(w, "User not found", http.StatusNotFound)
      return
    }
    http.Error(w, "Internal server error: "+err.Error(), http.StatusInternalServerError)
    return
  }


  // Return user data as JSON
  w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(user)
}
