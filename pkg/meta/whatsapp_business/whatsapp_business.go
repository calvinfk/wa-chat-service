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

func (wb *Client) SendMessage(phoneNumberID, to, recipientType string, payload whatsapp_business_component.MessageComponent) (MessageResponse, int, error) {
	payloadData := map[string]any{
		"messaging_product": "whatsapp",
		"recipient_type":    recipientType,
		"to":                to,
		"type":              payload.GetType(),
	}
	maps.Copy(payloadData, payload.GetPayload())
	endpoint := fmt.Sprintf("%s/%s/messages", wb.GetBaseURLVersion(), phoneNumberID)
	body, httpCode, err := wb.accessAPI(endpoint, "POST", payloadData)
	if err != nil {
		return MessageResponse{}, httpCode, err
	} else if httpCode != http.StatusOK {
		var responseError WhatsAppBusinessError
		if err := json.Unmarshal(body, &responseError); err != nil {
			return MessageResponse{}, httpCode, fmt.Errorf("unexpected http code: %d", httpCode)
		}
		return MessageResponse{}, httpCode, responseError
	}
	var response MessageResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return MessageResponse{}, httpCode, err
	}
	return response, httpCode, nil
}

func (wb *Client) GetTemplateList() ([]any, int, error) {
	endpoint := fmt.Sprintf("%s/%s/message_templates", wb.GetBaseURLVersion(), wb.WabaID)
	body, httpCode, err := wb.accessAPI(endpoint, "GET", nil)
	if err != nil {
		return nil, httpCode, err
	}
	if httpCode != http.StatusOK {
		var responseError WhatsAppBusinessError
		if err := json.Unmarshal(body, &responseError); err != nil {
			return nil, httpCode, fmt.Errorf("unexpected http code: %d", httpCode)
		}
		return nil, httpCode, responseError
	}
	var response struct {
		Data []any `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, httpCode, err
	}
	return response.Data, httpCode, nil
}
