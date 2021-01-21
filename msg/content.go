package msg

type Content struct {
	StepID  string   `json:"step_id" doc:"Must be specified in continue request to ensure steps are not confused"`
	Message *Message `json:"message,omitempty"`
	Prompt  *Prompt  `json:"prompt,omitempty"`
	Choice  *Choice  `json:"choice,omitempty"`
	Final   bool     `json:"final" doc:"true if no input expected and application ended"`
}

type Message struct {
	Text string `json:"text"`
}

type Prompt struct {
	Text string `json:"text"`
	//todo: indicate type of value, validation rules, regex, etc...
}

type Choice struct {
	Header  string         `json:"header"`
	Options []ChoiceOption `json:"options"`
}

type ChoiceOption struct {
	ID   string `json:"id"` //id is value to send to select this option
	Text string `json:"text"`
}
