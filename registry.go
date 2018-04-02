package lile

// Registery is the interface to implement for external registery providers
type Registery interface {
	// Register a service
	Register(s *Service) error
	// Deregister a service
	DeRegister(s *Service) error
}
