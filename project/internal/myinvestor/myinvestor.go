package myinvestor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type MyInvestor struct {
	client *http.Client
	url    string
}

func NewMyInvestor() *MyInvestor {
	return &MyInvestor{
		client: &http.Client{},
		url:    "https://api.myinvestor.es/myinvestor-server/rest/public/funds/top",
	}
}

type BestSeller struct {
	ISIN           string   `json:"isin"`
	Name           string   `json:"name"`
	Manager        string   `json:"entity-management"`
	Currency       string   `json:"currency"`
	TER            float64  `json:"ter"`
	Distributing   bool     `json:"fondo-distribucion"`
	ESG            bool     `json:"socialmente-responsable"`
	NoRetrocession bool     `json:"no-retrocession"`
	Return1Y       *float64 `json:"1-year-return"`
	Return3Y       *float64 `json:"3-year-return"`
	Return5Y       *float64 `json:"5-year-return"`
	ReturnsDate    *string  `json:"returns-update-date"`
}

func (my *MyInvestor) GetBestSellers(ctx context.Context) ([]BestSeller, error) {
	f, err := os.Open("best-sellers.json")
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	var res []BestSeller
	if err := json.NewDecoder(f).Decode(&res); err != nil {
		return nil, fmt.Errorf("json decode: %w", err)
	}
	return res, nil
}

func (my *MyInvestor) RefreshBestSellers(ctx context.Context) ([]BestSeller, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", my.url+"/BEST_SELLERS", nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	res, err := my.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("wrong status: %d, body: %s", res.StatusCode, body)
	}

	var body []BestSeller
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode body: %w", err)
	}
	return body, nil
}
