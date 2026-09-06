package main

import (
	"encoding/json"
	"go-ping-api/pkg/algorithms"
	"go-ping-api/pkg/dao"
	"go-ping-api/pkg/fractals"
	grpcservice "go-ping-api/pkg/grpc/pb"
	pb "go-ping-api/pkg/grpc/pb/proto"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	fractalEngine *fractals.Engine
	daoManager    *dao.DAOManager
)

func main() {

	maxMsgSize := 20 * 1024 * 1024 // 20 MB

	fractalEngine = fractals.NewEngine()
	daoManager = dao.NewDAOManager()

	// 1. Initialize native gRPC Server
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize),
	)
	pb.RegisterFractalServiceServer(grpcServer, grpcservice.NewServer(fractalEngine))
	reflection.Register(grpcServer) // Enables auto-discovery for grpcurl

	// 2. HTTP REST Routes
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ping", handlePing)
	mux.HandleFunc("GET /api/fractals/generate", handleFractalGenerate)
	mux.HandleFunc("GET /GenerateRandomVertex_SpringBoot", handleRandomVertex)
	mux.HandleFunc("GET /api/data/getAllLogs", handleGetAllLogs)
	mux.HandleFunc("GET /api/data/getAllPersons", handleGetAllPersons)

	// Wrap REST mux with custom middleware pipeline
	restHandler := recoveryMiddleware(corsMiddleware(loggingMiddleware(mux)))

	// 3. Wrap gRPC for Browser Clients (gRPC-Web)
	wrappedGrpc := grpcweb.WrapServer(
		grpcServer,
		grpcweb.WithOriginFunc(func(origin string) bool {
			return origin == "https://apereznwo.github.io" || origin == "http://localhost:4200"
		}),
	)

	// Combined multiplexer passing requests to gRPC-Web, native gRPC (via ServeHTTP), or restHandler
	combinedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Flush buffer upon completion to force Render proxy stream resets
		if flusher, ok := w.(http.Flusher); ok {
			defer flusher.Flush()
		}

		// Handle gRPC-Web requests
		if wrappedGrpc.IsGrpcWebRequest(r) || wrappedGrpc.IsAcceptableGrpcCorsRequest(r) {
			wrappedGrpc.ServeHTTP(w, r)
			return
		}

		// Handle native gRPC requests over HTTP/2
		if r.ProtoMajor == 2 && r.Header.Get("Content-Type") == "application/grpc" {
			grpcServer.ServeHTTP(w, r)
			return
		}

		// Serves HTTP/REST through recovery, CORS, and logging middleware
		restHandler.ServeHTTP(w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 4. Instantiate server using helper function
	server := createServer(port, combinedHandler)

	log.Printf("Go server initialized (REST + gRPC + gRPC-Web) on port %s...", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server launch failure: %v", err)
	}
}

// Factory function encapsulating server timeout configurations and h2c multiplexing
func createServer(port string, handler http.Handler) *http.Server {
	h2s := &http2.Server{}
	return &http.Server{
		Addr:         ":" + port,
		Handler:      h2c.NewHandler(handler, h2s),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  30 * time.Second, // Recycles idle Render proxy connections
	}
}

// Middlewares
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		rw := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)
		log.Printf("HTTP %s %s -> Status: %d [Time: %v]", r.Method, r.URL.Path, rw.statusCode, time.Since(startTime))
	})
}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Internal error panic: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Handlers
func handlePing(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func handleFractalGenerate(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	kindVal, _ := strconv.Atoi(query.Get("kind"))

	kind, err := fractals.ParseFractalKind(kindVal)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defaultBounds := fractals.Bounds{XMin: -1.5, XMax: 1.5, YMin: -1.5, YMax: 1.5}
	if kind == fractals.Mandelbrot {
		defaultBounds = fractals.Bounds{XMin: -2.0, XMax: 1.0, YMin: -1.2, YMax: 1.2}
	}

	bounds := defaultBounds
	if xMin, err := strconv.ParseFloat(query.Get("xMin"), 64); err == nil {
		if xMax, err := strconv.ParseFloat(query.Get("xMax"), 64); err == nil {
			if yMin, err := strconv.ParseFloat(query.Get("yMin"), 64); err == nil {
				if yMax, err := strconv.ParseFloat(query.Get("yMax"), 64); err == nil {
					bounds = fractals.Bounds{XMin: xMin, XMax: xMax, YMin: yMin, YMax: yMax}
				}
			}
		}
	}

	maxIter := 500
	if iterStr := query.Get("maxIterations"); iterStr != "" {
		if parsedIter, err := strconv.Atoi(iterStr); err == nil {
			maxIter = parsedIter
		}
	}

	points := fractalEngine.GetFractal(kind, bounds, maxIter)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(points)
}

func handleRandomVertex(w http.ResponseWriter, r *http.Request) {
	result := algorithms.GenerateRandomPoints(9, 23, 0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(result))
}

func handleGetAllLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := daoManager.GetAllLogs()
	if err != nil {
		log.Printf("SQL Execution error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func handleGetAllPersons(w http.ResponseWriter, r *http.Request) {
	persons, err := daoManager.GetAllPersons()
	if err != nil {
		log.Printf("SQL Execution error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(persons)
}