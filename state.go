package main

import (
	"github.com/JStephens72/gator/internal/config"
	"github.com/JStephens72/gator/internal/database"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}
