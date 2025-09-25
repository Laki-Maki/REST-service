// cmd/server/main.go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
	"github.com/gorilla/mux"

	"subscription-service/internal/db"
	"subscription-service/internal/handler"
)

func main() {
	// read env
	host := getenv("DB_HOST", "localhost")
	port := getenv("DB_PORT", "5432")
	user := getenv("DB_USER", "postgres")
	pass := getenv("DB_PASSWORD", "postgres")
	name := getenv("DB_NAME", "subscription_db")
	appPort := getenv("PORT", "8080")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, name)

	dbConn, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	// Set reasonable timeouts/pool size
	dbConn.SetConnMaxLifetime(time.Minute * 5)
	dbConn.SetMaxOpenConns(20)
	dbConn.SetMaxIdleConns(10)

	// ping with retry
	var i int
	for {
		if err := dbConn.Ping(); err != nil {
			i++
			log.Printf("waiting for db (%d): %v", i, err)
			time.Sleep(time.Second * 2)
			if i > 15 {
				log.Fatalf("db not available: %v", err)
			}
			continue
		}
		break
	}
	log.Println("connected to db")

	store := &db.Store{DB: dbConn}
	h := handler.NewHandler(store)

	r := mux.NewRouter()
	h.RegisterRoutes(r)

	// simple logging middleware
	logged := loggingMiddleware(r)

	addr := ":" + appPort
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, logged); err != nil {
		log.Fatalf("listen: %v", err)
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
