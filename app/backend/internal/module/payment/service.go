package payment

import (
	"errors"
	"strings"

	"gorm.io/gorm"
	"uuid"
)

var (
	ErrNotFound   = errors.New("resource not found")
	ErrNotPayable = errors.New("order is not payable")
)

// Service owns payment sessions, callback handling and the simulator.
type Service struct {
	db     *gorm.DB
	cfg    Config
	sender Sender
}

// NewService wires the service. The sender is only required when the simulator is enabled.
func NewService(db *gorm.DB, cfg Config, sender Sender) (*Service, error) {
	if db == nil {
		return nil, errors.New("payment service requires a database")
	}
	if cfg.SimulatorEnabled && sender == nil {
		return nil, errors.New("payment simulator requires a sender")
	}
	return &Service{db: db, cfg: cfg, sender: sender}, nil
}

// SimulatorEnabled reports whether simulator control endpoints are exposed.
func (service *Service) SimulatorEnabled() bool { return service.cfg.SimulatorEnabled }

// canonicalUUID returns the lower-case canonical form, or false when the text is not a UUID.
func canonicalUUID(value string) (string, bool) {
	if len(value) != 36 {
		return "", false
	}
	parsed, err := uuid.Parse(value)
	if err != nil {
		return "", false
	}
	return strings.ToLower(parsed.String()), true
}
