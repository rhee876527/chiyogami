package main

import (
	"chiyogami/core"
	"chiyogami/db"
	"embed"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

//go:embed LICENSE
var License embed.FS

func main() {
	db.Init()

	r := mux.NewRouter()
	r.HandleFunc("/paste", core.CreatePasteHandler).Methods("POST", "PUT")
	r.HandleFunc("/paste/{title}", core.GetPasteHandler).Methods("GET")
	r.HandleFunc("/paste/{title}", core.DeletePasteHandler).Methods("DELETE")
	r.HandleFunc("/pastes", core.ListPastesHandler).Methods("GET")
	r.HandleFunc("/user/pastes", core.ListUserPastesHandler).Methods("GET")
	r.HandleFunc("/register", core.RegisterHandler).Methods("POST")
	r.HandleFunc("/login", core.LoginHandler).Methods("POST")
	r.HandleFunc("/logout", core.LogoutHandler).Methods("POST")
	r.HandleFunc("/delete-account", core.DeleteAccountHandler).Methods("DELETE")
	r.HandleFunc("/generate-qr", core.GenerateQR).Methods("GET")
	r.HandleFunc("/health", core.HealthCheck).Methods("GET")
	r.HandleFunc("/info", core.GetConfigVariables).Methods("GET")
	r.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/list.html")
	})
	r.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/about.html")
	})
	r.HandleFunc("/license", core.ServeLicense(License))
	r.HandleFunc("/robots.txt", core.RobotsTxtHandler)
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./public/"))))

	// Apply CSRF protection middleware
	csrfProtectedRouter := (&http.CrossOriginProtection{}).Handler(r)

	// Set port
	port := os.Getenv("PORT")

	if _, ok := os.LookupEnv("DOCKER_ENV"); ok {
		port = "8000"
	}

	if port == "" {
		port = "8000"
	}

	log.Printf("Server started on port %s", port)

	log.Fatal(http.ListenAndServe(":"+port, csrfProtectedRouter))
}
