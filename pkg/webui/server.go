package webui

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
	mux.HandleFunc("GET /deposits", s.handleDeposits)
	mux.HandleFunc("GET /dashboard", s.handleDashboard)

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
	stats, err := GetBridgeStats(r.Context(), s.db)
	if err != nil {
		s.logger.Error("failed to get bridge stats", "error", err)
		http.Error(w, "Failed to get bridge stats", http.StatusInternalServerError)
		return
	}

	component := Dashboard(stats)
	err = component.Render(r.Context(), w)
	if err != nil {
		s.logger.Error("failed to render dashboard", "error", err)
		http.Error(w, "Failed to render dashboard", http.StatusInternalServerError)
		return
	}
}

// handleDeposits handles the deposits endpoint
func (s *Server) handleDeposits(w http.ResponseWriter, r *http.Request) {
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

	component := DepositsList(deposits, page, totalPages)
	err = component.Render(r.Context(), w)
	if err != nil {
		s.logger.Error("failed to render deposits list", "error", err)
		http.Error(w, "Failed to render deposits list", http.StatusInternalServerError)
		return
	}
}

// handleDashboard handles the dashboard content updates for auto-refresh
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	stats, err := GetBridgeStats(r.Context(), s.db)
	if err != nil {
		s.logger.Error("failed to get bridge stats", "error", err)
		http.Error(w, "Failed to get bridge stats", http.StatusInternalServerError)
		return
	}

	// Only render the dashboard content component, not the full layout
	err = DashboardContent(stats).Render(r.Context(), w)
	if err != nil {
		s.logger.Error("failed to render dashboard content", "error", err)
		http.Error(w, "Failed to render dashboard content", http.StatusInternalServerError)
		return
	}
}
