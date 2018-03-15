package tracker

// HealthChecker wraps CheckHealth method
type HealthChecker interface {
	CheckHealth() error
}

