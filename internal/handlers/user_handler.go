package handlers

import (
	"banner-manager/internal/models"
	"banner-manager/db"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"
	"unicode"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("secret_key")

func GetUserBannerHandler(w http.ResponseWriter, r *http.Request) {
	tagID := r.URL.Query().Get("tag_id")
	featureID := r.URL.Query().Get("feature_id")
	useLastRevision := r.URL.Query().Get("use_last_revision")

	userToken := r.Header.Get("token")

	if tagID == "" || featureID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if userToken == "" || !isValidUserToken(userToken) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	banner, err := getBannerForUser(tagID, featureID, useLastRevision)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to get banner: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(banner)
}

func isValidUserToken(token string) bool {
	return token == "user_token"
}

func getBannerForUser(tagID, featureID, useLastRevision string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"title": "Example Banner",
		"text":  "This is an example banner.",
		"url":   "https://example.com/banner",
	}, nil
}

func isEmailValid(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func isPasswordSafe(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

func Register(w http.ResponseWriter, r *http.Request) {
	db, err := db.GetPostgresDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var user models.User
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE Email = $1", user.Email).Scan(&count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.Error(w, "Email уже занят", http.StatusConflict)
		return
	}

	if isEmailValid(user.Email) && isPasswordSafe(user.Password) {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		query := "INSERT INTO users(Email, Password, IsAdmin) VALUES($1, $2, $3) RETURNING Id"
		err = db.QueryRow(query, user.Email, string(hashedPassword), user.IsAdmin).Scan(&user.Id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func Authorize(w http.ResponseWriter, r *http.Request) {
	database, err := db.GetPostgresDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var hashedPassword string
	err = database.QueryRow("SELECT Password FROM Users WHERE Email = $1", user.Email).Scan(&hashedPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(user.Password)); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}


	var UserId int
	err = database.QueryRow("SELECT Id FROM users WHERE Email = $1", user.Email).Scan(&UserId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var IsAdmin bool
	err = database.QueryRow("SELECT IsAdmin FROM users WHERE Email = $1", user.Email).Scan(&IsAdmin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &models.Claims{
		UserId: UserId,
		IsAdmin: IsAdmin,
		StandardClaims: jwt.StandardClaims{
			Subject:   strconv.Itoa(UserId),
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Ошибка генерации токена", http.StatusInternalServerError)
		return
	}

	err = db.AddRedisToken(tokenString, 5 * time.Minute)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}
