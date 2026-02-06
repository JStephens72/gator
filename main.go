package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/JStephens72/gator/internal/config"
	"github.com/JStephens72/gator/internal/database"
	_ "github.com/lib/pq"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Println("insufficient arguments")
		os.Exit(1)
	}

	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("Error reading configuration file: %v\n", err)
		os.Exit(1)
	}

	st := &state{
		cfg: &cfg,
	}

	dbURL := st.cfg.DbURL
	db, err := sql.Open("postgres", dbURL)
	dbQueries := database.New(db)
	st.db = dbQueries

	cmds := commands{
		handlers: make(map[string]func(*state, command) error),
	}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerResetDatabase)
	cmds.register("users", handlerListUsers)
	cmds.register("agg", handlerAggregate)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerListFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cmds.register("browse", handlerBrowse)

	cmd := command{
		name: os.Args[1],
		args: os.Args[2:],
	}

	if err := cmds.run(st, cmd); err != nil {
		fmt.Printf("error running %s command: %v\n", cmd.name, err)
		os.Exit(1)
	}
}
