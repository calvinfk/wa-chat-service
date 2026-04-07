package whatsapp_business

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"mime/multipart"
	"net/http"
	"net/textproto"
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
		return parseMetaErrorResponse(MessageResponse{}, body, httpCode)
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
		return parseMetaErrorResponse([]any{}, body, httpCode)
	}
	var response struct {
		Data []any `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, httpCode, err
	}
	return response.Data, httpCode, nil
}

func (wb *Client) UploadMedia(fileBytes []byte, filename string, mimeType string) (UploadMediaResponse, int, error) {
	var emptyResponse UploadMediaResponse
	// Implementation for uploading media
	endpoint := fmt.Sprintf("%s/%s/media", wb.GetBaseURLVersion(), wb.PhoneNumberID)
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	// Add messaging_product field
	_ = w.WriteField("messaging_product", "whatsapp")

	// Add the file part with the correct Content-Type header
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename))
	h.Set("Content-Type", mimeType)

	part, err := w.CreatePart(h)
	if err != nil {
		return emptyResponse, 0, fmt.Errorf("create part: %w", err)
	}
	if _, err = part.Write(fileBytes); err != nil {
		return emptyResponse, 0, fmt.Errorf("write bytes: %w", err)
	}
	if err := w.Close(); err != nil {
		return emptyResponse, 0, fmt.Errorf("close writer: %w", err)
	}
	req, err := http.NewRequest("POST", endpoint, &buf)
	if err != nil {
		return emptyResponse, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+wb.UserAccessToken)
	req.Header.Set("Content-Type", w.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return emptyResponse, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return emptyResponse, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return parseMetaErrorResponse(emptyResponse, body, resp.StatusCode)
	}
	var response UploadMediaResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return emptyResponse, resp.StatusCode, err
	}
	return response, resp.StatusCode, nil
}

func (wb *Client) GetMediaURL(mediaID string) (GetMediaURLResponse, int, error) {
	endpoint := fmt.Sprintf("%s/%s", wb.GetBaseURLVersion(), mediaID)
	body, httpCode, err := wb.accessAPI(endpoint, "GET", nil)
	if err != nil {
		return GetMediaURLResponse{}, httpCode, err
	}
	if httpCode != http.StatusOK {
		return parseMetaErrorResponse(GetMediaURLResponse{}, body, httpCode)
	}
	var response GetMediaURLResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return GetMediaURLResponse{}, httpCode, err
	}
	return response, httpCode, nil
}

func (wb *Client) DownloadMedia(mediaURL string) ([]byte, http.Header, int, error) {
	req, err := http.NewRequest("GET", mediaURL, nil)
	if err != nil {
		return nil, nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+wb.UserAccessToken)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, resp.Header, resp.StatusCode, fmt.Errorf("failed to download media, status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.Header, resp.StatusCode, err
	}
	return body, resp.Header, resp.StatusCode, nil
}

func (wb *Client) DeleteMedia(mediaID string) (DeleteMediaResponse, int, error) {
	endpoint := fmt.Sprintf("%s/%s", wb.GetBaseURLVersion(), mediaID)
	body, httpCode, err := wb.accessAPI(endpoint, "DELETE", nil)
	if err != nil {
		return DeleteMediaResponse{}, httpCode, err
	}
	if httpCode != http.StatusOK {
		return parseMetaErrorResponse(DeleteMediaResponse{}, body, httpCode)
	}
	var response DeleteMediaResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return DeleteMediaResponse{}, httpCode, err
	}
	return response, httpCode, nil
}

func (wb *Client) GetHeaders(url string) (http.Header, error) {
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp.Header, nil
}

func ParseMediaExtension(mimeType string) string {
	extension, exists := mimeTypeExtensionMap[mimeType]
	if !exists {
		return ""
	}
	return extension
}

func parseMetaErrorResponse[T any](emptyResponse T, body []byte, httpCode int) (T, int, error) {
	var responseError WhatsAppBusinessError
	if err := json.Unmarshal(body, &responseError); err != nil {
		return emptyResponse, httpCode, fmt.Errorf("unexpected http code: %d", httpCode)
	}
	return emptyResponse, httpCode, responseError
}
