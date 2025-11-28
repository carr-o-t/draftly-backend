package health

type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func TransformHealthCheckResponse(message string) HealthResponse {
	return HealthResponse{
		Status:  "ok",
		Message: message,
	}
}
