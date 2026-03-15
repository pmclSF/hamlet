package handlers

type HealthStatus struct {
	Status   string
	Version  string
	Database bool
}

func CheckHealth(version string) HealthStatus {
	return HealthStatus{
		Status:   "healthy",
		Version:  version,
		Database: true,
	}
}
