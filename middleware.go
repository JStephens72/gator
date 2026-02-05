package main

import (
	"context"
	"fmt"

	"github.com/JStephens72/gator/internal/database"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		userName := s.cfg.CurrentUserName
		currentUser, err := s.db.GetUser(context.Background(), userName)
		if err != nil {
			return fmt.Errorf("error retrieving current user information: %w", err)
		}
		return handler(s, cmd, currentUser)
	}
}
