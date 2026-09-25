package api

import (
	"encoding/json"
	"errors"
	"github.com/eduardoArequipa/semantic-validator/internal/access"
	"io"
	"net/http"

	"github.com/eduardoArequipa/semantic-validator/internal/providers"
	"github.com/eduardoArequipa/semantic-validator/internal/validation"
)

type ValidateHandler struct {
	service *validation.Service
}

type CheckHandler struct {
	service *validation.Service
}

func NewValidateHandler(service *validation.Service) *ValidateHandler {
	return &ValidateHandler{service: service}
}

func NewCheckHandler(service *validation.Service) *CheckHandler {
	return &CheckHandler{service: service}
}

type validateRequest struct {
	Rule  string `json:"rule"`
	Value string `json:"value"`
}

type checkRequest struct {
	Value    string `json:"value"`
	Question string `json:"question"`
}

type batchRequest struct {
	Items []batchRequestItem `json:"items"`
}

type batchRequestItem struct {
	ID    string `json:"id"`
	Rule  string `json:"rule"`
	Value string `json:"value"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (h *ValidateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method must be POST")
		return
	}
	var request validateRequest
	if !decodeRequest(w, r, &request) {
		return
	}
	if request.Rule == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "rule is required")
		return
	}

	if !access.ReserveRequest(w, r, 1) {
		return
	}
	result, err := h.service.Validate(r.Context(), request.Rule, request.Value)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeResult(w, result)
}

func (h *CheckHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method must be POST")
		return
	}
	var request checkRequest
	if !decodeRequest(w, r, &request) {
		return
	}
	if !access.ReserveRequest(w, r, 1) {
		return
	}
	result, err := h.service.Check(r.Context(), request.Value, request.Question)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeResult(w, result)
}

type BatchHandler struct {
	service *validation.Service
}

func NewBatchHandler(service *validation.Service) *BatchHandler {
	return &BatchHandler{service: service}
}

func (h *BatchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method must be POST")
		return
	}
	var request batchRequest
	if !decodeRequest(w, r, &request) {
		return
	}
	if len(request.Items) == 0 || len(request.Items) > 100 {
		writeError(w, http.StatusBadRequest, "invalid_request", "items must contain between 1 and 100 elements")
		return
	}

	items := make([]validation.BatchItem, len(request.Items))
	seenIDs := make(map[string]struct{}, len(request.Items))
	for index, item := range request.Items {
		if item.ID == "" {
			writeError(w, http.StatusBadRequest, "invalid_request", "each item requires an id")
			return
		}
		if _, exists := seenIDs[item.ID]; exists {
			writeError(w, http.StatusBadRequest, "invalid_request", "item ids must be unique")
			return
		}
		seenIDs[item.ID] = struct{}{}
		items[index] = validation.BatchItem{ID: item.ID, Rule: item.Rule, Value: item.Value}
	}

	if !access.ReserveRequest(w, r, len(items)) {
		return
	}
	results := h.service.ValidateBatch(r.Context(), items)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(struct {
		Items []validation.BatchResult `json:"items"`
	}{Items: results})
}

func decodeRequest(w http.ResponseWriter, r *http.Request, destination any) bool {
	defer r.Body.Close()
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "request must be valid JSON")
		return false
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		writeError(w, http.StatusBadRequest, "invalid_request", "request must contain one JSON object")
		return false
	}
	return true
}

func writeServiceError(w http.ResponseWriter, err error) {
	if errors.Is(err, access.ErrAccounting) {
		writeError(w, http.StatusServiceUnavailable, "usage_unavailable", "usage storage unavailable")
		return
	}
	if errors.Is(err, validation.ErrInvalidValue) || errors.Is(err, validation.ErrInvalidQuestion) {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if errors.Is(err, validation.ErrUnknownRule) {
		writeError(w, http.StatusNotFound, "unknown_rule", "requested rule is not registered")
		return
	}
	if errors.Is(err, providers.ErrTimeout) {
		writeError(w, http.StatusGatewayTimeout, "provider_timeout", "semantic provider request timed out")
		return
	}
	writeError(w, http.StatusBadGateway, "provider_error", "semantic provider request failed")
}

func writeResult(w http.ResponseWriter, result validation.Result) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: errorBody{Code: code, Message: message}})
}
