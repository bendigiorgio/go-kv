package utils

import (
	"os"

	"github.com/nil-go/konf"
	"github.com/nil-go/konf/provider/fs"
	"github.com/rs/zerolog/log"
)

type DatabaseConfig struct {
	FilePath      string `default:"./db/data.db" usage:"Path for the main database file"`
	FlushFilePath string `default:"./db/flush.db" usage:"Path for the flush database file"`
	MaxMemory     int    `default:"5242880" usage:"Maximum memory to use for the database"`
}

type ConfigStructure struct {
	AppPort   int    `default:"8080" usage:"Port to run the application on"`
	LogLevel  int8   `default:"-1" usage:"Log level for the application"`
	LogFile   string `default:"./logs/app.log" usage:"Path for the log file of the application"`
	LogOutput string `default:"console" usage:"Output for the logs (console, file)"`
	Database  DatabaseConfig
}

func LoadConfig() (*ConfigStructure, error) {
	var config konf.Config

	cfg := ConfigStructure{
		AppPort:  8080,
		LogLevel: -1,
		LogFile:  "./logs/app.log",
		Database: DatabaseConfig{
			FilePath:      "./db/data.db",
			FlushFilePath: "./db/flush.db",
			MaxMemory:     5242880,
		},
	}
	err := config.Load(fs.New(os.DirFS("kv-setup.json"), "kv-setup.json"))

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	return &cfg, nil
}
