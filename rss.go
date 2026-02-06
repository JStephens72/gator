package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"

	"github.com/JStephens72/gator/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	client := http.DefaultClient
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("User-Agent", "gator")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error status response from server: %d - %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var feed RSSFeed

	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("XML decode error: %w", err)
	}

	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
	}

	return &feed, nil

}

func scrapeFeeds(s *state) error {
	feedInfo, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("error retrieving feed: %w", err)
	}

	if err := s.db.MarkFeedFetched(context.Background(), feedInfo.ID); err != nil {
		return fmt.Errorf("error marking feed '%s' as updated: %w", feedInfo.Name, err)
	}

	feed, err := fetchFeed(context.Background(), feedInfo.Url)
	if err != nil {
		return fmt.Errorf("error downloading rss feed '%s': %w", feedInfo.Url, err)
	}

	for _, article := range feed.Channel.Item {
		t, err := parsePubDate(article.PubDate)
		var pubTime sql.NullTime
		if err == nil {
			pubTime = sql.NullTime{Time: t, Valid: true}
		} else {
			pubTime = sql.NullTime{Valid: false}
		}

		params := database.AddPostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       stringToNs(article.Title),
			Url:         article.Link,
			Description: stringToNs(article.Description),
			PublishedAt: pubTime,
			FeedID:      feedInfo.ID,
		}

		if _, err := s.db.AddPost(context.Background(), params); err != nil {
			if isUniqueViolation(err) { //duplicate URL
				continue
			}
			return fmt.Errorf("error inserting post into database: %w", err)
		}
	}
	return nil
}

func nsToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func stringToNs(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func parsePubDate(s string) (time.Time, error) {
	layouts := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.RFC3339,
	}

	var lastErr error
	for _, layout := range layouts {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t, nil
		}
		lastErr = err
	}
	return time.Time{}, lastErr
}

func isUniqueViolation(err error) bool {
	if pqErr, ok := err.(*pq.Error); ok {
		if pqErr.Code == "23505" {
			return true
		}
		return false
	}
	return false
}
