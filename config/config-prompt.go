package config

import (
	"github.com/go-msvc/errors"
	"github.com/go-msvc/japp/msg"
	"github.com/go-msvc/jsessions"
)

type Prompt struct {
	Text       Text   `json:"text"`
	Set        string `json:"set"`
	NextStepID string `json:"next"`
}

func (c Prompt) Validate() error {
	return nil
}

func (c Prompt) Content(stepID string, data map[string]interface{}) (msg.Content, error) {
	text, err := c.Text.Render(data)
	if err != nil {
		return msg.Content{}, errors.Wrapf(err, "failed to render prompt")
	}
	return msg.Content{
		StepID: stepID,
		Prompt: &msg.Prompt{
			Text: text,
		},
	}, nil
}

func (c Prompt) Exec(session jsessions.ISession, values map[string]interface{}) (nextStepID string, err error) {
	userInput, ok := values["user"].(string)
	if !ok {
		return "", errors.Errorf("no 'user' input")
	}
	log.Debugf("  %s=\"%s\"\n", c.Set, userInput)
	if _, err := session.SetString(c.Set, userInput); err != nil {
		return "", errors.Errorf("failed to store")
	}
	return c.NextStepID, nil
}
