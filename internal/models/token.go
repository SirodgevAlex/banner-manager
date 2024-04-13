package models

import (
    "github.com/dgrijalva/jwt-go"
)

type Claims struct {
    UserId int `json:"user_id"`
	IsAdmin bool `json:"is_admin"`
    jwt.StandardClaims
}