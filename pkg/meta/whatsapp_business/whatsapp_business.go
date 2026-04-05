package whatsapp_business

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"strings"
	whatsapp_business_component "wa_chat_service/pkg/meta/whatsapp_business/component"
)

func New(userAccessToken string, wabaID string, phoneNumberID string) *Client {
	return &Client{
		BaseURL:         "https://graph.facebook.com",
		Version:         os.Getenv("META_GRAPH_API_VERSION"),
		UserAccessToken: userAccessToken,
		WabaID:          wabaID,
		PhoneNumberID:   phoneNumberID,
	}
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
	client := &http.Client{}
	resp, err := client.Do(req)
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

func (wb *Client) SendMessage(phoneNumberID, to string, payload whatsapp_business_component.MessageComponent) (MessageResponse, error) {
	payloadData := map[string]any{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                to,
		"type":              payload.GetType(),
	}
	maps.Copy(payloadData, payload.GetPayload())
	endpoint := fmt.Sprintf("%s/%s/messages", wb.GetBaseURLVersion(), phoneNumberID)
	body, httpCode, err := wb.accessAPI(endpoint, "POST", payloadData)
	if err != nil {
		return MessageResponse{}, err
	} else if httpCode != http.StatusOK {
		var responseError WhatsAppBusinessError
		if err := json.Unmarshal(body, &responseError); err != nil {
			return MessageResponse{}, fmt.Errorf("unexpected http code: %d", httpCode)
		}
		return MessageResponse{}, responseError
	}
	var response MessageResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return MessageResponse{}, err
	}
	return response, nil
}

func (wb *Client) GetTemplateList() ([]any, error) {
	endpoint := fmt.Sprintf("%s/%s/message_templates", wb.GetBaseURLVersion(), wb.WabaID)
	body, httpCode, err := wb.accessAPI(endpoint, "GET", nil)
	if err != nil {
		return nil, err
	}
	if httpCode != http.StatusOK {
		var responseError WhatsAppBusinessError
		if err := json.Unmarshal(body, &responseError); err != nil {
			return nil, fmt.Errorf("unexpected http code: %d", httpCode)
		}
		return nil, responseError
	}
	var response struct {
		Data []any `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	return response.Data, nil
}
