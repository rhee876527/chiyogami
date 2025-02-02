package main

import (
	"chiyogami/core"
	"chiyogami/db"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	db.Init()

	r := mux.NewRouter()
	r.HandleFunc("/paste", core.CreatePasteHandler).Methods("POST")
	r.HandleFunc("/paste/{title}", core.GetPasteHandler).Methods("GET")
	r.HandleFunc("/paste/{title}", core.DeletePasteHandler).Methods("DELETE")
	r.HandleFunc("/pastes", core.ListPastesHandler).Methods("GET")
	r.HandleFunc("/user/pastes", core.ListUserPastesHandler).Methods("GET")
	r.HandleFunc("/register", core.RegisterHandler).Methods("POST")
	r.HandleFunc("/login", core.LoginHandler).Methods("POST")
	r.HandleFunc("/logout", core.LogoutHandler).Methods("POST")
	r.HandleFunc("/delete-account", core.DeleteAccountHandler).Methods("POST")
	r.HandleFunc("/generate-qr", core.GenerateQR).Methods("GET")
	r.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/list.html")
	})
	r.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./public/about.html")
	})
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./public/"))))

	log.Println("Server started on port 8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}
