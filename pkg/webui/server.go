package webui

//go:generate go install github.com/a-h/templ/cmd/templ@latest
//go:generate templ generate

import (
	"context"
	"database/sql"
	"log/slog"
	"math"
	"net/http"
	"strconv"
)

const (
	// ItemsPerPage defines how many deposits to show per page
	ItemsPerPage = 10
)

// Server represents the web UI server
type Server struct {
	db     *sql.DB
	logger *slog.Logger
	addr   string
}

// NewServer creates a new web UI server
func NewServer(db *sql.DB, logger *slog.Logger, addr string) *Server {
	return &Server{
		db:     db,
		logger: logger,
		addr:   addr,
	}
}

// Start starts the web UI server
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// Register handlers
	mux.HandleFunc("GET /", s.handleIndex)

	// Dashboard component handlers
	mux.HandleFunc("GET /dashboard/metrics", s.handleDashboardMetrics)
	mux.HandleFunc("GET /dashboard/performance", s.handleBridgePerformance)
	mux.HandleFunc("GET /dashboard/unmatched", s.handleUnmatchedDepositsSection)
	mux.HandleFunc("GET /dashboard/timeline", s.handleDepositsTimelineSection)

	server := &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	s.logger.Info("starting web UI server", "addr", s.addr)

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		s.logger.Info("shutting down web UI server")
		return server.Shutdown(context.Background())
	}
}

// handleIndex handles the index page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	component := Dashboard()
	err := component.Render(r.Context(), w)
	if err != nil {
		s.logger.Error("failed to render dashboard", "error", err)
		http.Error(w, "Failed to render dashboard", http.StatusInternalServerError)
		return
	}
}

// handleDashboardMetrics handles the dashboard metrics component
func (s *Server) handleDashboardMetrics(w http.ResponseWriter, r *http.Request) {
	stats, err := GetBridgeStats(r.Context(), s.db)
	if err != nil {
		s.logger.Error("failed to get bridge stats", "error", err)
		http.Error(w, "Failed to get bridge stats", http.StatusInternalServerError)
		return
	}

	component := DashboardMetrics(stats)
	err = component.Render(r.Context(), w)
	if err != nil {
		s.logger.Error("failed to render dashboard metrics", "error", err)
		http.Error(w, "Failed to render dashboard metrics", http.StatusInternalServerError)
		return
	}
}

// handleBridgePerformance handles the bridge performance component
func (s *Server) handleBridgePerformance(w http.ResponseWriter, r *http.Request) {
	stats, err := GetBridgeStats(r.Context(), s.db)
	if err != nil {
		s.logger.Error("failed to get bridge stats", "error", err)
		http.Error(w, "Failed to get bridge stats", http.StatusInternalServerError)
		return
	}

	component := BridgePerformance(stats)
	err = component.Render(r.Context(), w)
	if err != nil {
		s.logger.Error("failed to render bridge performance", "error", err)
		http.Error(w, "Failed to render bridge performance", http.StatusInternalServerError)
		return
	}
}

// handleUnmatchedDepositsSection handles the unmatched deposits section component
func (s *Server) handleUnmatchedDepositsSection(w http.ResponseWriter, r *http.Request) {
	page := 1
	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		parsedPage, err := strconv.Atoi(pageStr)
		if err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	offset := (page - 1) * ItemsPerPage

	deposits, err := GetUnmatchedDeposits(r.Context(), s.db, ItemsPerPage, offset)
	if err != nil {
		s.logger.Error("failed to get unmatched deposits", "error", err)
		http.Error(w, "Failed to get unmatched deposits", http.StatusInternalServerError)
		return
	}

	totalCount, err := GetTotalUnmatchedDeposits(r.Context(), s.db)
	if err != nil {
		s.logger.Error("failed to get total unmatched count", "error", err)
		http.Error(w, "Failed to get total unmatched count", http.StatusInternalServerError)
		return
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(ItemsPerPage)))

	component := UnmatchedDepositsSection(deposits, page, totalPages)
	err = component.Render(r.Context(), w)
	if err != nil {
		s.logger.Error("failed to render unmatched deposits section", "error", err)
		http.Error(w, "Failed to render unmatched deposits section", http.StatusInternalServerError)
		return
	}
}

// handleDepositsTimelineSection handles the deposits timeline section component
func (s *Server) handleDepositsTimelineSection(w http.ResponseWriter, r *http.Request) {
	page := 1
	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		parsedPage, err := strconv.Atoi(pageStr)
		if err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	offset := (page - 1) * ItemsPerPage

	deposits, err := GetMatchedDeposits(r.Context(), s.db, ItemsPerPage, offset)
	if err != nil {
		s.logger.Error("failed to get deposits", "error", err)
		http.Error(w, "Failed to get deposits", http.StatusInternalServerError)
		return
	}

	totalCount, err := GetTotalMatchedDeposits(r.Context(), s.db)
	if err != nil {
		s.logger.Error("failed to get total count", "error", err)
		http.Error(w, "Failed to get total count", http.StatusInternalServerError)
		return
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(ItemsPerPage)))

	component := DepositsTimelineSection(deposits, page, totalPages)
	err = component.Render(r.Context(), w)
	if err != nil {
		s.logger.Error("failed to render deposits timeline section", "error", err)
		http.Error(w, "Failed to render deposits timeline section", http.StatusInternalServerError)
		return
	}
}
