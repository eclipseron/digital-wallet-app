package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/eclipseron/digital-wallet-app/dto"
	"github.com/eclipseron/digital-wallet-app/middleware"
	"github.com/google/uuid"
)

func (c *Controller) GetAccountBalanceHandler(w http.ResponseWriter, r *http.Request) {
	var response dto.ResponseModel
	response.ID = uuid.New()
	response.Timestamp = time.Now().UTC()
	w.Header().Add("Content-Type", "application/json")

	param := r.PathValue("accountId")
	accountId, err := uuid.Parse(param)
	if err != nil {
		detail := err.Error()
		w.WriteHeader(http.StatusBadRequest)
		response.Data = dto.ErrorModel{Message: "invalid id", Details: []*string{&detail}}
		json.NewEncoder(w).Encode(&response)
		return
	}

	_uid, _ := r.Context().Value(middleware.USERID).(string)

	type Account struct {
		ID            uuid.UUID `json:"accountId"`
		UserId        uuid.UUID `json:"userId"`
		AccountNumber string    `json:"accountNumber"`
		Balance       int64     `json:"balance"`
		UpdatedAt     time.Time `json:"lastTransaction"`
	}

	var account Account

	tx := c.DB.Raw(`
	SELECT id, user_id, account_number, balance, updated_at
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
	response.Data = account
	json.NewEncoder(w).Encode(&response)
}
