package msg

import "github.com/pkg/errors"

type IValidator interface {
	Validate() error
}

type StartRequest struct {
	ClientID string                 `json:"client-id" doc:"May be empty then you must store the unique client id in the response and keep using it."`
	Data     map[string]interface{} `json:"data"`
}

func (req StartRequest) Validate() error {
	return nil
}

type StartResponse struct {
	ClientID  string  `json:"client-id" doc:"Echoed from the request, store and use for next start requests if possible."`
	SessionID string  `json:"session-id" doc:"Use this in each cont/end request"`
	Content   Content `json:"content"`
}

type ContinueRequest struct {
	SessionID string                 `json:"session-id"`
	StepID    string                 `json:"step_id" doc:"Echoed from last content that user reply to"`
	Data      map[string]interface{} `json:"data"`
}

func (req ContinueRequest) Validate() error {
	if req.SessionID == "" {
		return errors.Errorf("missiong session-id")
	}
	return nil
}

type ContinueResponse struct {
	SessionID string  `json:"session-id"`
	Content   Content `json:"content"`
}

type EndRequest struct {
	SessionID string `json:"session-id"`
}

func (req EndRequest) Validate() error {
	if req.SessionID == "" {
		return errors.Errorf("missiong session-id")
	}
	return nil
}

type EndResponse struct {
	SessionID string `json:"session-id"`
}
