package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/eclipseron/digital-wallet-app/conf"
	"github.com/eclipseron/digital-wallet-app/controller"
	"github.com/eclipseron/digital-wallet-app/dto"
	"github.com/eclipseron/digital-wallet-app/middleware"
	"github.com/eclipseron/digital-wallet-app/models"
	"github.com/eclipseron/digital-wallet-app/utils"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func TestWithdrawSuccess(t *testing.T) {
	godotenv.Load("../.env")
	db := conf.SetupDB()
	hash, _ := utils.CreateHash(TEST_PASSWORD)

	u := models.User{
		Email:    TEST_EMAIL,
		Password: hash,
	}
	db.Create(&u)

	acc := models.Account{
		UserID:        u.ID,
		Balance:       50000,
		AccountNumber: strconv.Itoa(int(time.Now().UnixMilli())),
	}
	db.Create(&acc)

	c := controller.NewController(db)
	srv := http.NewServeMux()

	srv.Handle("/api/v1/transaction/withdraw",
		middleware.RequireAuth(http.HandlerFunc(c.WithdrawHandler)))

	token, _ := utils.CreateJWT(u.ID)

	body := strings.NewReader(fmt.Sprintf(`{"amount": %d, "accountId":"%s"}`, 50000, acc.ID.String()))
	req := httptest.NewRequest("POST", "/api/v1/transaction/withdraw", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	type DataModel struct {
		AccountId     uuid.UUID `json:"accountId"`
		AccountNumber string    `json:"accountNumber"`
		Amount        int64     `json:"amount"`
		Type          string    `json:"type"`
		FinalBalance  int64     `json:"finalBalance"`
		At            time.Time `json:"at"`
	}

	type ResponseModel struct {
		ID        string    `json:"_id"`
		Data      DataModel `json:"data"`
		Timestamp time.Time `json:"timestamp"`
	}

	var res ResponseModel
	json.NewDecoder(w.Result().Body).Decode(&res)

	expectedBalance := acc.Balance - 50000
	if res.Data.FinalBalance != expectedBalance {
		t.Fatalf("expected %d, got %d", expectedBalance, res.Data.FinalBalance)
	}

	t.Cleanup(func() {
		db.Where("id = ?", acc.ID).Delete(&models.Account{})
		db.Where("id = ?", u.ID).Delete(&models.User{})
	})
}

func TestWithdrawInsufficientBalance(t *testing.T) {
	godotenv.Load("../.env")
	db := conf.SetupDB()
	hash, _ := utils.CreateHash(TEST_PASSWORD)

	u := models.User{
		Email:    TEST_EMAIL,
		Password: hash,
	}
	db.Create(&u)

	acc := models.Account{
		UserID:        u.ID,
		Balance:       50000,
		AccountNumber: strconv.Itoa(int(time.Now().UnixMilli())),
	}
	db.Create(&acc)

	c := controller.NewController(db)
	srv := http.NewServeMux()

	srv.Handle("/api/v1/transaction/withdraw",
		middleware.RequireAuth(http.HandlerFunc(c.WithdrawHandler)))

	token, _ := utils.CreateJWT(u.ID)

	body := strings.NewReader(fmt.Sprintf(`{"amount": %d, "accountId":"%s"}`, 100000, acc.ID.String()))
	req := httptest.NewRequest("POST", "/api/v1/transaction/withdraw", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	type ResponseModel struct {
		ID        string         `json:"_id"`
		Data      dto.ErrorModel `json:"data"`
		Timestamp time.Time      `json:"timestamp"`
	}

	var res ResponseModel
	json.NewDecoder(w.Result().Body).Decode(&res)

	if res.Data.Message != "insufficient balance" {
		t.Fatalf("expected %s, got %s", "insufficient balance", res.Data.Message)
	}

	t.Cleanup(func() {
		db.Where("id = ?", acc.ID).Delete(&models.Account{})
		db.Where("id = ?", u.ID).Delete(&models.User{})
	})
}
