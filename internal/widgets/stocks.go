package widgets

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/warunacds/ccstatuswidgets/internal/httpclient"
	"github.com/warunacds/ccstatuswidgets/internal/protocol"
)

const defaultStocksBaseURL = "https://query1.finance.yahoo.com"

// StocksWidget displays daily price changes for configured stock symbols.
type StocksWidget struct{}

func (w *StocksWidget) Name() string {
	return "stocks"
}

// yahooChartResponse models the subset of Yahoo Finance chart API we need.
type yahooChartResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				RegularMarketPrice float64 `json:"regularMarketPrice"`
				PreviousClose      float64 `json:"previousClose"`
			} `json:"meta"`
		} `json:"result"`
	} `json:"chart"`
}

func (w *StocksWidget) Render(input *protocol.StatusLineInput, cfg map[string]interface{}) (*protocol.WidgetOutput, error) {
	if cfg == nil {
		return nil, nil
	}

	symbolsRaw, ok := cfg["symbols"]
	if !ok {
		return nil, nil
	}

	symbolsList, ok := symbolsRaw.([]interface{})
	if !ok || len(symbolsList) == 0 {
		return nil, nil
	}

	baseURL := defaultStocksBaseURL
	if bu, ok := cfg["base_url"].(string); ok && bu != "" {
		baseURL = bu
	}

	client := httpclient.New()
	var parts []string

	for _, sym := range symbolsList {
		symbol, ok := sym.(string)
		if !ok || symbol == "" {
			continue
		}

		url := fmt.Sprintf("%s/v8/finance/chart/%s?range=1d&interval=1d", baseURL, symbol)
		body, err := client.Get(url)
		if err != nil {
			continue
		}

		var resp yahooChartResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			continue
		}

		if len(resp.Chart.Result) == 0 {
			continue
		}

		meta := resp.Chart.Result[0].Meta
		if meta.PreviousClose == 0 {
			continue
		}

		change := ((meta.RegularMarketPrice - meta.PreviousClose) / meta.PreviousClose) * 100

		var color, sign string
		if change >= 0 {
			color = "\033[0;32m"
			sign = "+"
		} else {
			color = "\033[0;31m"
			sign = ""
		}

		parts = append(parts, fmt.Sprintf("%s%s %s%.1f%%\033[0m", color, symbol, sign, change))
	}

	if len(parts) == 0 {
		return nil, nil
	}

	return &protocol.WidgetOutput{
		Text:  strings.Join(parts, " "),
		Color: "",
	}, nil
}
