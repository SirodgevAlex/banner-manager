package handlers

import (
	"banner-manager/db"
	"banner-manager/internal/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"regexp"
	"strconv"
	"time"
	"unicode"
	_ "strings"
)

var jwtKey = []byte("secret_key")

func GetUserBannerHandler(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		TagID           string `json:"tag_id"`
		FeatureID       string `json:"feature_id"`
		UseLastRevision bool   `json:"use_last_revision"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	tagID := requestData.TagID
	featureID := requestData.FeatureID
	useLastRevision := requestData.UseLastRevision

	userToken := r.Header.Get("token")

	if tagID == "" || featureID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// tokenExists, err := db.IsRedisTokenExists(userToken)
	// if err != nil {
	// 	http.Error(w, "Ошибка проверки токена", http.StatusInternalServerError)
	// 	return
	// }
	tokenExists := true

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

	response := map[string]string{
		"title": banner.Title,
		"text":  banner.Text,
		"url":   banner.URL,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func getBannerForUser(tagID, featureID string, useLastRevision bool) (*models.Banner, error) {
	var banner *models.Banner
	var err error

	if useLastRevision {
		banner, err = getLastBannerRevisionFromDB(tagID, featureID)
		if err != nil {
			return nil, err
		}
	} else {
		banner, err = db.GetBannerFromRedis(tagID, featureID)
		if err != nil {
			banner, err = getLastBannerRevisionFromDB(tagID, featureID)
			if err != nil {
				return nil, err
			}
			err := db.CacheBannerInRedis(tagID, featureID, banner)
			if err != nil {
				fmt.Println("Ошибка при сохранении баннера в кэше Redis:", err)
			}
		}
	}

	return banner, nil
}

func getLastBannerRevisionFromDB(tagID, featureID string) (*models.Banner, error) {
	database, err := db.GetPostgresDB()
	if err != nil {
		return nil, err
	}

	var title, text, url string
	err = database.QueryRow("SELECT title, text, url FROM banners WHERE tag_id = $1 AND feature_id = $2 ORDER BY updated_at DESC LIMIT 1", tagID, featureID).Scan(&title, &text, &url)
	if err != nil {
		return nil, err
	}

	banner := &models.Banner{
		Title: title,
		Text:  text,
		URL:   url,
	}

	return banner, nil
}

func isUserTokenValid(token string) (bool, error) { // todo
	isUser, err := IsAdminTokenValid(token)
	if err != nil {
		return isUser, err
	}

	isUser = !isUser
	return isUser, nil
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

	err = db.AddRedisToken(tokenString, 5*time.Second)
	if (err != nil) {
		
	}

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
	    http.Error(w, "Unauthorized", http.StatusUnauthorized)
	    return
	}

	if !isAdmin {
		http.Error(w, "Forbidden", http.StatusForbidden)
	    return
	}

	bannerID, err := CreateBanner(banner)
	if err != nil {
		fmt.Println(err)
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

func GetBannersByFeatureOrTagHandler(w http.ResponseWriter, r *http.Request) {
	adminToken := r.Header.Get("token")
	isAdmin, err := IsAdminTokenValid(adminToken)

	if err != nil {
	    http.Error(w, "Unauthorized", http.StatusUnauthorized)
	    return
	}

	if !isAdmin {
		http.Error(w, "Forbidden", http.StatusForbidden)
	    return
	}

	var req struct {
		FeatureID int `json:"feature_id,omitempty"`
		TagID     int `json:"tag_id,omitempty"`
		Limit     int `json:"limit,omitempty"`
		Offset    int `json:"offset,omitempty"`
	}

	fmt.Println(r.Body)

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON format: %v", err), http.StatusBadRequest)
		return
	}

	if req.FeatureID == 0 && req.TagID == 0 {
		http.Error(w, "Feature ID or Tag ID must be provided", http.StatusBadRequest)
		return
	}

	banners, err := GetBannersByFeatureOrTag(req.FeatureID, req.TagID)
	if err != nil {
		http.Error(w, "Failed to get banners", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(banners)
}

func GetBannersByFeatureOrTag(featureID, tagID int) ([]models.Banner, error) {
	database, err := db.GetPostgresDB()
	if err != nil {
		return nil, err
	}

	var banners []models.Banner

	query := "SELECT * FROM banners WHERE true"
	if featureID != 0 {
		query += " AND feature_id = $1"
	}
	if tagID != 0 {
		query += " AND tag_id = $2"
	}

	fmt.Println(query)

	var rows *sql.Rows

	if tagID != 0 && featureID != 0 {
		rows, err = database.Query(query, featureID, tagID)
	} else if featureID != 0 {
		rows, err = database.Query(query, featureID)
	} else if tagID != 0 {
		rows, err = database.Query(query, tagID)
	} else {
		rows, err = database.Query(query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var banner models.Banner
		fmt.Println("ya")
		if err := rows.Scan(&banner.ID, &banner.FeatureID, &banner.TagID, &banner.Title, &banner.Text, &banner.URL, &banner.IsActive, &banner.CreatedAt, &banner.UpdatedAt); err != nil {
			fmt.Println(err)
			return nil, err
		}
		banners = append(banners, banner)
		fmt.Println("ya")
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	fmt.Println("XUY")

	return banners, nil
}

// func UpdateBannerHandler(w http.ResponseWriter, r *http.Request) {
// 	idStr := mux.Vars(r)["id"]
// 	id, err := strconv.Atoi(idStr)
// 	if err != nil {
// 		http.Error(w, "Invalid banner ID", http.StatusBadRequest)
// 		return
// 	}

	// var requestBody struct {
	// 	TagIDs    []int  `json:"tag_ids,omitempty"`
	// 	FeatureID int    `json:"feature_id,omitempty"`
	// 	Content   struct {
	// 		Title string `json:"title,omitempty"`
	// 		Text  string `json:"text,omitempty"`
	// 		URL   string `json:"url,omitempty"`
	// 	} `json:"content,omitempty"`
	// 	IsActive *bool `json:"is_active,omitempty"`
	// }
	// if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
	// 	http.Error(w, "Invalid JSON format", http.StatusBadRequest)
	// 	return
	// }

// 	query := "UPDATE banners SET"
// 	var queryParams []interface{}
// 	var updateFields []string

// 	if len(requestBody.TagIDs) > 0 {
// 		updateFields = append(updateFields, "tag_ids = $1")
// 		queryParams = append(queryParams, requestBody.TagIDs)
// 	}
// 	if requestBody.FeatureID != 0 {
// 		updateFields = append(updateFields, "feature_id = $2")
// 		queryParams = append(queryParams, requestBody.FeatureID)
// 	}
// 	if requestBody.Content.Title != "" {
// 		updateFields = append(updateFields, "title = $3")
// 		queryParams = append(queryParams, requestBody.Content.Title)
// 	}
// 	if requestBody.Content.Text != "" {
// 		updateFields = append(updateFields, "text = $4")
// 		queryParams = append(queryParams, requestBody.Content.Text)
// 	}
// 	if requestBody.Content.URL != "" {
// 		updateFields = append(updateFields, "url = $5")
// 		queryParams = append(queryParams, requestBody.Content.URL)
// 	}
// 	if requestBody.IsActive != nil {
// 		updateFields = append(updateFields, "is_active = $6")
// 		queryParams = append(queryParams, requestBody.IsActive)
// 	}

// 	updateFields = append(updateFields, "updated_at = $7")
// 	queryParams = append(queryParams, time.Now())

// 	query += " " + strings.Join(updateFields, ", ") + " WHERE id = $8"
// 	queryParams = append(queryParams, id)

// 	database, err := db.GetPostgresDB()
// 	if err != nil {
// 		http.Error(w, "Failed to get database connection", http.StatusInternalServerError)
// 		return
// 	}

// 	_, err = database.Exec(query, queryParams...)
// 	if err != nil {
// 		http.Error(w, "Failed to update banner", http.StatusInternalServerError)
// 		return
// 	}

// 	// err = updateBannerInRedis(id, requestBody)
// 	// if err != nil {
// 	// 	// Если произошла ошибка при обновлении данных в Redis, удаляем данные из кэша
// 	// 	err := deleteBannerFromRedis(id)
// 	// 	if err != nil {
// 	// 		fmt.Println("Failed to delete banner from Redis:", err)
// 	// 	}
// 	// 	// Выводим сообщение об ошибке в консоль или журнал
// 	// 	fmt.Println("Failed to update banner in Redis:", err)
// 	// }

// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte("Banner updated successfully"))
// }

func DeleteBannerHandler(w http.ResponseWriter, r *http.Request) {
    adminToken := r.Header.Get("token")
	isAdmin, err := IsAdminTokenValid(adminToken)

	if err != nil {
	    http.Error(w, "Unauthorized", http.StatusUnauthorized)
	    return
	}

	if !isAdmin {
		http.Error(w, "Forbidden", http.StatusForbidden)
	    return
	}

    idStr := mux.Vars(r)["id"]
    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Invalid banner ID", http.StatusBadRequest)
        return
    }

    //featureID, tagID, err := GetFeatureIDAndTagIDByBannerID(id)
    if err != nil {
        http.Error(w, "Failed to get feature ID and tag ID", http.StatusInternalServerError)
        return
    }

    if err := DeleteBannerFromDB(id); err != nil {
        http.Error(w, "Failed to delete banner from database", http.StatusInternalServerError)
        return
    }

	// featureIDStr := strconv.Itoa(featureID)
	// tagIDStr := strconv.Itoa(tagID)
    // if err := db.DeleteBannerFromRedis(featureIDStr, tagIDStr); err != nil {
    //     fmt.Println("Failed to delete banner from Redis:", err)
    // }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Banner deleted successfully"))
}

func GetFeatureIDAndTagIDByBannerID(bannerID int) (int, int, error) {
    database, err := db.GetPostgresDB()
    if err != nil {
        return 0, 0, err
    }

    var featureID, tagID int
    query := "SELECT feature_id, tag_id FROM banners WHERE id = $1"
    err = database.QueryRow(query, bannerID).Scan(&featureID, &tagID)
    if err != nil {
        return 0, 0, err
    }

    return featureID, tagID, nil
}

func DeleteBannerFromDB(id int) error {
    database, err := db.GetPostgresDB()
    if err != nil {
        return err
    }

    query := "DELETE FROM banners WHERE id = $1"
    _, err = database.Exec(query, id)
    if err != nil {
        return err
    }

    return nil
}

func UpdateBannerHandler(w http.ResponseWriter, r *http.Request) {
    type UpdateBannerRequest struct {
        TagIDs    []int  `json:"tag_ids,omitempty"`
        FeatureID int    `json:"feature_id,omitempty"`
        Content   struct {
            Title string `json:"title,omitempty"`
            Text  string `json:"text,omitempty"`
            URL   string `json:"url,omitempty"`
        } `json:"content,omitempty"`
        IsActive bool `json:"is_active,omitempty"`
    }

    adminToken := r.Header.Get("token")
    isAdmin, err := IsAdminTokenValid(adminToken)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    if !isAdmin {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    idStr := mux.Vars(r)["id"]
    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Invalid banner ID", http.StatusBadRequest)
        return
    }

    var req UpdateBannerRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON format", http.StatusBadRequest)
        return
    }

    existingBanner, err := GetBannerByID(id)
    if err != nil {
        http.Error(w, "Failed to get banner", http.StatusInternalServerError)
        return
    }

    if err := DeleteBannerFromDB(id); err != nil {
        http.Error(w, "Failed to delete banner from database", http.StatusInternalServerError)
        return
    }

    if req.Content.Title != "" {
        existingBanner.Title = req.Content.Title
    }
    if req.Content.Text != "" {
        existingBanner.Text = req.Content.Text
    }
    if req.Content.URL != "" {
        existingBanner.URL = req.Content.URL
    }
    existingBanner.IsActive = req.IsActive
    existingBanner.UpdatedAt = time.Now()

	_, err = CreateBanner(*existingBanner)
    if err != nil {
        http.Error(w, "Failed to create banner in database", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Banner updated successfully"))
}

func GetBannerByID(id int) (*models.Banner, error) {
    database, err := db.GetPostgresDB()
    if err != nil {
        return nil, err
    }
    defer database.Close()

    query := "SELECT * FROM banners WHERE id = $1"

    var banner models.Banner
    row := database.QueryRow(query, id)
    err = row.Scan(&banner.ID, &banner.FeatureID, &banner.TagID, &banner.Title, &banner.Text, &banner.URL, &banner.IsActive, &banner.CreatedAt, &banner.UpdatedAt)
    if err != nil {
        return nil, err
    }

    return &banner, nil
}