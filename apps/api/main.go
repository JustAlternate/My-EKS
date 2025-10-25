package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5"
)

type App struct {
	conn *pgx.Conn
}

func connect() (*pgx.Conn, error) {
	dbUsername := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbURL := os.Getenv("DB_URL")
	dbName := os.Getenv("DB_NAME")

	log.Printf("Attempting to connect to database with: username=%s, url=%s, dbname=%s", dbUsername, dbURL, dbName)

	url := fmt.Sprintf("postgres://%s:%s@%s:5432/%s?sslmode=require&connect_timeout=5",
		dbUsername, dbPassword, dbURL, dbName)

	log.Printf("Connection string: %s", url)

	conn, err := pgx.Connect(context.Background(), url)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	log.Println("Successfully connected to database")
	return conn, nil
}

func InitTableIfNotExist(conn *pgx.Conn) error {
	sql := `CREATE TABLE IF NOT EXISTS counters (
        id integer PRIMARY KEY,
        counter integer NOT NULL DEFAULT 0
    )`
	_, err := conn.Exec(context.Background(), sql)
	if err != nil {
		return fmt.Errorf("error while creating table: %v", err)
	}

	var exists bool
	err = conn.QueryRow(context.Background(), "SELECT EXISTS(SELECT 1 FROM counters WHERE id = 0)").Scan(&exists)
	if err != nil {
		return fmt.Errorf("error checking if row exists: %v", err)
	}

	if !exists {
		sql = `INSERT INTO counters (id, counter) VALUES (0, 0)`
		_, err = conn.Exec(context.Background(), sql)
		if err != nil {
			return fmt.Errorf("error while inserting first row: %v", err)
		}
	}

	return nil
}

func updateCounter(conn *pgx.Conn) error {
	sql := `UPDATE counters SET counter = counter + 1 WHERE id = 0`
	_, err := conn.Exec(context.Background(), sql)
	if err != nil {
		return fmt.Errorf("error while updating counter: %v", err)
	}
	return nil
}

func getCounter(conn *pgx.Conn) (int, error) {
	var counter int
	err := conn.QueryRow(context.Background(), "SELECT counter FROM counters WHERE id = 0").Scan(&counter)
	if err != nil {
		return 0, fmt.Errorf("error while getting counter: %v", err)
	}
	return counter, nil
}

func (app *App) root(w http.ResponseWriter, r *http.Request) {
	log.Printf("received call from: %v\n", r.RemoteAddr)

	err := updateCounter(app.conn)
	if err != nil {
		log.Printf("Error updating counter: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	counter, err := getCounter(app.conn)
	if err != nil {
		log.Printf("Error getting counter: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = fmt.Fprintf(w, "%d", counter)
	if err != nil {
		log.Printf("Error when writting the counter in the response: %v", err)
	}
}

var (
	ready bool = false
	mu    sync.Mutex
)

func main() {
	server := &http.Server{
		Addr: ":3030",
	}

	log.Println("Serving liveness and readiness...")
	http.HandleFunc("/liveness", liveness)
	http.HandleFunc("/readiness", readiness)

	go func() {
		log.Println("Starting server on :3030")
		err := server.ListenAndServe()
		if err != nil {
			log.Fatalf("error when listen and serve: %v", err)
		}
	}()

	log.Println("Connecting to the database...")
	conn, err := connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Initializing database table...")
	err = InitTableIfNotExist(conn)
	if err != nil {
		log.Fatalf("Failed to initialize table: %v", err)
	}

	app := &App{
		conn: conn,
	}

	log.Println("Database initialization complete")

	http.HandleFunc("/", app.root)

	mu.Lock()
	ready = true
	mu.Unlock()

	log.Println("Application is READY")

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
	<-sigint

	log.Println("Shutting down server...")

	mu.Lock()
	ready = false
	mu.Unlock()

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	if app.conn != nil {
		if err := app.conn.Close(context.Background()); err != nil {
			log.Printf("Error closing conn: %v", err)
		}
	}
	log.Println("Server stopped")
}

func liveness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := fmt.Fprintln(w, "OK")
	if err != nil {
		log.Fatal(err)
	}

}

func readiness(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	if ready {
		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprintln(w, "OK")
		if err != nil {
			log.Fatal(err)
		}
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, err := fmt.Fprintln(w, "Service Unavailable")
		if err != nil {
			log.Fatal(err)
		}
	}
}
