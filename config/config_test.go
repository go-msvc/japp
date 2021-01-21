package config_test

import (
	"testing"

	"github.com/go-msvc/japp/config"
	"github.com/go-msvc/jsessions"
	jsessionmem "github.com/go-msvc/jsessions/mem"
)

func Test1(t *testing.T) {
	cfg := config.Config{
		Start: "main",
		Steps: map[string]config.Step{
			"main": {
				Choice: &config.Choice{
					Header: "*** MAIN MENU ***",
					Options: []config.Option{
						{ID: "1", Text: "Register", NextStepID: "reg_name"},
						{ID: "2", Text: "My name", NextStepID: "show_name"},
					},
				},
			},
			"reg_name": {
				Prompt: &config.Prompt{
					Text:       "Enter your name",
					Set:        "name",
					NextStepID: "do_reg",
				},
			},
			"do_reg": {
				Message: &config.Message{
					Text:       "You are now registered",
					NextStepID: "main",
				},
			},
			"show_name": {
				Message: &config.Message{
					Text:       "Your name is {{.name}}",
					NextStepID: "main",
				},
			},
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("invalid config: %+v", err)
	}

	//create new session
	ss := jsessionmem.New()
	s := ss.Get("")

	//start with landing page then process user inputs
	start(t, cfg, s, "MAIN MENU")
	next(t, cfg, s, "1", "Enter your name")
	next(t, cfg, s, "Jan", "Registered")
	next(t, cfg, s, "1", "MAIN MENU")
	next(t, cfg, s, "2", "your name is jan")
}

func start(t *testing.T, c config.Config, s jsessions.ISession, exp string) {
	t.Logf("start")
	stepID := c.Start
	stepConfig, ok := c.Step(stepID)
	if !ok {
		t.Fatalf("unknown step \"%s\"", stepID)
	}
	step := stepConfig.Step().(config.IUserStep)

	/*promptContent,_ :=*/
	step.Content(nil)
	///...to change...
	// t.Logf("PROMPT: %s", prompt)
	// if strings.Index(string(prompt), exp) < 0 {
	// 	t.Fatalf("expected: %s", exp)
	// }
	// s.SetString("app.step", stepID)
}

func next(t *testing.T, c config.Config, s jsessions.ISession, input string, exp string) {
	t.Logf("input=\"%s\"", input)
	stepID, _ := s.GetString("app.step")
	stepConfig, ok := c.Step(stepID)
	if !ok {
		t.Fatalf("unknown step \"%s\"", stepID)
	}
	step := stepConfig.Step().(config.IUserStep)

	nextStepID, err := step.Exec(s, map[string]interface{}{"user": input})
	if err != nil {
		t.Fatalf("step(%s).input(%s) -> Error %+v", stepID, input, err)
	}

	//todo: if final...

	//go to next step
	stepConfig, ok = c.Step(nextStepID)
	if !ok {
		t.Fatalf("step(%s).input(%s) -> unknown next step(%s)", stepID, input, nextStepID)
	}
	s.SetString("app.step", nextStepID)
	step = stepConfig.Step().(config.IUserStep)
	stepID = nextStepID

	//check expected prompt
	/*prompt,_ := */
	step.Content(nil)
	//todo change
	// t.Logf("PROMPT: %s", prompt)
	// if strings.Index(strings.ToLower(string(prompt)), strings.ToLower(exp)) < 0 {
	// 	t.Fatalf("expected(%s) not found in prompt(%s)", exp, prompt)
	// }
}
