package whatsapp_business

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
)

type Client struct {
	BaseURL         string
	WabaID          string
	PhoneNumberID   string
	Version         string
	UserAccessToken string
	HTTPClient      *http.Client
}

func New(userAccessToken string, wabaID string, phoneNumberID string) *Client {
	return &Client{
		BaseURL:         "https://graph.facebook.com",
		Version:         os.Getenv("META_GRAPH_API_VERSION"),
		UserAccessToken: userAccessToken,
		WabaID:          wabaID,
		PhoneNumberID:   phoneNumberID,
		HTTPClient:      &http.Client{},
	}
}

func (wb *Client) httpClient() *http.Client {
	if wb.HTTPClient != nil {
		return wb.HTTPClient
	}
	return &http.Client{}
}

func (wb *Client) accessAPI(endpoint string, method string, payload any) ([]byte, int, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+wb.UserAccessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := wb.httpClient().Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	return body, resp.StatusCode, nil
}

func (wb *Client) GetBaseURLVersion() string {
	var urlBuilder strings.Builder
	urlBuilder.WriteString(wb.BaseURL)
	if wb.Version != "" {
		urlBuilder.WriteString("/" + wb.Version)
	}
	return urlBuilder.String()
}
