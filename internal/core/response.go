package core

type Response struct {
	Error    ErrorResponse `json:"error"`
	Response interface{}   `json:"response"`
	Data     interface{}   `json:"data"`
}

type ErrorResponse struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}
