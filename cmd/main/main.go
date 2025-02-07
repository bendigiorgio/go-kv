package main

import (
	"github.com/bendigiorgio/go-kv/internal/api"
	"github.com/bendigiorgio/go-kv/internal/engine"
)

func main() {
	e := engine.NewEngine("data.db", "flush.db", 1024*1024)
	router := api.NewRouter(e, true)
	router.Start("8080")

}
