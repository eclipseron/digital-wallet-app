package main

import (
	"log"
	"net/http"
	"time"

	"github.com/eclipseron/digital-wallet-app/conf"
	"github.com/eclipseron/digital-wallet-app/controller"
)

func main() {
	db := conf.SetupDB()
	// migrations.Run(db)

	c := controller.NewController(db)

	s := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 20 * time.Second,
		IdleTimeout:  20 * time.Second,
	}

	http.HandleFunc("GET /api/v1/accounts/{accountId}/balance", c.GetAccountBalanceHandler)
	http.HandleFunc("POST /api/v1/register", c.RegisterHandler)
	http.HandleFunc("POST /api/v1/login", c.LoginHandler)

	if err := s.ListenAndServe(); err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
