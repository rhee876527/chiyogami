package core

import (
	"chiyogami/db"
	"chiyogami/models"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

// Init sessions
var store = sessions.NewCookieStore([]byte(GetSessionKey()))

// Json all the client errors
func JsonRespond(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"message": message})
}

// Create paste
func CreatePasteHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	userID, ok := session.Values["user_id"].(uint)

	// Implement rate limiter
	ip := GetIPAddress(r)

	// Create a composite identifier using both session and IP
	sessionIdentifier := "anon"
	if ok && userID != 0 {
		sessionIdentifier = fmt.Sprintf("user-%d", userID)
	}
	identifier := fmt.Sprintf("%s|%s", sessionIdentifier, ip)

	// Return rate limit status for request
	if !CheckAndRecordRateLimit(identifier) {
		JsonRespond(w, http.StatusTooManyRequests, "Rate limit exceeded. Please try again later.")
		return
	}

	// Begin client input decode
	var pasteRequest struct {
		Content     string `json:"content"`
		Visibility  string `json:"visibility"`
		Expiration  string `json:"expiration"`
		IsEncrypted bool   `json:"isEncrypted"`
	}

	// Reject invalid inputs
	if err := json.NewDecoder(r.Body).Decode(&pasteRequest); err != nil {
		JsonRespond(w, http.StatusBadRequest, "Smelly! Request body not compatible JSON format")
		return
	}

	if content := strings.TrimSpace(pasteRequest.Content); content == "" || len(content) > 30000 {
		JsonRespond(w, http.StatusBadRequest, "Content invalid or size exceeds 30000 max chars")
		return
	}

	// Specify destination object
	paste := models.Paste{
		Content:     pasteRequest.Content,
		Visibility:  pasteRequest.Visibility,
		IsEncrypted: pasteRequest.IsEncrypted,
		Title:       GenerateUniqueTitle(),
		UserID:      userID,
		IsUserPaste: (userID != 0),
	}

	// Set default visibility
	if paste.Visibility == "" {
		paste.Visibility = "Public"
	}

	// Paste expiration
	defaultExpiration := os.Getenv("PASTE_DEFAULT_EXPIRATION")
	if defaultExpiration == "" {
		defaultExpiration = "24h"
	}

	expiration := pasteRequest.Expiration
	if expiration == "" {
		expiration = defaultExpiration
	}

	if strings.ToLower(expiration) != "never" {
		duration, err := time.ParseDuration(expiration)
		if err != nil {
			if expiration == defaultExpiration {
				JsonRespond(w, http.StatusInternalServerError, "PASTE_DEFAULT_EXPIRATION invalid. Contact administrator.")
			} else {
				JsonRespond(w, http.StatusBadRequest, "Invalid expiration value")
			}
			return
		}
		expirationTime := time.Now().Add(duration)
		paste.Expiration = &expirationTime
	}

	// Save to db & record any errors
	if err := db.DB.Create(&paste).Error; err != nil {
		JsonRespond(w, http.StatusInternalServerError, "Failed to save paste")
		return
	}

	// Return created title
	json.NewEncoder(w).Encode(map[string]string{"title": paste.Title})
}

// Get created pastes
func GetPasteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	title := vars["title"]

	log.Println("Fetching paste with title:", title)

	// Auto manage expired pastes
	DeleteExpiredPastes()

	// Return existent & non-expired pastes
	var paste models.Paste
	if db.DB.First(&paste, "title = ? AND (expiration IS NULL OR expiration > ?)", title, time.Now()).Error != nil {
		JsonRespond(w, http.StatusNotFound, "Paste not found or has expired")
		return
	}

	// Make responses api
	if r.Header.Get("Accept") == "application/json" {
		json.NewEncoder(w).Encode(paste)
		return
	}

	// Template paste to frontend
	tmpl, err := template.New("paste").ParseFiles("./public/tmpl.html")
	if err != nil {
		JsonRespond(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	tmpl.Execute(w, struct {
		Title       string
		Content     template.HTML
		CreatedAt   string
		IsEncrypted bool
		Expiration  string
	}{
		Title:       paste.Title,
		Content:     template.HTML(paste.Content),
		CreatedAt:   paste.CreatedAt.Format(time.RFC3339),
		IsEncrypted: paste.IsEncrypted,
		Expiration:  TimeUntilExpiration(paste.Expiration),
	})
}

// Register user
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		JsonRespond(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if len(user.Username) > 8 {
		JsonRespond(w, http.StatusBadRequest, "Username must be at most 8 characters")
		return
	}

	// Store user password securely
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		JsonRespond(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}
	user.Password = string(hashedPassword)

	if err := db.DB.Create(&user).Error; err != nil {
		JsonRespond(w, http.StatusInternalServerError, err.Error())
		return
	}

	JsonRespond(w, http.StatusCreated, "User registered successfully")
}

// Login user
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var loginData struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&loginData)

	var user models.User
	if db.DB.Where("username = ?", loginData.Username).First(&user).Error != nil {
		JsonRespond(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password)) != nil {
		JsonRespond(w, http.StatusUnauthorized, "Bad password")
		return
	}

	session, _ := store.Get(r, "session")
	session.Values["user_id"] = user.ID
	session.Save(r, w)
	w.WriteHeader(http.StatusOK)
}

