package config

import (
	"github.com/go-msvc/errors"
	"github.com/go-msvc/japp/msg"
	"github.com/go-msvc/jsessions"
)

type Message struct {
	Text       Text   `json:"text"`
	NextStepID string `json:"next"`
}

func (c Message) Validate() error {
	return nil
}

func (c Message) Content(stepID string, data map[string]interface{}) (msg.Content, error) {
	text, err := c.Text.Render(data)
	if err != nil {
		return msg.Content{}, errors.Wrapf(err, "failed to render prompt")
	}
	return msg.Content{
		StepID: stepID,
		Message: &msg.Message{
			Text: text,
		},
	}, nil
}

func (c Message) Exec(session jsessions.ISession, values map[string]interface{}) (nextStepID string, err error) {
	//do nothing just proceed to next
	return c.NextStepID, nil
}
