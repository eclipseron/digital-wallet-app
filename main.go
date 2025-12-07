package main

import (
	"log"
	"net/http"
	"time"

	"github.com/eclipseron/digital-wallet-app/conf"
	"github.com/eclipseron/digital-wallet-app/controller"
	"github.com/eclipseron/digital-wallet-app/middleware"
	"github.com/eclipseron/digital-wallet-app/migrations"
)

func main() {
	db := conf.SetupDB()
	migrations.Run(db)

	c := controller.NewController(db)

	s := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  20 * time.Second,
	}

	http.Handle("GET /api/v1/accounts/{accountId}/balance",
		middleware.RequireAuth(http.HandlerFunc(c.GetAccountBalanceHandler)))
	http.Handle("POST /api/v1/transaction/withdraw",
		middleware.RequireAuth(http.HandlerFunc(c.WithdrawHandler)))
	http.HandleFunc("POST /api/v1/register", c.RegisterHandler)
	http.HandleFunc("POST /api/v1/login", c.LoginHandler)

	if err := s.ListenAndServe(); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
