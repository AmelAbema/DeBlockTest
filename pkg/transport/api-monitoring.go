package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/tel-io/tel/v2"
)

type addressProvider interface {
	GetAddressCount() int
}

type processingProvider interface {
	GetLastProcessedBlock(ctx context.Context) (uint64, error)
}

type MonitoringAPI struct {
	addresses  addressProvider
	processing processingProvider
}

func NewMonitoringAPI(addresses addressProvider, processing processingProvider) *MonitoringAPI {
	return &MonitoringAPI{
		addresses:  addresses,
		processing: processing,
	}
}

func (api *MonitoringAPI) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/health", api.handleHealthCheck)
	mux.HandleFunc("/api/v1/stats", api.handleStats)
	mux.HandleFunc("/api/v1/addresses/count", api.handleAddressCount)
	mux.HandleFunc("/api/v1/monitoring/status", api.handleMonitoringStatus)
}

func (api *MonitoringAPI) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":  "healthy",
		"service": "deblock-monitoring",
		"version": "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (api *MonitoringAPI) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	lastBlock, err := api.processing.GetLastProcessedBlock(ctx)
	if err != nil {
		tel.Global().Error("failed to get last processed block", tel.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	stats := map[string]interface{}{
		"monitored_addresses":  api.addresses.GetAddressCount(),
		"last_processed_block": lastBlock,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (api *MonitoringAPI) handleAddressCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	count := api.addresses.GetAddressCount()

	response := map[string]interface{}{
		"address_count": count,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (api *MonitoringAPI) handleMonitoringStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()

	lastBlock, err := api.processing.GetLastProcessedBlock(ctx)
	if err != nil {
		tel.Global().Error("failed to get monitoring status", tel.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	status := map[string]interface{}{
		"status":               "monitoring",
		"last_processed_block": lastBlock,
		"monitored_addresses":  api.addresses.GetAddressCount(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
