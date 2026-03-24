package widgets

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/warunacds/ccstatuswidgets/internal/httpclient"
	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// teamAbbreviations maps full team names to short codes.
var teamAbbreviations = map[string]string{
	"Sri Lanka":    "SL",
	"Australia":    "AUS",
	"India":        "IND",
	"England":      "ENG",
	"Pakistan":     "PAK",
	"South Africa": "SA",
	"New Zealand":  "NZ",
	"West Indies":  "WI",
	"Bangladesh":   "BAN",
	"Afghanistan":  "AFG",
	"Zimbabwe":     "ZIM",
	"Ireland":      "IRE",
	"Netherlands":  "NED",
	"Scotland":     "SCO",
	"Nepal":        "NEP",
	"Oman":         "OMA",
	"Namibia":      "NAM",
	"UAE":          "UAE",
	"USA":          "USA",
}

// cricketAPIResponse represents the response from the cricket API.
type cricketAPIResponse struct {
	Data []cricketMatch `json:"data"`
}

type cricketMatch struct {
	Name         string         `json:"name"`
	Status       string         `json:"status"`
	MatchType    string         `json:"matchType"`
	Teams        []string       `json:"teams"`
	Score        []cricketScore `json:"score"`
	MatchStarted bool           `json:"matchStarted"`
	MatchEnded   bool           `json:"matchEnded"`
}

type cricketScore struct {
	Runs    int     `json:"r"`
	Wickets int     `json:"w"`
	Overs   float64 `json:"o"`
	Inning  string  `json:"inning"`
}

// CricketWidget displays live cricket scores.
type CricketWidget struct{}

func (w *CricketWidget) Name() string {
	return "cricket"
}

func (w *CricketWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	apiKey, ok := cfg["api_key"].(string)
	if !ok || apiKey == "" {
		return nil, nil
	}

	baseURL := "https://api.cricapi.com/v1"
	if v, ok := cfg["base_url"].(string); ok && v != "" {
		baseURL = v
	}

	teamFilter := ""
	if v, ok := cfg["team"].(string); ok {
		teamFilter = v
	}

	url := fmt.Sprintf("%s/currentMatches?apikey=%s", baseURL, apiKey)

	client := httpclient.New()
	body, err := client.Get(url)
	if err != nil {
		return nil, nil
	}

	var resp cricketAPIResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, nil
	}

	if len(resp.Data) == 0 {
		return nil, nil
	}

	// Find the relevant match
	var match *cricketMatch
	for i := range resp.Data {
		m := &resp.Data[i]
		if teamFilter != "" {
			if !matchInvolvesTeam(m, teamFilter) {
				continue
			}
		}
		match = m
		break
	}

	if match == nil {
		return nil, nil
	}

	text := formatMatchText(match)

	return &protocol.WidgetOutput{
		Text:  text,
		Color: "green",
	}, nil
}

// matchInvolvesTeam checks if the match involves the given team abbreviation.
func matchInvolvesTeam(m *cricketMatch, teamCode string) bool {
	for _, team := range m.Teams {
		abbr := abbreviateTeam(team)
		if strings.EqualFold(abbr, teamCode) {
			return true
		}
	}
	return false
}

// abbreviateTeam converts a full team name to its short code.
func abbreviateTeam(name string) string {
	if abbr, ok := teamAbbreviations[name]; ok {
		return abbr
	}
	// Fallback: use first 3 characters uppercased
	if len(name) >= 3 {
		return strings.ToUpper(name[:3])
	}
	return strings.ToUpper(name)
}

// formatMatchText produces the display string for a cricket match.
func formatMatchText(m *cricketMatch) string {
	if m.MatchEnded {
		return formatCompletedMatch(m)
	}
	return formatLiveMatch(m)
}

// formatLiveMatch shows the current batting score.
func formatLiveMatch(m *cricketMatch) string {
	if len(m.Score) == 0 {
		// No score yet, show teams
		return fmt.Sprintf("\U0001F3CF %s v %s", abbreviateTeam(m.Teams[0]), abbreviateTeam(m.Teams[1]))
	}

	// Show the latest innings score
	latest := m.Score[len(m.Score)-1]
	// Extract the team abbreviation from the inning string
	teamAbbr := extractTeamFromInning(latest.Inning, m.Teams)

	return fmt.Sprintf("\U0001F3CF %s %d/%d (%.1f)", teamAbbr, latest.Runs, latest.Wickets, latest.Overs)
}

// formatCompletedMatch shows the result.
func formatCompletedMatch(m *cricketMatch) string {
	if len(m.Teams) >= 2 {
		return fmt.Sprintf("\U0001F3CF %s v %s - %s", abbreviateTeam(m.Teams[0]), abbreviateTeam(m.Teams[1]), m.Status)
	}
	return fmt.Sprintf("\U0001F3CF %s", m.Status)
}

// extractTeamFromInning finds which team is batting from the inning string.
func extractTeamFromInning(inning string, teams []string) string {
	for _, team := range teams {
		if strings.Contains(inning, team) {
			return abbreviateTeam(team)
		}
	}
	// Fallback: use first team
	if len(teams) > 0 {
		return abbreviateTeam(teams[0])
	}
	return "???"
}
