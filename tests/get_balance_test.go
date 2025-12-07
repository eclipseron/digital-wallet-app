package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/eclipseron/digital-wallet-app/conf"
	"github.com/eclipseron/digital-wallet-app/controller"
	"github.com/eclipseron/digital-wallet-app/middleware"
	"github.com/eclipseron/digital-wallet-app/models"
	"github.com/eclipseron/digital-wallet-app/utils"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

var TEST_PASSWORD string = "password123"
var TEST_EMAIL string = "unnamed@test.com"

func TestGetBalanceSuccess(t *testing.T) {
	/* Test for success call */
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
	srv.Handle("/api/v1/accounts/{accountId}/balance",
		middleware.RequireAuth(http.HandlerFunc(c.GetAccountBalanceHandler)))

	target := fmt.Sprintf("/api/v1/accounts/%s/balance", acc.ID.String())
	req := httptest.NewRequest("GET", target, nil)
	token, _ := utils.CreateJWT(u.ID)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	type DataModel struct {
		AccountID       string    `json:"accountId"`
		UserID          string    `json:"userId"`
		AccountNumber   string    `json:"accountNumber"`
		Balance         int64     `json:"balance"`
		LastTransaction time.Time `json:"lastTransaction"`
	}

	type ResponseModel struct {
		ID        string    `json:"_id"`
		Data      DataModel `json:"data"`
		Timestamp time.Time `json:"timestamp"`
	}

	var res ResponseModel
	json.NewDecoder(w.Result().Body).Decode(&res)
	if res.Data.Balance != acc.Balance {
		t.Fatalf("expected %d, got %d", acc.Balance, res.Data.Balance)
	}

	t.Cleanup(func() {
		db.Where("id = ?", acc.ID).Delete(&models.Account{})
		db.Where("id = ?", u.ID).Delete(&models.User{})
	})
}

func TestGetBalanceUnauthorized(t *testing.T) {
	/* Test for missing authorization header */
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
	srv.Handle("/api/v1/accounts/{accountId}/balance",
		middleware.RequireAuth(http.HandlerFunc(c.GetAccountBalanceHandler)))

	target := fmt.Sprintf("/api/v1/accounts/%s/balance", acc.ID.String())
	req := httptest.NewRequest("GET", target, nil)

	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}

	t.Cleanup(func() {
		db.Where("id = ?", acc.ID).Delete(&models.Account{})
		db.Where("id = ?", u.ID).Delete(&models.User{})
	})
}

func TestGetBalanceAccountNotFound(t *testing.T) {
	/* Test for missing authorization header */
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
	srv.Handle("/api/v1/accounts/{accountId}/balance",
		middleware.RequireAuth(http.HandlerFunc(c.GetAccountBalanceHandler)))

	target := fmt.Sprintf("/api/v1/accounts/%s/balance", uuid.New().String())
	req := httptest.NewRequest("GET", target, nil)
	token, _ := utils.CreateJWT(u.ID)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}

	t.Cleanup(func() {
		db.Where("id = ?", acc.ID).Delete(&models.Account{})
		db.Where("id = ?", u.ID).Delete(&models.User{})
	})
}
