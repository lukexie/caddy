package caddy2

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

var currentCfg *Config
var currentCfgMu sync.Mutex

// Start runs Caddy with the given config.
func Start(cfg Config) error {
	cfg.runners = make(map[string]Runner)

	for modName, rawMsg := range cfg.Modules {
		val, err := LoadModule(modName, rawMsg)
		if err != nil {
			return fmt.Errorf("loading module '%s': %v", modName, err)
		}
		cfg.runners[modName] = val.(Runner)
	}

	for name, r := range cfg.runners {
		err := r.Run()
		if err != nil {
			// TODO: If any one has an error, stop the others
			return fmt.Errorf("%s module: %v", name, err)
		}
	}

	currentCfgMu.Lock()
	if currentCfg != nil {
		for _, r := range currentCfg.runners {
			err := r.Cancel()
			if err != nil {
				log.Println(err)
			}
		}
	}
	currentCfg = &cfg
	currentCfgMu.Unlock()

	return nil
}

// Runner is a thing that Caddy runs.
type Runner interface {
	Run() error
	Cancel() error
}

// Config represents a Caddy configuration.
type Config struct {
	TestVal string                     `json:"testval"`
	Modules map[string]json.RawMessage `json:"modules"`
	runners map[string]Runner
}

// Duration is a JSON-string-unmarshable duration type.
type Duration time.Duration

// UnmarshalJSON satisfies json.Unmarshaler.
func (d *Duration) UnmarshalJSON(b []byte) error {
	dd, err := time.ParseDuration(strings.Trim(string(b), `"`))
	if err != nil {
		return err
	}
	cd := Duration(dd)
	d = &cd
	return nil
}

// MarshalJSON satisfies json.Marshaler.
func (d Duration) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, time.Duration(d).String())), nil
}
