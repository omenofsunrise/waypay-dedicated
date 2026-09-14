package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"

	"waypay_dedicated/internal/config"
	"waypay_dedicated/internal/kernel"
	"waypay_dedicated/internal/migrator"
	"waypay_dedicated/internal/modules/clients"
	"waypay_dedicated/internal/rbac"
	"waypay_dedicated/sdk"
)

func main() {
	logger := func(msg string, args ...any) {
		log.Println(append([]any{msg}, args...)...)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("pgx", cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(cfg.DBConnMaxLifetime)

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	rbacEngine := rbac.NewEngine()
	mig := migrator.New(db)
	k := kernel.New(rbacEngine, mig, db, logger)

	modules := []sdk.Module{
		clients.New(),
	}

	if err := k.LoadAll(modules); err != nil {
		log.Fatal(err)
	}

	if err := mig.ApplyAll(); err != nil {
		log.Fatal(err)
	}

	log.Println("listening on", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, k.BuildMux()); err != nil {
		log.Fatal(err)
	}
}
