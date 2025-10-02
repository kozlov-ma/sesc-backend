package respond

type HealthCheck struct {
	Status string `json:"status"`
}

func NewHealthCheckOK() HealthCheck {
	return HealthCheck{
		Status: "ok",
	}
}
