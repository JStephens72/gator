# gator
Welcome to gator.

## Prerequisites:

You will need to install **PostgreSQL** and **Go** in order to build and run **gator**.

## Install:

Run  
`go install https://github.com/JStephens72/gator`

## Configuration:

The configuration is a simple JSON file stored in your home directory.

~/.gatorconfig.json
<pre>
{
  "db_url": "database connection string",
}
</pre>
<pre>
NAME
	gator

SYNOPSIS
	gator OPTION [parameters...]

DESCRIPTION
	Manage and store RSS newsfeeds.
	
	login <username>
		Change the current username.
	
	register <username>
		Add a new username to the database.
	
	reset
		Removes all users from the database. All associated
		feeds and posts will also be removed.
	
	users
		Lists all currently registered users.
	
	agg <interval>
		Puts gator in a perpetual loop checking for subscribed feed updates.
		Examples of the interval are 30s, 45m, 4h, etc.
		At each interval, only the oldest unupdated feed is retrieved, so if
		three feeds, and your interval is 1h, then all feeds will be updated in
		2 hours (agg updates the first feed immediately.
		
	addfeed <name> <url>
		Adds a feed to the database and automatically subscribes the 
		current user to the feed.
		
	feeds
		Lists all the feeds in the database.
		
	follow <name>
		Follows the specified feed, assuming the feed is already in the
		database.
		
	following
		Lists all feeds subscribed by the current user.
		
	unfollow <name>
		Stops following the specified feed for the current user.
		
	browse <limit>; default 2
		Displays posts that have been retrieved and stored. <limit> specifies
		the number of posts to display. Default is 2.
</pre>
