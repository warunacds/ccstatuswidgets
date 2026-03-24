package widgets

import (
	"encoding/json"
	"fmt"

	"github.com/warunacds/ccstatuswidgets/internal/httpclient"
	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

const (
	hnDefaultBaseURL = "https://hacker-news.firebaseio.com"
	hnMaxTitleLen    = 40
)

// HackernewsWidget displays the current top story from Hacker News.
type HackernewsWidget struct{}

type hnItem struct {
	Title string `json:"title"`
	Score int    `json:"score"`
	URL   string `json:"url"`
}

func (w *HackernewsWidget) Name() string {
	return "hackernews"
}

func (w *HackernewsWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	baseURL := hnDefaultBaseURL
	if v, ok := cfg["base_url"].(string); ok && v != "" {
		baseURL = v
	}

	showScore := false
	if v, ok := cfg["show_score"].(bool); ok {
		showScore = v
	}

	item, err := fetchTopStory(baseURL)
	if err != nil {
		return nil, nil
	}

	title := truncateTitle(item.Title, hnMaxTitleLen)

	var text string
	if showScore {
		text = fmt.Sprintf("HN: %s (%dpts)", title, item.Score)
	} else {
		text = fmt.Sprintf("HN: %s", title)
	}

	return &protocol.WidgetOutput{
		Text:  text,
		Color: "yellow",
	}, nil
}

// fetchTopStory fetches the #1 top story from the HN API.
func fetchTopStory(baseURL string) (*hnItem, error) {
	client := httpclient.New()

	// Fetch top story IDs.
	body, err := client.Get(baseURL + "/v0/topstories.json")
	if err != nil {
		return nil, err
	}

	var ids []int
	if err := json.Unmarshal(body, &ids); err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("no top stories")
	}

	// Fetch the first item.
	itemBody, err := client.Get(fmt.Sprintf("%s/v0/item/%d.json", baseURL, ids[0]))
	if err != nil {
		return nil, err
	}

	var item hnItem
	if err := json.Unmarshal(itemBody, &item); err != nil {
		return nil, err
	}

	return &item, nil
}

// truncateTitle trims a title to maxLen characters, appending "..." if truncated.
func truncateTitle(title string, maxLen int) string {
	if len(title) <= maxLen {
		return title
	}
	return title[:maxLen] + "..."
}
