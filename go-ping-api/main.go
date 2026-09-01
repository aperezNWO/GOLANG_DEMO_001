package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"go-ping-api/pkg/algorithms"
	"go-ping-api/pkg/dao"
	"go-ping-api/pkg/fractals"
	grpcservice "go-ping-api/pkg/grpc/pb"
	pb "go-ping-api/pkg/grpc/pb/proto"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"
)

var (
	fractalEngine *fractals.Engine
	daoManager    *dao.DAOManager
)

func main() {
	fractalEngine = fractals.NewEngine()
	daoManager = dao.NewDAOManager()

	mux := http.NewServeMux()

	// REST Routes
	mux.HandleFunc("GET /ping", handlePing)
	mux.HandleFunc("GET /api/fractals/generate", handleFractalGenerate)
	mux.HandleFunc("GET /GenerateRandomVertex_SpringBoot", handleRandomVertex)
	mux.HandleFunc("GET /api/data/getAllLogs", handleGetAllLogs)
	mux.HandleFunc("GET /api/data/getAllPersons", handleGetAllPersons)

	// Wrap REST Mux with Middlewares
	restHandler := recoveryMiddleware(corsMiddleware(loggingMiddleware(mux)))

	// Register gRPC Service & Wrap with gRPC-Web
	grpcServer := grpc.NewServer()
	pb.RegisterFractalServiceServer(grpcServer, grpcservice.NewServer(fractalEngine))
	wrappedGrpc := grpcweb.WrapServer(grpcServer)

	// Combine REST and gRPC-Web into a single HTTP Multiplexing Handler
	combinedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if wrappedGrpc.IsGrpcWebRequest(r) || wrappedGrpc.IsAcceptableGrpcCorsRequest(r) {
			wrappedGrpc.ServeHTTP(w, r)
			return
		}
		restHandler.ServeHTTP(w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Go server initialized (REST + gRPC-Web) on port %s...", port)
	if err := http.ListenAndServe(":"+port, combinedHandler); err != nil {
		log.Fatalf("Server launch failure: %v", err)
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