// Logout user
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	session.Values["user_id"] = nil
	session.Save(r, w)
	w.WriteHeader(http.StatusOK)
}

// Show public pastes
func ListPastesHandler(w http.ResponseWriter, r *http.Request) {
	DeleteExpiredPastes()

	searchQuery := r.URL.Query().Get("search")
	var publicPastes []models.Paste

	// Exclude non-public & encrypted
	query := db.DB.Model(&models.Paste{}).Where("visibility = ? AND is_encrypted = ?", "Public", false)

	// Search relevant keywords
	if searchQuery != "" {
		searchPattern := "%" + searchQuery + "%"
		query = query.Where("title LIKE ? OR content LIKE ?", searchPattern, searchPattern)
	}

	// Order by date
	err := query.Order("created_at DESC").Find(&publicPastes).Error
	if err != nil {
		JsonRespond(w, http.StatusInternalServerError, "Error fetching pastes")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(publicPastes); err != nil {
		JsonRespond(w, http.StatusInternalServerError, "Error encoding response")
	}
}

// Pastes created by user
func ListUserPastesHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	userID, ok := session.Values["user_id"].(uint)
	if !ok {
		JsonRespond(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	var userPastes []models.Paste
	db.DB.Where("user_id = ? AND is_user_paste = ?", userID, true).Find(&userPastes)
	json.NewEncoder(w).Encode(userPastes)
}

// Delete user account
func DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	session, err := store.Get(r, "session")
	if err != nil {
		JsonRespond(w, http.StatusInternalServerError, "Failed to get session")
		return
	}

	userID, ok := session.Values["user_id"].(uint)
	if !ok {
		JsonRespond(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Delete pastes when user deletes account
	if err := db.DB.Where("user_id = ?", userID).Delete(&models.Paste{}).Error; err != nil {
		log.Printf("Error deleting user's pastes: %v", err)
		JsonRespond(w, http.StatusInternalServerError, "Failed to delete user pastes")
		return
	}

	if err := db.DB.Delete(&models.User{}, userID).Error; err != nil {
		log.Printf("Error deleting user: %v", err)
		JsonRespond(w, http.StatusInternalServerError, "Failed to delete account")
		return
	}

	session.Values["user_id"] = nil
	if err := session.Save(r, w); err != nil {
		log.Printf("Error clearing session: %v", err)
		JsonRespond(w, http.StatusInternalServerError, "Failed to clear session after deletion")
		return
	}

	log.Printf("Successfully deleted account for user ID: %d", userID)
	w.WriteHeader(http.StatusOK)
}

// Delete invalid pastes
func DeleteExpiredPastes() {
	db.DB.Where("expiration IS NOT NULL AND expiration < ?", time.Now()).Delete(&models.Paste{})
}

// Delete pastes using session
func DeletePasteHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session")
	userID, ok := session.Values["user_id"].(uint)
	if !ok || userID == 0 {
		JsonRespond(w, http.StatusUnauthorized, "Unauthorized action")
		return
	}

	vars := mux.Vars(r)
	title := vars["title"]

	var paste models.Paste
	if db.DB.First(&paste, "title = ?", title).Error != nil {
		JsonRespond(w, http.StatusNotFound, "Paste not found")
		return
	}

	if paste.UserID != userID {
		JsonRespond(w, http.StatusForbidden, "Forbidden")
		return
	}

	if err := db.DB.Delete(&paste).Error; err != nil {
		JsonRespond(w, http.StatusInternalServerError, "Failed to delete paste")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Paste deleted successfully"})
}
