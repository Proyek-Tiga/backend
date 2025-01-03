package model


import "time"


type User struct {
  UserID    string    `json:"user_id"`    
  RoleID    string    `json:"role_id"`    
  Name      string    `json:"name"`      
  Email     string    `json:"email"`      
  Password  string    `json:"password"`  
  CreatedAt time.Time `json:"created_at"`
  UpdatedAt time.Time `json:"updated_at"`
}
