package whatsapp_business

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (wb *Client) GetTemplateList() ([]TemplateResponse, int, error) {
	endpoint := fmt.Sprintf("%s/%s/message_templates", wb.GetBaseURLVersion(), wb.WabaID)
	body, httpCode, err := wb.accessAPI(endpoint, "GET", nil)
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
