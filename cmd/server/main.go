package main

import (
	"log"
	"net/http"
	"waypay_dedicated/internal/kernel"
	"waypay_dedicated/internal/modules/clients"
	"waypay_dedicated/internal/rbac"
)

func main() {
	logger := func(msg string, args ...any) { log.Println(append([]any{msg}, args...)...) }

	rbacEngine := rbac.NewEngine()
	k := kernel.New(rbacEngine, logger)

	if err := k.Register(clients.New()); err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	for _, r := range k.Routes() {
		handler := k.WithAuth(k.RequirePerm(r.RequiredPerm, r.Handler))
		mux.HandleFunc(r.Method+" "+r.Path, handler)
	}

	log.Println("listening on :8080")
	http.ListenAndServe(":8080", mux)
}
