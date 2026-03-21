package common

// ApiResponse is the standard API response envelope.
// All external API responses use this structure:
//
//	{"status": "...", "reason": "...", "next_action_hint": "..."}
type ApiResponse struct {
	Status         string `json:"status"`
	Reason         string `json:"reason"`
	NextActionHint string `json:"next_action_hint"`
}

// NewApiResponse creates an ApiResponse with all fields set.
func NewApiResponse(status, reason, nextActionHint string) ApiResponse {
	return ApiResponse{
		Status:         status,
		Reason:         reason,
		NextActionHint: nextActionHint,
	}
}

// OkResponse creates a successful ApiResponse with status "ok".
func OkResponse(reason, nextActionHint string) ApiResponse {
	return NewApiResponse("ok", reason, nextActionHint)
}

// ErrorResponse creates an error ApiResponse with status "error".
func ErrorResponse(reason, nextActionHint string) ApiResponse {
	return NewApiResponse("error", reason, nextActionHint)
}
