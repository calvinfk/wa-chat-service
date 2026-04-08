package whatsapp_business

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	endpointTemplate = "message_templates"
)

func (wb *Client) GetTemplateList(query ...string) ([]TemplateResponse, int, error) {
	endpoint := fmt.Sprintf("%s/%s/%s", wb.GetBaseURLVersion(), wb.WabaID, endpointTemplate)
	if len(query) > 0 {
		endpoint += "?" + strings.Join(query, "&")
	}
	body, httpCode, err := wb.accessAPI(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, httpCode, err
	}
	if httpCode != http.StatusOK {
		return parseMetaErrorResponse([]TemplateResponse{}, body, httpCode)
	}
	var response struct {
		Data []TemplateResponse `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, httpCode, err
	}
	return response.Data, httpCode, nil
}

func (wb *Client) GetTemplateByID(templateID string, query ...string) (TemplateResponse, int, error) {
	endpoint := fmt.Sprintf("%s/%s/%s", wb.GetBaseURLVersion(), wb.WabaID, templateID)
	if len(query) > 0 {
		endpoint += "?" + strings.Join(query, "&")
	}
	body, httpCode, err := wb.accessAPI(http.MethodGet, endpoint, nil)
	if err != nil {
		return TemplateResponse{}, httpCode, err
	}
	if httpCode != http.StatusOK {
		return parseMetaErrorResponse(TemplateResponse{}, body, httpCode)
	}
	var response TemplateResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return TemplateResponse{}, httpCode, err
	}
	return response, httpCode, nil
}

func (wb *Client) CreateTemplate(payload TemplateCreateRequest) (TemplateCreateResponse, int, error) {
	endpoint := fmt.Sprintf("%s/%s/%s", wb.GetBaseURLVersion(), wb.WabaID, endpointTemplate)
	// TODO: validate payload before sending request
	body, httpCode, err := wb.accessAPI(http.MethodPost, endpoint, payload)
	if err != nil {
		return TemplateCreateResponse{}, httpCode, err
	}
	if httpCode != http.StatusOK && httpCode != http.StatusCreated {
		return parseMetaErrorResponse(TemplateCreateResponse{}, body, httpCode)
	}
	var response TemplateCreateResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return TemplateCreateResponse{}, httpCode, err
	}
	return response, httpCode, nil
}

func (wb *Client) DeleteTemplate(templateID string, templateName string) (TemplateDeleteResponse, int, error) {
	if templateID == "" || templateName == "" {
		return TemplateDeleteResponse{}, 0, fmt.Errorf("templateID and templateName are required")
	}
	endpoint := fmt.Sprintf("%s/%s/%s?hsm_id=%s&name=%s", wb.GetBaseURLVersion(), wb.WabaID, endpointTemplate, templateID, templateName)
	req, err := http.NewRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return TemplateDeleteResponse{}, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+wb.UserAccessToken)
	resp, err := wb.httpClient.Do(req)
	if err != nil {
		return TemplateDeleteResponse{}, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TemplateDeleteResponse{}, resp.StatusCode, err
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return parseMetaErrorResponse(TemplateDeleteResponse{}, body, resp.StatusCode)
	}
	var response TemplateDeleteResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return TemplateDeleteResponse{Success: true}, resp.StatusCode, nil
	}
	return response, resp.StatusCode, nil
}
