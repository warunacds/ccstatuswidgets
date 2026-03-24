package widgets

import (
	"fmt"
	"strings"

	"github.com/warunacds/ccstatuswidgets/internal/httpclient"
	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// WeatherWidget displays current weather via wttr.in.
type WeatherWidget struct{}

func (w *WeatherWidget) Name() string {
	return "weather"
}

func (w *WeatherWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	baseURL := "https://wttr.in"
	if v, ok := cfg["base_url"].(string); ok && v != "" {
		baseURL = v
	}

	city := ""
	if v, ok := cfg["city"].(string); ok {
		city = v
	}

	unitParam := "m" // metric by default
	if v, ok := cfg["units"].(string); ok && v == "imperial" {
		unitParam = "u"
	}

	url := baseURL
	if city != "" {
		url = fmt.Sprintf("%s/%s", baseURL, city)
	}
	url = fmt.Sprintf("%s?format=%%c+%%t&%s", url, unitParam)

	client := httpclient.New()
	body, err := client.Get(url)
	if err != nil {
		return nil, nil
	}

	text := strings.TrimSpace(string(body))

	return &protocol.WidgetOutput{
		Text:  text,
		Color: "yellow",
	}, nil
}
