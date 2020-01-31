package request_domains

type GetOldMessagesRequest struct {
	From int64 `json:"from"`
	To   int64 `json:"to"`
}
