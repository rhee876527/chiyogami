package core

import (
	"chiyogami/db"
	"chiyogami/models"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/skip2/go-qrcode"
)

// Make unique titles
func GenerateUniqueTitle() string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	for {
		b := make([]byte, 4)
		if _, err := rand.Read(b); err != nil {
			return "Error: failed to generate random bytes - " + err.Error()
		}
		for i := range b {
			b[i] = byte(letters[int(b[i])%len(letters)])
		}
		title := string(b)
		var paste models.Paste
		if db.DB.First(&paste, "title = ?", title).Error != nil {
			return title
		}
	}
}

// Manage secret key for session(s)
func GetSessionKey() string {
	key := os.Getenv("SECRET_KEY")
	if key == "" {
		log.Println("No SECRET_KEY found. Generating a random key.")
		key = GenerateRandomKey()
	}
	return key
}

func GenerateRandomKey() string {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "Error: failed to generate random key - " + err.Error()
	}
	return base64.StdEncoding.EncodeToString(key)
}

// Generate QR png
func GenerateQR(w http.ResponseWriter, r *http.Request) {
	// Only allow requests with custom header
	if r.Header.Get("X-Requested-By") != "qr-allowed" {
		JsonRespond(w, http.StatusForbidden, "Forbidden")
		return
	}

	url := r.URL.Query().Get("url")
	if url == "" {
		JsonRespond(w, http.StatusBadRequest, "Missing 'url' query parameter")
		return
	}

	png, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		log.Println("Error generating QR code:", err)
		JsonRespond(w, http.StatusInternalServerError, "Failed to generate QR code")
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)

	_, writeErr := w.Write(png)
	if writeErr != nil {
		log.Println("Error writing QR code to response:", writeErr)
	}
}

// Human-readable expiry duration
func TimeUntilExpiration(expiration *time.Time) string {
	if expiration == nil {
		return "Never"
	}
	diff := time.Until(*expiration)
	units := []struct {
		d time.Duration
		s string
	}{
		{24 * time.Hour, "day"},
		{time.Hour, "hour"},
		{time.Minute, "minute"},
		{time.Second, "second"},
	}

	for _, unit := range units {
		if d := int(diff / unit.d); d > 0 {
			pluralSuffix := "s"
			if d == 1 {
				pluralSuffix = ""
			}
			return fmt.Sprintf("in %d %s%s", d, unit.s, pluralSuffix)
		}
	}
	return ""
}
