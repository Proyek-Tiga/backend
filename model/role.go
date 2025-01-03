package model

import "time"

type Role struct {
  RoleID string `json:"role_id"`
  RoleName   string `json:"role_name"`
  CreatedAt time.Time `json:"created_at"`
  UpdatedAt time.Time `json:"updated_at"`
}
