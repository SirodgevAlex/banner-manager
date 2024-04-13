package handlers

import (
	"banner-manager/db"
	"banner-manager/internal/models"
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

	tokenExists, err := db.IsRedisTokenExists(userToken)
	if err != nil {
		http.Error(w, "Ошибка проверки токена", http.StatusInternalServerError)
		return
	}

	if userToken == "" || !tokenExists {
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
	database, err := db.GetPostgresDB()
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
	fmt.Println(user)

	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", user.Email).Scan(&count)
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

		query := "INSERT INTO users(email, password, is_admin) VALUES($1, $2, $3) RETURNING id"
		err = database.QueryRow(query, user.Email, string(hashedPassword), user.IsAdmin).Scan(&user.ID)
		if err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	fmt.Println(user)

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
	err = database.QueryRow("SELECT password FROM users WHERE email = $1", user.Email).Scan(&hashedPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(user.Password)); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var UserID int
	err = database.QueryRow("SELECT id FROM users WHERE email = $1", user.Email).Scan(&UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var IsAdmin bool
	err = database.QueryRow("SELECT is_admin FROM users WHERE email = $1", user.Email).Scan(&IsAdmin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &models.Claims{
		UserID:  UserID,
		IsAdmin: IsAdmin,
		StandardClaims: jwt.StandardClaims{
			Subject:   strconv.Itoa(UserID),
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Ошибка генерации токена", http.StatusInternalServerError)
		return
	}

	err = db.AddRedisToken(tokenString, 5*time.Minute)
	fmt.Println(err)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func CreateBannerHandler(w http.ResponseWriter, r *http.Request) {
	var banner models.Banner
	err := json.NewDecoder(r.Body).Decode(&banner)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	adminToken := r.Header.Get("token")

	isAdmin, err := IsAdminTokenValid(adminToken)
	if err != nil {
		http.Error(w, "Ошибка при проверке токена", http.StatusInternalServerError)
		return
	}
	if !isAdmin {
		http.Error(w, "Недостаточно прав", http.StatusForbidden)
		return
	}

	bannerID, err := CreateBanner(banner)
	if err != nil {
		http.Error(w, "Ошибка при создании баннера", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"banner_id": bannerID})
}

func CreateBanner(banner models.Banner) (int, error) {
	database, err := db.GetPostgresDB()
	if err != nil {
		return 0, err
	}

	createdAt := time.Now()
	updatedAt := createdAt

	query := `
		INSERT INTO banners (title, text, url, feature_id, tag_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	var id int
	err = database.QueryRow(query, banner.Title, banner.Text, banner.URL, banner.FeatureID, banner.TagID, banner.IsActive, createdAt, updatedAt).Scan(&id)
	if err != nil {
		return 0, err
	}

	banner.ID = id
	banner.CreatedAt = createdAt
	banner.UpdatedAt = updatedAt

	return id, nil
}

func IsAdminTokenValid(tokenString string) (bool, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return false, err
	}

	if claims, ok := token.Claims.(*models.Claims); ok && token.Valid {
		return claims.IsAdmin, nil
	}

	return false, err
}
