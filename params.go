package main

import (
	"fmt"
)

// Params is used to parameterize the purge
type Params struct {
	Hosts []string `json:"hosts,omitempty"`
}

// ValidateRequiredProperties checks whether all needed properties are set
func (p *Params) ValidateRequiredProperties() (bool, []error) {

	errors := []error{}

	if len(p.Hosts) == 0 {
		errors = append(errors, fmt.Errorf("At least one host is required; set on or more hosts via the hosts array property on this stage"))
	}

	return len(errors) == 0, errors
}
