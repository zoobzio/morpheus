package postmark

// EmailRequest is the request body for the Postmark /email endpoint.
type EmailRequest struct {
	From     string  `json:"From"`
	To       string  `json:"To"`
	Subject  string  `json:"Subject"`
	TextBody string  `json:"TextBody"`
	HtmlBody *string `json:"HtmlBody,omitempty"`
	Tag      *string `json:"Tag,omitempty"`
}

// EmailResponse is the response body from the Postmark /email endpoint.
type EmailResponse struct {
	MessageID string `json:"MessageID"`
	ErrorCode int    `json:"ErrorCode"`
	Message   string `json:"Message"`
}
