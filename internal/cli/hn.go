package cli

import (
	"encoding/json"
	"fmt"

	"github.com/warunacds/ccstatuswidgets/internal/httpclient"
)

const (
	hnDefaultURL = "https://hacker-news.firebaseio.com"
	hnTopN       = 5
)

type hnStory struct {
	Title string `json:"title"`
	Score int    `json:"score"`
	URL   string `json:"url"`
}

// RunHN fetches the top 5 Hacker News stories and prints them.
// If baseURL is empty, the default HN API URL is used.
func RunHN(baseURL string) error {
	if baseURL == "" {
		baseURL = hnDefaultURL
	}

	client := httpclient.New()

	// Fetch top story IDs.
	body, err := client.Get(baseURL + "/v0/topstories.json")
	if err != nil {
		return fmt.Errorf("failed to fetch top stories: %w", err)
	}

	var ids []int
	if err := json.Unmarshal(body, &ids); err != nil {
		return fmt.Errorf("failed to parse top stories: %w", err)
	}

	// Limit to top N.
	if len(ids) > hnTopN {
		ids = ids[:hnTopN]
	}

	for i, id := range ids {
		itemBody, err := client.Get(fmt.Sprintf("%s/v0/item/%d.json", baseURL, id))
		if err != nil {
			// Skip stories that fail to fetch.
			continue
		}

		var story hnStory
		if err := json.Unmarshal(itemBody, &story); err != nil {
			continue
		}

		fmt.Printf("%d. %s (%d pts)\n", i+1, story.Title, story.Score)
		fmt.Printf("   %s\n", story.URL)
	}

	return nil
}
