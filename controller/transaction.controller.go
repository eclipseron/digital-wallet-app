package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/eclipseron/digital-wallet-app/dto"
	"github.com/eclipseron/digital-wallet-app/middleware"
	"github.com/eclipseron/digital-wallet-app/models"
	"github.com/google/uuid"
)

func (c *Controller) WithdrawHandler(w http.ResponseWriter, r *http.Request) {
	var response dto.ResponseModel
	response.ID = uuid.New()
	response.Timestamp = time.Now().UTC()
	w.Header().Add("Content-Type", "application/json")

	_uid, _ := r.Context().Value(middleware.USERID).(string)

	type RequestModel struct {
		Amount    int64  `json:"amount"`
		AccountID string `json:"accountId"`
	}
	var payload RequestModel
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		detail := err.Error()
		w.WriteHeader(http.StatusBadRequest)
		response.Data = dto.ErrorModel{Message: "invalid request body", Details: []*string{&detail}}
		json.NewEncoder(w).Encode(&response)
		return
	}

	if payload.Amount <= 0 {
		detail := "withdraw amount must be positive"
		w.WriteHeader(http.StatusBadRequest)
		response.Data = dto.ErrorModel{Message: "invalid request body", Details: []*string{&detail}}
		json.NewEncoder(w).Encode(&response)
		return
	}
	if payload.AccountID == "" {
		detail := "accountId is required"
		w.WriteHeader(http.StatusBadRequest)
		response.Data = dto.ErrorModel{Message: "invalid request body", Details: []*string{&detail}}
		json.NewEncoder(w).Encode(&response)
		return
	}

	accountId, err := uuid.Parse(payload.AccountID)
	if err != nil {
		detail := err.Error()
		w.WriteHeader(http.StatusBadRequest)
		response.Data = dto.ErrorModel{Message: "invalid id", Details: []*string{&detail}}
		json.NewEncoder(w).Encode(&response)
		return
	}

	type Account struct {
		ID            uuid.UUID
		UserId        uuid.UUID
		AccountNumber string
		Balance       int64
	}

	var account Account

	tx := c.DB.Raw(`
	SELECT id, user_id, account_number, balance
	FROM accounts WHERE id = ?
	`, accountId.String()).Scan(&account)
	if tx.RowsAffected == 0 {
		detail := fmt.Sprintf("account with id: %s not exist", accountId.String())
		w.WriteHeader(http.StatusNotFound)
		response.Data = dto.ErrorModel{Message: "account not found", Details: []*string{&detail}}
		json.NewEncoder(w).Encode(&response)
		return
	}
	if tx.Error != nil {
		detail := tx.Error.Error()
		w.WriteHeader(http.StatusInternalServerError)
		response.Data = dto.ErrorModel{Message: "an error occured", Details: []*string{&detail}}
		json.NewEncoder(w).Encode(&response)
		return
	}
	if _uid != account.UserId.String() {
		detail := "This account does not belong to the user"
		w.WriteHeader(http.StatusForbidden)
		response.Data = dto.ErrorModel{Message: "forbidden", Details: []*string{&detail}}
		json.NewEncoder(w).Encode(&response)
		return
	}
	if account.Balance < payload.Amount {
		w.WriteHeader(http.StatusBadRequest)
		response.Data = dto.ErrorModel{Message: "insufficient balance"}
		json.NewEncoder(w).Encode(&response)
		return
	}

	account.Balance -= payload.Amount
	tx = c.DB.Begin()
	if err := tx.Save(&account).Error; err != nil {
		tx.Rollback()
		detail := err.Error()
		w.WriteHeader(http.StatusInternalServerError)
		response.Data = dto.ErrorModel{Message: "failed to update balance", Details: []*string{&detail}}
		json.NewEncoder(w).Encode(&response)
		return
	}

	desc := "ATM Cash Withdrawal"
	accTx := models.Transactions{
		AccountID:   account.ID,
		Amount:      -payload.Amount,
		Type:        "WITHDRAW",
		Description: &desc,
	}
	if err := tx.Create(&accTx).Error; err != nil {
		tx.Rollback()
		detail := err.Error()
		w.WriteHeader(http.StatusInternalServerError)
		response.Data = dto.ErrorModel{Message: "failed to create transaction", Details: []*string{&detail}}
		json.NewEncoder(w).Encode(&response)
		return
	}

	if err := tx.Commit().Error; err != nil {
		detail := err.Error()
		w.WriteHeader(http.StatusInternalServerError)
		response.Data = dto.ErrorModel{Message: "database error", Details: []*string{&detail}}
		json.NewEncoder(w).Encode(&response)
		return
	}

	type WithdrawalResponseModel struct {
		AccountId     uuid.UUID `json:"accountId"`
		AccountNumber string    `json:"accountNumber"`
		Amount        int64     `json:"amount"`
		Type          string    `json:"type"`
		FinalBalance  int64     `json:"finalBalance"`
		At            time.Time `json:"at"`
	}

	response.Data = WithdrawalResponseModel{
		AccountId:     account.ID,
		AccountNumber: account.AccountNumber,
		Amount:        payload.Amount,
		Type:          accTx.Type,
		FinalBalance:  account.Balance,
		At:            accTx.CreatedAt.UTC(),
	}
	json.NewEncoder(w).Encode(&response)
}
