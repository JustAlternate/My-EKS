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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// GOLDEN SIGNAL: TRAFFIC
	// Total number of requests received
	reqsReceived = promauto.NewCounter(prometheus.CounterOpts{
		Name: "received_request_total",
		Help: "The total number of received requests",
	})

	// Total number of requests processed successfully
	reqsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "processed_request_total",
		Help: "The total number of processed requests",
	})

	// GOLDEN SIGNAL: ERRORS
	// Total number of failed requests
	reqsErrored = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "request_errors_total",
		Help: "The total number of failed requests by error type",
	}, []string{"error_type"})

	// GOLDEN SIGNAL: LATENCY
	// Request duration histogram
	requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests in seconds",
		Buckets: prometheus.DefBuckets, // [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
	}, []string{"endpoint", "status"})

	// Database operation duration
	dbOperationDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "db_operation_duration_seconds",
		Help:    "Duration of database operations in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation"})

	// GOLDEN SIGNAL: SATURATION
	// Number of in-flight requests
	inFlightRequests = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "http_requests_in_flight",
		Help: "Current number of HTTP requests being processed",
	})

	// Database connection status
	dbConnectionStatus = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "db_connection_status",
		Help: "Database connection status (1 = connected, 0 = disconnected)",
	})
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
		dbConnectionStatus.Set(0)
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	dbConnectionStatus.Set(1)
	log.Println("Successfully connected to database")
	return conn, nil
}

func InitTableIfNotExist(conn *pgx.Conn) error {
	start := time.Now()
	defer func() {
		dbOperationDuration.WithLabelValues("init_table").Observe(time.Since(start).Seconds())
	}()

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
	start := time.Now()
	defer func() {
		dbOperationDuration.WithLabelValues("update_counter").Observe(time.Since(start).Seconds())
	}()

	sql := `UPDATE counters SET counter = counter + 1 WHERE id = 0`
	_, err := conn.Exec(context.Background(), sql)
	if err != nil {
		return fmt.Errorf("error while updating counter: %v", err)
	}
	return nil
}

func getCounter(conn *pgx.Conn) (int, error) {
	start := time.Now()
	defer func() {
		dbOperationDuration.WithLabelValues("get_counter").Observe(time.Since(start).Seconds())
	}()

	var counter int
	err := conn.QueryRow(context.Background(), "SELECT counter FROM counters WHERE id = 0").Scan(&counter)
	if err != nil {
		return 0, fmt.Errorf("error while getting counter: %v", err)
	}
	return counter, nil
}

func (app *App) root(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	status := "200"

	// Track in-flight requests (saturation)
	inFlightRequests.Inc()
	defer inFlightRequests.Dec()

	// Track traffic
	reqsReceived.Inc()
	log.Printf("received call from: %v\n", r.RemoteAddr)

	defer func() {
		// Track latency
		requestDuration.WithLabelValues("/", status).Observe(time.Since(start).Seconds())
	}()

	err := updateCounter(app.conn)
	if err != nil {
		log.Printf("Error updating counter: %v", err)
		status = "500"
		reqsErrored.WithLabelValues("db_update_failed").Inc()
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	counter, err := getCounter(app.conn)
	if err != nil {
		log.Printf("Error getting counter: %v", err)
		status = "500"
		reqsErrored.WithLabelValues("db_select_failed").Inc()
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = fmt.Fprintf(w, "%d", counter)
	if err != nil {
		log.Printf("Error when writting the counter in the response: %v", err)
		reqsErrored.WithLabelValues("response_write_failed").Inc()
	}
	reqsProcessed.Inc()
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

	http.Handle("/metrics", promhttp.Handler())

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
