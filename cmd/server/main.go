package main
import (
	"context"
	"document-version-compare-service/internal/application"
	"document-version-compare-service/internal/infrastructure"
	transport "document-version-compare-service/internal/transport/http"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)
	//go:embed web/*
	var webFiles embed.FS
func main() {
	dataFile := os.Getenv("DATA_FILE")
	if dataFile == "" {
		dataFile = "data/document-store.json"
	}
	store := infrastructure.NewMemoryStore(dataFile)
	if err := store.LoadError(); err != nil {
		log.Fatalf("load data: %v", err)
	}
	queue := infrastructure.NewQueue(128)
	services := application.NewServices(store, queue.Enqueue)
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	services.Comparisons.SetContext(ctx)
	queue.Run(ctx, 2, services.Comparisons.Process)
	for _, jobID := range store.ResumeJobs() {
		queue.Enqueue(jobID)
	}
	api := transport.NewServer(services)
	assets, _ := fs.Sub(webFiles, "web")
	mux := http.NewServeMux()
	mux.Handle("/api/", api.Handler())
	mux.Handle("/healthz", api.Handler())
	mux.Handle("/readyz", api.Handler())
	mux.Handle("/", http.FileServer(http.FS(assets)))
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{Addr: ":" + port, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		log.Printf("document compare service listening on http://localhost:%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	<-ctx.Done()
	shutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()
	_ = server.Shutdown(shutdown)
}
