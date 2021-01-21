package config

import (
	"regexp"

	"github.com/go-msvc/errors"
)

const idPattern = `[a-z]([a-z0-9_]*[a-z0-9])*`

var idRegex = regexp.MustCompile(idPattern)

type Step struct {
	Choice  *Choice  `json:"choice"`
	Prompt  *Prompt  `json:"prompt"`
	Message *Message `json:"message"`
}

func (s *Step) Validate() error {
	log.Debugf("Validating: (%T) %+v", s, *s)
	count := 0
	if s.Choice != nil {
		if err := s.Choice.Validate(); err != nil {
			return errors.Wrapf(err, "invalid choice")
		}
		count++
	}
	if s.Prompt != nil {
		if err := s.Prompt.Validate(); err != nil {
			return errors.Wrapf(err, "invalid prompt")
		}
		count++
	}
	if s.Message != nil {
		if err := s.Message.Validate(); err != nil {
			return errors.Wrapf(err, "invalid message")
		}
		count++
	}
	if count != 1 {
		return errors.Errorf("step requires exactly one of choice|prompt|message: %+v", s)
	}
	return nil
}

func (s Step) Step() IStep {
	if s.Choice != nil {
		return s.Choice
	}
	if s.Prompt != nil {
		return s.Prompt
	}
	if s.Message != nil {
		return s.Message
	}
	return nil
}
