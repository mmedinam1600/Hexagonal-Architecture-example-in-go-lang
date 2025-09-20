package inhttp

import (
	"encoding/json"
	"errors"
	"hexagonal-bank/internal/core/application/ports"
	"hexagonal-bank/internal/core/application/usecase"
	"hexagonal-bank/internal/core/domain"
	"hexagonal-bank/internal/platform/logging"
	"hexagonal-bank/internal/shared/httpx"
	"net/http"
	"strings"
)

// Handlers are thin: translate HTTP <-> UseCase DTOs.

type API struct {
	logger               logging.Logger
	openAccountUseCase   *usecase.OpenAccountUseCase
	depositMoneyUseCase  *usecase.DepositMoneyUseCase
	transferMoneyUseCase *usecase.TransferMoneyUseCase
}

func NewAPI(
	logger logging.Logger,
	accountReader ports.AccountReader,
	accountWriter ports.AccountWriter,
	paymentGateway ports.PaymentGateway,
	eventPublisher ports.EventPublisher,
) *API {
	return &API{
		logger:               logger,
		openAccountUseCase:   usecase.NewOpenAccountUseCase(accountWriter),
		depositMoneyUseCase:  usecase.NewDepositMoneyUseCase(accountReader, accountWriter),
		transferMoneyUseCase: usecase.NewTransferMoneyUseCase(accountReader, accountWriter, paymentGateway, eventPublisher),
	}
}

func (api *API) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", api.health)
	mux.HandleFunc("/accounts", api.handleAccounts)       // POST
	mux.HandleFunc("/accounts/", api.handleAccountDetail) // GET /:id, POST /:id/deposit
	mux.HandleFunc("/transfers", api.transfer)            // POST
	return mux
}

func (api *API) health(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// /accounts -> POST
func (api *API) handleAccounts(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		api.createAccount(w, r)
		return
	}
	httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
}

// /accounts/{id} (GET) or /accounts/{id}/deposit (POST)
func (api *API) handleAccountDetail(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/accounts/")
	parts := strings.Split(path, "/")
	accountID := parts[0]
	if accountID == "" {
		httpx.WriteError(w, http.StatusBadRequest, "missing id")
		return
	}
	if len(parts) == 1 && r.Method == http.MethodGet {
		api.getAccount(w, r, accountID)
		return
	}
	if len(parts) == 2 && parts[1] == "deposit" && r.Method == http.MethodPost {
		api.deposit(w, r, accountID)
		return
	}
	httpx.WriteError(w, http.StatusNotFound, "route not found")
}

type createAccountRequest struct {
	HolderName string `json:"holder_name"`
	CLABE      string `json:"clabe"`
}

func (api *API) createAccount(w http.ResponseWriter, r *http.Request) {
	var requestBody createAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}
	output, err := api.openAccountUseCase.Execute(r.Context(), usecase.OpenAccountInput{
		HolderName: requestBody.HolderName,
		CLABE:      requestBody.CLABE,
	})
	if err != nil {
		api.mapDomainErr(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, output)
}

func (api *API) getAccount(w http.ResponseWriter, r *http.Request, accountID string) {
	account, err := api.depositMoneyUseCase.ReaderPort().ByID(r.Context(), accountID)
	if err != nil {
		httpx.WriteError(w, http.StatusNotFound, "account not found")
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"id": account.ID, "holder_name": account.HolderName(), "clabe": account.CLABE(), "balance_cents": account.Balance(),
	})
}

type depositRequest struct {
	Cents int64 `json:"cents"`
}

func (api *API) deposit(w http.ResponseWriter, r *http.Request, accountID string) {
	var requestBody depositRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}
	output, err := api.depositMoneyUseCase.Execute(r.Context(), usecase.DepositInput{
		AccountID: accountID,
		Cents:     requestBody.Cents,
	})
	if err != nil {
		api.mapDomainErr(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, output)
}

type transferRequest struct {
	FromID string `json:"from_id"`
	ToID   string `json:"to_id"`
	Cents  int64  `json:"cents"`
}

func (api *API) transfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpx.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var requestBody transferRequest
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}
	output, err := api.transferMoneyUseCase.Execute(r.Context(), usecase.TransferInput{
		FromID: strings.TrimSpace(requestBody.FromID),
		ToID:   strings.TrimSpace(requestBody.ToID),
		Cents:  requestBody.Cents,
	})
	if err != nil {
		api.mapDomainErr(w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusAccepted, output)
}

// Map domain errors to HTTP responses (adapter concern).
func (api *API) mapDomainErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidAmount):
		httpx.WriteError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrInsufficientFund):
		httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, domain.ErrInvalidCLABE), errors.Is(err, domain.ErrEmptyHolder):
		httpx.WriteError(w, http.StatusUnprocessableEntity, err.Error())
	default:
		api.logger.Error("unexpected error", "err", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal error")
	}
}
