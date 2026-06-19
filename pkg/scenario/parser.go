// Package scenario loads and validates scenario definitions.
package scenario

import (
	"fmt"
	"os"
	"time"

	"github.com/madstone-tech/ogou/pkg/engine"
	"gopkg.in/yaml.v3"
)

// Load reads a scenario from a YAML file.
func Load(path string) (*engine.Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return ParseBytes(data)
}

// ParseBytes unmarshals a scenario from raw YAML.
func ParseBytes(data []byte) (*engine.Scenario, error) {
	var s engine.Scenario
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("unmarshal yaml: %w", err)
	}
	if err := validate(&s); err != nil {
		return nil, err
	}
	return &s, nil
}

func validate(s *engine.Scenario) error {
	if s.Name == "" {
		return fmt.Errorf("scenario name is required")
	}
	if s.BaseURL == "" {
		return fmt.Errorf("base_url is required")
	}
	if len(s.Phases) == 0 {
		return fmt.Errorf("at least one phase is required")
	}
	for i, p := range s.Phases {
		if p.Duration == 0 {
			// Default to 30s if missing
			s.Phases[i].Duration = 30 * time.Second
		}
		if p.Rate.Constant == nil && p.Rate.Ramp == nil {
			return fmt.Errorf("phase %d (%s): rate.constant or rate.ramp is required", i, p.Name)
		}
		if len(p.Steps) == 0 {
			return fmt.Errorf("phase %d (%s): at least one step is required", i, p.Name)
		}
		for j, st := range p.Steps {
			if st.Name == "" {
				return fmt.Errorf("phase %d step %d: name is required", i, j)
			}
			if st.Path == "" {
				return fmt.Errorf("phase %d step %d (%s): path is required", i, j, st.Name)
			}
			if st.Method == "" {
				s.Phases[i].Steps[j].Method = "GET"
			}
		}
	}
	return nil
}
