package whatsapp_business

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
)

func NewWhatsappBusiness(version, userAccessToken string) *Client {
	return &Client{
		Version:         version,
		UserAccessToken: userAccessToken,
	}
}

func (wb *Client) GetMessageEndpoint(phoneNumberID string) string {
	return fmt.Sprintf("https://graph.facebook.com/%s/%s/messages", wb.Version, phoneNumberID)
}

func (wb *Client) postAPI(endpoint string, payload any) ([]byte, int, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, err
	}
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
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

func (wb *Client) SendMessage(phoneNumberID, to string, payload MessageComponent) (MessageResponse, error) {
	payloadData := map[string]any{
		"messaging_product": "whatsapp",
		"recipient_type":    "individual",
		"to":                to,
	}
	maps.Copy(payloadData, payload.GetPayload())
	body, httpCode, err := wb.postAPI(wb.GetMessageEndpoint(phoneNumberID), payloadData)
	if err != nil {
		return MessageResponse{}, err
	} else if httpCode != http.StatusOK {
		var responseError WhatsAppBusinessError
		if err := json.Unmarshal(body, &responseError); err != nil {
			return MessageResponse{}, fmt.Errorf("unexpected http code: %d", httpCode)
		}
		return MessageResponse{}, fmt.Errorf("unexpected http code: %d, status code: %d, response: %s", httpCode, responseError.Error.Code, responseError.Error.Message)
	}
	var response MessageResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return MessageResponse{}, err
	}
	return response, nil
}
