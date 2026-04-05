package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/joss12/api-time-machine/internal/db"
	"github.com/joss12/api-time-machine/internal/proxy"
)

func main() {

	godotenv.Load()
	db.Connect()
	db.CreateTable()
	db.ClosestaleSession()

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/sessions", proxy.StartSessionHandler)
	mux.HandleFunc("POST /api/sessions/{id}/end", proxy.EndSessionHanlder)
	mux.HandleFunc("GET /api/sessions", proxy.ListSessionsHandler)
	mux.HandleFunc("GET /api/sessions/{id}/requests", proxy.GetSessionRequestsHandler)
	mux.HandleFunc("POST /api/sessions/{id}/replay", proxy.ReplaySessionHandler)

	mux.HandleFunc("/proxy/", proxy.ProxyHandler)
	mux.HandleFunc("GET /health", proxy.HealthHandler)
	mux.HandleFunc("DELETE /api/sessions/{id}", proxy.DeleteSessionHandler)

	handler := proxy.CORSMiddleware(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
	}

	log.Printf("server running on :%s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), handler))
}
