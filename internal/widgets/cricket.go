package widgets

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/warunacds/ccstatuswidgets/internal/httpclient"
	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// CricketWidget displays live cricket scores using ESPN's free API.
// No API key required.
type CricketWidget struct{}

func (w *CricketWidget) Name() string {
	return "cricket"
}

// ESPN scoreboard response structures.
type espnScoreboard struct {
	Events []espnEvent `json:"events"`
}

type espnEvent struct {
	Name         string           `json:"name"`
	Status       espnStatus       `json:"status"`
	Competitions []espnCompetition `json:"competitions"`
}

type espnStatus struct {
	Type espnStatusType `json:"type"`
}

type espnStatusType struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Detail      string `json:"detail"`
	State       string `json:"state"`
}

type espnCompetition struct {
	Competitors []espnCompetitor `json:"competitors"`
	Status      espnStatus       `json:"status"`
}

type espnCompetitor struct {
	Team  espnTeam `json:"team"`
	Score string   `json:"score"`
}

type espnTeam struct {
	Abbreviation string `json:"abbreviation"`
	DisplayName  string `json:"displayName"`
	ShortName    string `json:"shortDisplayName"`
}

// Default leagues to check: IPL (8048), International (8676)
var defaultLeagues = []string{"8048", "8676"}

func (w *CricketWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	baseURL := "http://site.api.espn.com/apis/site/v2/sports/cricket"
	if v, ok := cfg["base_url"].(string); ok && v != "" {
		baseURL = v
	}

	// Get leagues to check
	leagues := defaultLeagues
	if v, ok := cfg["leagues"]; ok {
		if arr, ok := v.([]interface{}); ok {
			leagues = nil
			for _, l := range arr {
				if s, ok := l.(string); ok {
					leagues = append(leagues, s)
				}
			}
		}
	}

	teamFilter := ""
	if v, ok := cfg["team"].(string); ok {
		teamFilter = strings.ToUpper(v)
	}

	client := httpclient.New()

	for _, league := range leagues {
		url := fmt.Sprintf("%s/%s/scoreboard", baseURL, league)
		body, err := client.Get(url)
		if err != nil {
			continue
		}

		var sb espnScoreboard
		if err := json.Unmarshal(body, &sb); err != nil {
			continue
		}

		for _, event := range sb.Events {
			if len(event.Competitions) == 0 {
				continue
			}

			comp := event.Competitions[0]
			if len(comp.Competitors) < 2 {
				continue
			}

			// Team filter
			if teamFilter != "" {
				found := false
				for _, c := range comp.Competitors {
					if strings.EqualFold(c.Team.Abbreviation, teamFilter) {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			text := formatESPNMatch(comp, event.Status)
			if text == "" {
				continue
			}

			return &protocol.WidgetOutput{
				Text:  text,
				Color: "green",
			}, nil
		}
	}

	return nil, nil
}

func formatESPNMatch(comp espnCompetition, status espnStatus) string {
	if len(comp.Competitors) < 2 {
		return ""
	}

	t1 := comp.Competitors[0]
	t2 := comp.Competitors[1]

	state := strings.ToLower(status.Type.State)

	switch state {
	case "in":
		// Live match — show scores
		parts := []string{"\U0001F3CF"}
		for _, c := range comp.Competitors {
			if c.Score != "" {
				parts = append(parts, fmt.Sprintf("%s %s", c.Team.Abbreviation, c.Score))
			}
		}
		if status.Type.Detail != "" {
			parts = append(parts, fmt.Sprintf("(%s)", status.Type.Detail))
		}
		return strings.Join(parts, " ")

	case "post":
		// Completed — show result
		desc := status.Type.Description
		if desc == "" {
			desc = status.Type.Detail
		}
		return fmt.Sprintf("\U0001F3CF %s v %s - %s", t1.Team.Abbreviation, t2.Team.Abbreviation, desc)

	case "pre":
		// Upcoming
		return fmt.Sprintf("\U0001F3CF %s v %s - %s", t1.Team.Abbreviation, t2.Team.Abbreviation, status.Type.Detail)

	default:
		return fmt.Sprintf("\U0001F3CF %s v %s", t1.Team.Abbreviation, t2.Team.Abbreviation)
	}
}
