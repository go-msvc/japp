package config

import (
	"fmt"

	"github.com/go-msvc/errors"
	"github.com/go-msvc/japp/msg"
	"github.com/go-msvc/jsessions"
)

type Choice struct {
	Header  Text     `json:"header"`
	Options []Option `json:"options"`
}

func (c *Choice) Validate() error {
	log.Debugf("Validating: (%T) %+v", c, *c)
	if err := c.Header.Validate(); err != nil {
		return errors.Wrapf(err, "invalid header")
	}
	if len(c.Options) == 0 {
		return errors.Errorf("missing options")
	}
	next := 1
	for i, o := range c.Options {
		if err := o.Validate(); err != nil {
			return errors.Wrapf(err, "invalid options[%d]", i)
		}
		if o.ID == "" {
			nextStrID := fmt.Sprintf("%d", next)
			for c.HasID(nextStrID) {
				next++
				nextStrID = fmt.Sprintf("%d", next)
			}
			o.ID = nextStrID
			next++
		}
		c.Options[i] = o
		log.Debugf("validated option: %+v", o)
	}
	return nil
}

func (c Choice) HasID(id string) bool {
	for _, o := range c.Options {
		if o.ID == id {
			return true
		}
	}
	return false
}

func (c Choice) Content(stepID string, data map[string]interface{}) (msg.Content, error) {
	choice := msg.Choice{
		Options: []msg.ChoiceOption{},
	}
	var err error

	choice.Header, err = c.Header.Render(data)
	if err != nil {
		return msg.Content{}, errors.Wrapf(err, "failed to render prompt")
	}

	for idx, o := range c.Options {
		optionText, err := o.Text.Render(data)
		if err != nil {
			return msg.Content{}, errors.Wrapf(err, "failed to render choice[%d]", idx)
		}
		choice.Options = append(choice.Options, msg.ChoiceOption{
			Text: optionText,
			ID:   o.ID,
		})
	}
	return msg.Content{StepID: stepID, Choice: &choice}, nil
}

func (c Choice) Prompt() Text {
	s := Text("")
	if len(c.Header) > 0 {
		s += c.Header
	}
	for _, o := range c.Options {
		s += "\n" + Text(o.ID+") ") + o.Text
	}
	return Text(s)
}

func (c Choice) Exec(session jsessions.ISession, values map[string]interface{}) (nextStepID string, err error) {
	userInput, ok := values["user"].(string)
	if !ok {
		return "", errors.Errorf("no 'user' input")
	}
	for _, o := range c.Options {
		if o.ID == userInput {
			return o.NextStepID, nil
		}
	}
	return "", errors.Errorf("invalid selection")
}

type Option struct {
	ID         string `json:"id"`
	Text       Text   `json:"text"`
	NextStepID string `json:"next"`
}

func (c *Option) Validate() error {
	if err := c.Text.Validate(); err != nil {
		return errors.Wrapf(err, "invalid text")
	}
	if c.NextStepID == "" {
		return errors.Errorf("missing next")
	}
	return nil
}
