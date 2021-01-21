package config

import (
	"github.com/go-msvc/japp/msg"
	"github.com/go-msvc/jsessions"
)

type IStep interface {
	Exec(session jsessions.ISession, request map[string]interface{}) (nextStepID string, err error)
}

type IUserStep interface {
	IStep
	Content(stepID string, data map[string]interface{}) (msg.Content, error)
}
