package services

type HealthService struct{}

func NewHealthService() *HealthService {
	return &HealthService{}
}

func (s *HealthService) Check() string {
	return "Server is running"
}
