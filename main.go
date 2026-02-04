package main

import (
	"fmt"
	"os"

	"github.com/JStephens72/gator/internal/config"
)

func main() {

	cfg, err := config.Read()
	if err != nil {
		fmt.Print(fmt.Errorf("Error reading configuration file: %w", err))
		os.Exit(1)
	}

	cfg.SetUser("jstephens")

	cfg, err = config.Read()
	fmt.Print(cfg)

}
