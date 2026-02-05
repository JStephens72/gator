package main

import (
	"context"
	"fmt"
	"time"

	"github.com/JStephens72/gator/internal/database"
	"github.com/google/uuid"
)

type command struct {
	name string
	args []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.handlers[cmd.name]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmd.name)
	}

	return handler(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("username is required")
	}

	userName := cmd.args[0]
	user, err := s.db.GetUser(context.Background(), userName)
	if err != nil {
		return fmt.Errorf("user '%s' not found", userName)
	}

	if err := s.cfg.SetUser(user.Name); err != nil {
		return fmt.Errorf("error setting the username: %w", err)
	}
	fmt.Printf("user set to %s\n", userName)

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("user name is required")
	}

	userParams := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
	}
	user, err := s.db.CreateUser(context.Background(), userParams)
	if err != nil {
		return err
	}

	s.cfg.SetUser(user.Name)
	fmt.Printf("user %s created\n", user.Name)
	return nil
}

func handlerResetDatabase(s *state, cmd command) error {
	if err := s.db.ResetDatabase(context.Background()); err != nil {
		return fmt.Errorf("error resetting database: %w", err)
	}
	fmt.Println("Database reset successfully")

	return nil
}

func handlerListUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error getting user list: %w", err)
	}

	for _, user := range users {
		if user.Name == s.cfg.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}

	return nil
}

func handlerAggregate(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("fetch interval is required")
	}
	time_between_reqs := cmd.args[0]
	timeBetweenRequests, err := time.ParseDuration(time_between_reqs)
	if err != nil {
		return fmt.Errorf("error parsing interval '%s': %w", time_between_reqs, err)
	}
	fmt.Printf("Collecting feeds every %s\n", timeBetweenRequests.String())

	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		if err := scrapeFeeds(s); err != nil {
			return fmt.Errorf("error checking for updates: %w", err)
		}
	}
}

func handlerAddFeed(s *state, cmd command, currentUser database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("feed name AND url are required")
	}
	feedName := cmd.args[0]
	feedUrl := cmd.args[1]

	feedParams := database.AddFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feedName,
		Url:       feedUrl,
		UserID:    currentUser.ID,
	}

	feed, err := s.db.AddFeed(context.Background(), feedParams)
	if err != nil {
		return fmt.Errorf("error adding feed to db: %w", err)
	}

	fmt.Printf("user %s has added feed %s\n", currentUser.Name, feed.Name)

	// now add a follow
	followParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    currentUser.ID,
		FeedID:    feed.ID,
	}

	feedFollow, err := s.db.CreateFeedFollow(context.Background(), followParams)
	if err != nil {
		return fmt.Errorf("error creating follow of %s for user %s", feed.Name, currentUser.Name)
	}

	fmt.Printf("user %s is following feed %s\n", feedFollow.UserName, feedFollow.FeedName)

	return nil
}

func handlerListFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error getting feeds: %w", err)
	}

	for _, feed := range feeds {
		user, err := s.db.GetUserById(context.Background(), feed.UserID)
		if err != nil {
			return fmt.Errorf("error retrieving user by ID: %w", err)
		}
		userName := user.Name
		fmt.Printf("Name:     %s\n", feed.Name)
		fmt.Printf("Url:      %s\n", feed.Url)
		fmt.Printf("Added by: %s\n", userName)
		fmt.Println("======================================")
	}

	return nil
}

func handlerFollow(s *state, cmd command, currentUser database.User) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("url to follow is required")
	}
	url := cmd.args[0]

	feed, err := s.db.GetFeedByUrl(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("invalid url (%s): %w\n", url, err)
	}

	feedFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    currentUser.ID,
		FeedID:    feed.ID,
	}

	feedFollowRow, err := s.db.CreateFeedFollow(context.Background(), feedFollowParams)
	if err != nil {
		return fmt.Errorf("failed to follow the RSS feed at %s: %w", url, err)
	}

	fmt.Printf("user %s is following feed %s\n", feedFollowRow.UserName, feedFollowRow.FeedName)

	return nil
}

func handlerFollowing(s *state, cmd command, currentUser database.User) error {
	followedFeeds, err := s.db.GetFeedFollowsForUser(context.Background(), currentUser.ID)
	if err != nil {
		return fmt.Errorf("error retrieving feeds for user %s: %w", currentUser.Name, err)
	}

	if len(followedFeeds) == 0 {
		fmt.Printf("user %s follows no feeds\n", currentUser.Name)
	} else {
		fmt.Printf("user %s is following:\n", currentUser.Name)
	}
	for _, feeds := range followedFeeds {
		fmt.Printf("- '%s'\n", feeds.FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, currentUser database.User) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("feed url is required")
	}
	url := cmd.args[0]

	feed, err := s.db.GetFeedByUrl(context.Background(), url)
	if err != nil {
		return fmt.Errorf("error getting feed %s: %w", url, err)
	}

	params := database.DeleteFeedFollowParams{
		UserID: currentUser.ID,
		FeedID: feed.ID,
	}
	err = s.db.DeleteFeedFollow(context.Background(), params)
	if err != nil {
		return fmt.Errorf("error unfollowing feed %s for user %s: %w", feed.Name, currentUser.Name, err)
	}

	fmt.Printf("user %s stopped following '%s'", currentUser.Name, feed.Name)
	return nil
}
