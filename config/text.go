package config

import (
	"bytes"
	"text/template"
)

type Text string

func (c Text) Validate() error {
	return nil
}

// type IRenderer interface {
// 	Render(data map[string]interface{}) (string, error)
// }

func (c Text) Render(data map[string]interface{}) (string, error) {
	//todo: compile template from config only once... not here!
	t := template.Must(template.New("content").Parse(string(c)))
	r := bytes.NewBuffer(nil)
	err := t.Execute(r, data)
	return string(r.Bytes()), err
}
