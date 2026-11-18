// Package model defines the domain entities for the Sentinel Device Manager.
package model

import "time"

// User is an operator account that can authenticate against the API.
type User struct {
	UserID       string    `json:"user_id" dynamodbav:"user_id"`
	Username     string    `json:"username" dynamodbav:"username"`
	PasswordHash string    `json:"-" dynamodbav:"password_hash"`
	Role         string    `json:"role" dynamodbav:"role"`
	CreatedAt    time.Time `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" dynamodbav:"updated_at"`
}

// ValidRole reports whether the role is one of the RBAC roles known to the API.
func (u *User) ValidRole() bool {
	switch u.Role {
	case "admin", "operator", "viewer":
		return true
	}
	return false
}
