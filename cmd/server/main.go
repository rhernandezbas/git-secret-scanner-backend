package main

import (
	"log"
	"net/http"

	"github.com/rhernandezba/git-secret-scanner/backend/internal/analyzer"
	"github.com/rhernandezba/git-secret-scanner/backend/internal/cloner"
	"github.com/rhernandezba/git-secret-scanner/backend/internal/config"
	"github.com/rhernandezba/git-secret-scanner/backend/internal/domain"
	"github.com/rhernandezba/git-secret-scanner/backend/internal/handler"
	"github.com/rhernandezba/git-secret-scanner/backend/internal/provider"
	"github.com/rhernandezba/git-secret-scanner/backend/internal/repository"
	"github.com/rhernandezba/git-secret-scanner/backend/internal/scanner"
	"github.com/rhernandezba/git-secret-scanner/backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	providerFactory := service.ProviderFactory(func(providerName string) (domain.RepoProvider, error) {
		return provider.New(providerName, cfg.GithubToken, cfg.GitlabToken)
	})

	repoCloner := cloner.New(cfg.TempDir, cfg.MaxRepoSizeMB)
	repoScanner := scanner.New()
	findingsRepo := repository.NewJSONFindingsRepository(cfg.FindingsFile)
	hub := handler.NewHub()

	var repoAnalyzer domain.Analyzer
	if cfg.AIAnalysisEnabled {
		repoAnalyzer, err = analyzer.New(cfg.AIProvider, cfg.AIAPIKey, cfg.AIModel)
		if err != nil {
			log.Fatalf("analyzer error: %v", err)
		}
	}

	svc := service.NewScanService(
		providerFactory,
		repoCloner,
		repoScanner,
		findingsRepo,
		hub,
		repoAnalyzer,
		cfg.AIAnalysisEnabled,
	)

	scanHandler := handler.NewScanHandler(svc)
	findingsHandler := handler.NewFindingsHandler(findingsRepo)
	wsHandler := handler.NewWSHandler(hub)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", scanHandler.Health)
	mux.HandleFunc("POST /scan", scanHandler.Scan)
	mux.HandleFunc("GET /findings", findingsHandler.GetAll)
	mux.HandleFunc("GET /ws", wsHandler.ServeWS)

	log.Printf("git-secret-scanner backend starting on :%s", cfg.Port)
	log.Printf("AI analysis: %v | provider: %s", cfg.AIAnalysisEnabled, cfg.AIProvider)

	if err := http.ListenAndServe(":"+cfg.Port, handler.CORSMiddleware(mux)); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
