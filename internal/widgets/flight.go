package widgets

import (
	"encoding/json"
	"fmt"

	"github.com/warunacds/ccstatuswidgets/internal/httpclient"
	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

// FlightWidget displays real-time flight status using the AviationStack API.
type FlightWidget struct{}

type aviationStackResponse struct {
	Data []aviationFlight `json:"data"`
}

type aviationFlight struct {
	FlightStatus string `json:"flight_status"`
}

func (w *FlightWidget) Name() string {
	return "flight"
}

func (w *FlightWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	apiKey, _ := cfg["api_key"].(string)
	flight, _ := cfg["flight"].(string)
	if apiKey == "" || flight == "" {
		return nil, nil
	}

	baseURL := "http://api.aviationstack.com"
	if override, ok := cfg["base_url"].(string); ok && override != "" {
		baseURL = override
	}

	url := fmt.Sprintf("%s/v1/flights?access_key=%s&flight_iata=%s", baseURL, apiKey, flight)

	client := httpclient.New()
	body, err := client.Get(url)
	if err != nil {
		return nil, nil
	}

	var resp aviationStackResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, nil
	}

	if len(resp.Data) == 0 {
		return nil, nil
	}

	status := resp.Data[0].FlightStatus

	var text string
	if status == "active" {
		text = fmt.Sprintf("✈ %s ⬆ %s", flight, status)
	} else {
		text = fmt.Sprintf("✈ %s %s", flight, status)
	}

	return &protocol.WidgetOutput{
		Text:  text,
		Color: "cyan",
	}, nil
}
