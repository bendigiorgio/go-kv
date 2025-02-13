package main

import (
	"strconv"

	"github.com/bendigiorgio/go-kv/internal/api"
	"github.com/bendigiorgio/go-kv/internal/engine"
	"github.com/bendigiorgio/go-kv/internal/utils"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := utils.LoadConfig()
	log.Debug().Msgf("Loaded config: %+v", cfg)
	if err != nil {
		log.Panic().Err(err)
	}
	utils.SetupLogger(cfg)
	e, err := engine.NewEngine(cfg.Database.FilePath, cfg.Database.FlushFilePath, cfg.Database.MaxMemory)
	if err != nil {
		log.Panic().Err(err)
	}
	router := api.NewRouter(e, true)
	router.Start(strconv.Itoa(cfg.AppPort))

}
