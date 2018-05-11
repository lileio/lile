package lile

// Registry is the interface to implement for external registry providers
type Registry interface {
	// Register a service
	Register(s *Service) error
	// Deregister a service
	DeRegister(s *Service) error
	// Get a service by name
	Get(name string) (string, error)
}
