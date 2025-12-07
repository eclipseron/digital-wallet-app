package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/eclipseron/digital-wallet-app/dto"
	"github.com/eclipseron/digital-wallet-app/models"
	"github.com/eclipseron/digital-wallet-app/utils"
	"github.com/google/uuid"
)

func (c *Controller) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var response dto.ResponseModel
	response.ID = uuid.New()
	response.Timestamp = time.Now().UTC()
	w.Header().Add("Content-Type", "application/json")

	type RequestModel struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var payload RequestModel
	json.NewDecoder(r.Body).Decode(&payload)

	if payload.Email == "" || payload.Password == "" || payload.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		response.Data = dto.ErrorModel{Message: "email, password, and name is required"}
		json.NewEncoder(w).Encode(&response)
		return
	}

	var _uid uuid.UUID
	tx := c.DB.Raw(`
	SELECT id FROM users WHERE email = ?
	`, payload.Email).Scan(&_uid)
	if tx.RowsAffected > 0 {
		w.WriteHeader(http.StatusConflict)
		response.Data = dto.ErrorModel{Message: "email already registered"}
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

	hash, err := utils.CreateHash(payload.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response.Data = dto.ErrorModel{Message: "failed to hash password"}
		json.NewEncoder(w).Encode(&response)
		return
	}

	user := models.User{
		Name:     payload.Name,
		Email:    payload.Email,
		Password: hash,
	}
	if err := c.DB.Create(&user).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response.Data = dto.ErrorModel{Message: "failed to create user"}
		json.NewEncoder(w).Encode(&response)
		return
	}

	type ResponseModel struct {
		UserID        uuid.UUID `json:"userId"`
		AccountID     uuid.UUID `json:"accountId"`
		AccountNumber string    `json:"accountNumber"`
		Email         string    `json:"email"`
	}

	account := models.Account{
		UserID:        user.ID,
		AccountNumber: strconv.Itoa(int(time.Now().UnixMicro())),
		Balance:       0,
	}
	if err := c.DB.Create(&account).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response.Data = dto.ErrorModel{Message: "failed to create account"}
		json.NewEncoder(w).Encode(&response)
		return
	}

	response.Data = ResponseModel{
		UserID:        user.ID,
		AccountID:     account.ID,
		AccountNumber: account.AccountNumber,
		Email:         user.Email,
	}
	json.NewEncoder(w).Encode(&response)
}

func (c *Controller) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var response dto.ResponseModel
	response.ID = uuid.New()
	response.Timestamp = time.Now().UTC()
	w.Header().Add("Content-Type", "application/json")

	type RequestModel struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var payload RequestModel
	json.NewDecoder(r.Body).Decode(&payload)

	if payload.Email == "" || payload.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		response.Data = dto.ErrorModel{Message: "email and password required"}
		json.NewEncoder(w).Encode(&response)
		return
	}

	var user models.User
	tx := c.DB.Raw(`
	SELECT id, email, password FROM users WHERE email = ?
	`, payload.Email).Scan(&user)
	if tx.RowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		response.Data = dto.ErrorModel{Message: "email not found"}
		json.NewEncoder(w).Encode(&response)
		return
	}
	if tx.Error != nil {
		detail := tx.Error.Error()
		w.WriteHeader(http.StatusInternalServerError)
		response.Data = dto.ErrorModel{Message: "failed to get user", Details: []*string{&detail}}
		json.NewEncoder(w).Encode(&response)
		return
	}

	if !utils.IsValid(user.Password, payload.Password) {
		w.WriteHeader(http.StatusUnauthorized)
		response.Data = dto.ErrorModel{Message: "invalid email or password"}
		json.NewEncoder(w).Encode(&response)
		return
	}

	token, err := utils.CreateJWT(user.ID)
	if err != nil {
		detail := tx.Error.Error()
		w.WriteHeader(http.StatusInternalServerError)
		response.Data = dto.ErrorModel{Message: "failed to generate token", Details: []*string{&detail}}
		json.NewEncoder(w).Encode(&response)
		return
	}

	type LoginResponseModel struct {
		UserID string `json:"user_id"`
		Email  string `json:"email"`
		Token  string `json:"token"`
	}
	response.Data = LoginResponseModel{
		UserID: user.ID.String(),
		Email:  user.Email,
		Token:  token,
	}
	json.NewEncoder(w).Encode(&response)
}
