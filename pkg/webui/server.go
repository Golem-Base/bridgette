package webui

//go:generate go install github.com/a-h/templ/cmd/templ@latest
//go:generate templ generate

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/Golem-Base/bridgette/pkg/sqlitestore"
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

	// API endpoints
	mux.HandleFunc("GET /api/chart-data", s.handleTimeSeriesData)

	// Static files
	mux.Handle("GET /static/", http.StripPrefix("/static/", createStaticHandler()))

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

// handleTimeSeriesData handles the API endpoint for chart data
func (s *Server) handleTimeSeriesData(w http.ResponseWriter, r *http.Request) {
	// Get the limit parameter (default to 20)
	limitStr := r.URL.Query().Get("limit")
	limit := 30
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Create a queries object using the sqlitestore package
	queries := sqlitestore.New(s.db)

	// Get the time series chart data
	data, err := queries.GetTimeSeriesChartData(r.Context(), int64(limit))
	if err != nil {
		s.logger.Error("failed to query deposit time differences", "error", err)
		http.Error(w, "Failed to query deposit time differences", http.StatusInternalServerError)
		return
	}

	// Transform data for the chart
	type chartDataPoint struct {
		Timestamp       string  `json:"timestamp"`
		TimeDiffSeconds float64 `json:"timeDiffSeconds"`
	}

	chartData := make([]chartDataPoint, 0, len(data))
	for _, point := range data {
		// Convert unix timestamp to a readable date/time
		t := time.Unix(point.Timestamp, 0)

		// Use seconds directly without converting to hours
		var timeDiffSeconds float64
		timeDiffSecs, ok := point.TimeDiffSeconds.(int64)
		if !ok {
			// Try as float64
			if timeDiffSecsFloat, ok := point.TimeDiffSeconds.(float64); ok {
				timeDiffSeconds = timeDiffSecsFloat
			} else {
				// Default to 0 if conversion fails
				s.logger.Warn("unexpected type for time_diff_seconds", "type", fmt.Sprintf("%T", point.TimeDiffSeconds))
				timeDiffSeconds = 0
			}
		} else {
			timeDiffSeconds = float64(timeDiffSecs)
		}

		chartData = append(chartData, chartDataPoint{
			Timestamp:       t.Format(time.RFC3339),
			TimeDiffSeconds: timeDiffSeconds,
		})
	}

	// Set content type and encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(chartData)
	if err != nil {
		s.logger.Error("failed to encode chart data as JSON", "error", err)
		http.Error(w, "Failed to encode chart data", http.StatusInternalServerError)
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
