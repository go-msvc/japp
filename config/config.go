package config

import (
	"github.com/go-msvc/errors"
	"github.com/go-msvc/logger"
)

var log = logger.Top().WithStream(logger.Terminal(logger.LogLevelDebug))

type Config struct {
	Start string          `json:"start"`
	Steps map[string]Step `json:"steps"`
}

func (c *Config) Validate() error {
	if c.Start == "" {
		c.Start = "start"
	}
	if _, ok := c.Steps[c.Start]; !ok {
		return errors.Errorf("start step missing, e.g. steps:{\"%s\":{...}}", c.Start)
	}
	for id, s := range c.Steps {
		if !idRegex.MatchString(id) {
			return errors.Errorf("invalid id for steps:{\"%s\":{...}}", id)
		}
		if err := s.Validate(); err != nil {
			return errors.Wrapf(err, "invalid steps:{\"%s\":{...}}", id)
		}
	}
	return nil
}

func (c Config) Step(id string) (Step, bool) {
	step, ok := c.Steps[id]
	if ok {
		return step, true
	}
	return Step{}, false
}
