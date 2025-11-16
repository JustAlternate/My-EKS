package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// GOLDEN SIGNAL: TRAFFIC
	reqsReceived = promauto.NewCounter(prometheus.CounterOpts{
		Name: "received_request_total",
		Help: "The total number of received requests",
	})

	reqsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "processed_request_total",
		Help: "The total number of processed requests",
	})

	requestsByEndpoint = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests by endpoint and status",
	}, []string{"endpoint", "status"})

	// GOLDEN SIGNAL: ERRORS
	reqsErrored = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "request_errors_total",
		Help: "The total number of failed requests by error type",
	}, []string{"error_type", "endpoint"})

	apiCallErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "api_call_errors_total",
		Help: "The total number of failed API calls by error type",
	}, []string{"error_type"})

	// GOLDEN SIGNAL: LATENCY
	requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "Duration of HTTP requests in seconds",
		Buckets: prometheus.DefBuckets, // [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]
	}, []string{"endpoint", "status"})

	// API call duration
	apiCallDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "api_call_duration_seconds",
		Help:    "Duration of API calls in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"status"})

	// GOLDEN SIGNAL: SATURATION
	inFlightRequests = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "http_requests_in_flight",
		Help: "Current number of HTTP requests being processed",
	})

	// In-flight API calls
	inFlightAPICalls = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "api_calls_in_flight",
		Help: "Current number of API calls being processed",
	})

	// API response status codes
	apiResponseStatus = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "api_response_status_total",
		Help: "Total number of API responses by status code",
	}, []string{"status_code"})

	// Response size
	responseSizeBytes = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_response_size_bytes",
		Help:    "Size of HTTP responses in bytes",
		Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000, 500000, 1000000},
	}, []string{"endpoint"})
)

type APIHandler struct {
	apiUrl string
}

func (a *APIHandler) send(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	status := "200"
	endpoint := "/send"

	// Track in-flight requests (saturation)
	inFlightRequests.Inc()
	defer inFlightRequests.Dec()

	// Track traffic
	reqsReceived.Inc()

	log.Printf("received call from: %v\n", r.RemoteAddr)

	defer func() {
		// Track latency
		duration := time.Since(start).Seconds()
		requestDuration.WithLabelValues(endpoint, status).Observe(duration)
		requestsByEndpoint.WithLabelValues(endpoint, status).Inc()
	}()

	// Call the external API
	apiStart := time.Now()
	inFlightAPICalls.Inc()
	resp, err := http.Get(a.apiUrl)
	inFlightAPICalls.Dec()
	apiCallDuration.WithLabelValues(fmt.Sprintf("%d", getStatusCode(resp, err))).Observe(time.Since(apiStart).Seconds())

	if err != nil {
		log.Printf("failed to call API: %v\n", err)
		status = "500"
		reqsErrored.WithLabelValues("api_call_failed", endpoint).Inc()
		apiCallErrors.WithLabelValues("connection_error").Inc()
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("Failed to fetch counter"))
		if err != nil {
			log.Printf("error writing response: %v", err)
			reqsErrored.WithLabelValues("response_write_failed", endpoint).Inc()
		}
		return
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("error closing response body: %v", err)
		}
	}()

	// Track API response status
	apiResponseStatus.WithLabelValues(fmt.Sprintf("%d", resp.StatusCode)).Inc()

	if resp.StatusCode != http.StatusOK {
		log.Printf("API returned non-OK status: %d\n", resp.StatusCode)
		status = "500"
		reqsErrored.WithLabelValues("api_non_ok_status", endpoint).Inc()
		apiCallErrors.WithLabelValues(fmt.Sprintf("status_%d", resp.StatusCode)).Inc()
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("API error"))
		if err != nil {
			log.Printf("error writing response: %v", err)
			reqsErrored.WithLabelValues("response_write_failed", endpoint).Inc()
		}
		return
	}

	// Read the counter response from the API
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		log.Printf("failed to read response body: %v\n", err)
		status = "500"
		reqsErrored.WithLabelValues("response_read_failed", endpoint).Inc()
		w.WriteHeader(http.StatusInternalServerError)
		_, err := w.Write([]byte("Failed to read counter"))
		if err != nil {
			log.Printf("error writing response: %v", err)
			reqsErrored.WithLabelValues("response_write_failed", endpoint).Inc()
		}
		return
	}

	// Track response size
	responseSizeBytes.WithLabelValues(endpoint).Observe(float64(len(body)))

	// Return the counter to the client
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(body)
	if err != nil {
		log.Printf("error writing response: %v", err)
		reqsErrored.WithLabelValues("response_write_failed", endpoint).Inc()
	} else {
		reqsProcessed.Inc()
	}
}

// Helper function to get status code from response or error
func getStatusCode(resp *http.Response, err error) int {
	if err != nil {
		return 0
	}
	if resp == nil {
		return 0
	}
	return resp.StatusCode
}

var (
	ready bool = false
	mu    sync.Mutex
)

func main() {
	server := &http.Server{
		Addr: ":8080",
	}

	http.HandleFunc("/liveness", liveness)
	http.HandleFunc("/readiness", readiness)

	go func() {
		log.Println("Starting server on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		log.Fatal("API_URL environment variable not set")
	}

	apiHandler := &APIHandler{apiUrl: apiURL}
	http.HandleFunc("/send", apiHandler.send)
	http.Handle("/", instrumentHandler("/", http.FileServer(http.Dir("./static"))))
	http.Handle("/metrics", promhttp.Handler())

	mu.Lock()
	ready = true
	mu.Unlock()

	log.Println("Application is READY")

	// Setup graceful shutdown
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

	log.Println("Server stopped")
}

// instrumentHandler wraps a handler to add metrics
func instrumentHandler(endpoint string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		inFlightRequests.Inc()
		defer inFlightRequests.Dec()

		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		handler.ServeHTTP(wrapped, r)

		duration := time.Since(start).Seconds()
		status := fmt.Sprintf("%d", wrapped.statusCode)

		requestDuration.WithLabelValues(endpoint, status).Observe(duration)
		requestsByEndpoint.WithLabelValues(endpoint, status).Inc()
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func liveness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	if _, err := fmt.Fprintln(w, "OK"); err != nil {
		log.Printf("error writing liveness response: %v", err)
	}
}

func readiness(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	if ready {
		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprintln(w, "OK"); err != nil {
			log.Fatal(err)
		}
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		if _, err := fmt.Fprintln(w, "Service Unavailable"); err != nil {
			log.Fatal(err)
		}
	}
}
