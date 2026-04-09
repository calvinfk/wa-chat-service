package whatsapp_business

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"wa_chat_service/pkg/formatter"
)

type Client struct {
	AppID           string
	BaseURL         string
	WabaID          string
	PhoneNumberID   string
	Version         string
	UserAccessToken string

	httpClient *http.Client
	validator  *formatter.StructValidator
}

func New(userAccessToken string, wabaID string, phoneNumberID string) *Client {
	validator := formatter.Validator()
	appID := os.Getenv("META_APP_ID")
	if appID == "" {
		log.Printf("[WARNING][pkg/meta/whatsapp_business/client.go][New] META_APP_ID is not set in environment variables")
	}
	return &Client{
		AppID:           appID,
		BaseURL:         "https://graph.facebook.com",
		Version:         os.Getenv("META_GRAPH_API_VERSION"),
		UserAccessToken: userAccessToken,
		WabaID:          wabaID,
		PhoneNumberID:   phoneNumberID,
		httpClient:      &http.Client{},
		validator:       validator,
	}
}

func (wb *Client) accessAPI(method string, endpoint string, payload any) ([]byte, int, error) {
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
	resp, err := wb.httpClient.Do(req)
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

func (wb *Client) accessAPIWithoutAuth(method string, endpoint string, payload any) ([]byte, int, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequest(method, endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := wb.httpClient.Do(req)
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
