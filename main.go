package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"

	"github.com/go-msvc/errors"
	"github.com/go-msvc/japp/config"
	"github.com/go-msvc/japp/msg"
	"github.com/go-msvc/japp/users"
	jsonfileusers "github.com/go-msvc/japp/users/jsonfile"
	"github.com/go-msvc/jsessions"
	memsessions "github.com/go-msvc/jsessions/mem"
	"github.com/go-msvc/logger"
	"github.com/gorilla/pat"
	"github.com/satori/uuid"
)

func main() {
	configFileName := flag.String("config", "./conf/app.json", "Config file")
	if err := flag.CommandLine.Parse(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid options: %+v\n", err)
		flag.CommandLine.Usage()
		os.Exit(1)
	}

	cfg, err := loadConfig(*configFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Config error: %+v\n", err)
		os.Exit(1)
	}

	app, err := newApp(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create configured app: %+v\n", err)
		os.Exit(1)
	}
	app.Debugf("Starting...")
	http.ListenAndServe("localhost:12345", app)
}

func newApp(cfg config.Config) (app, error) {
	uu, err := jsonfileusers.Load("./data/users.json")
	if err != nil {
		return app{}, errors.Wrapf(err, "cannot open users store")
	}
	r := pat.New()
	app := app{
		Handler:  r,
		ILogger:  logger.Top().NewLogger("japp").WithStream(logger.Terminal(logger.LogLevelDebug)),
		cfg:      cfg,
		sessions: memsessions.New(),
		users:    uu,
	}
	r.Post("/app/start", app.handler(msg.StartRequest{}, msg.StartResponse{}, app.start))
	r.Post("/app/cont", app.handler(msg.ContinueRequest{}, msg.ContinueResponse{}, app.cont))
	r.Post("/app/end", app.handler(msg.EndRequest{}, msg.EndResponse{}, app.end))
	r.Post("/", app.unknownHandler)
	r.Get("/", app.unknownHandler)
	return app, nil
}

type app struct {
	http.Handler
	logger.ILogger
	cfg      config.Config
	sessions jsessions.ISessions
	users    users.IUsers
}

func (app app) unknownHandler(httpRes http.ResponseWriter, httpReq *http.Request) {
	http.Error(httpRes, "unknown URL", http.StatusNotFound)
}

type appFunc func(req interface{}) (res interface{}, err error)

func (app app) handler(appReqTmpl, appRes interface{}, appFunc appFunc) http.HandlerFunc {
	return func(httpRes http.ResponseWriter, httpReq *http.Request) {
		status := http.StatusOK
		var err error

		defer func() {
			if status != http.StatusOK {
				app.Errorf("HTTP %s %s -> %s: %+v", httpReq.Method, httpReq.URL.Path, http.StatusText(status), err)
				http.Error(httpRes, fmt.Sprintf("%+v", err), status)
			}
		}()

		if httpReq.Method != http.MethodPost {
			err = errors.Errorf("invalid method != POST")
			status = http.StatusMethodNotAllowed
			return
		}

		var jsonReq []byte
		jsonReq, err = ioutil.ReadAll(httpReq.Body)
		if err != nil {
			err = errors.Wrapf(err, "failed to read HTTP POST body")
			status = http.StatusBadRequest
			return
		}

		appReqValue := reflect.New(reflect.TypeOf(appReqTmpl))
		if err = json.Unmarshal(jsonReq, appReqValue.Interface()); err != nil {
			err = errors.Wrapf(err, "cannot parse JSON body")
			status = http.StatusBadRequest
			return
		}

		if validator, ok := appReqValue.Interface().(msg.IValidator); ok {
			if err = validator.Validate(); err != nil {
				err = errors.Wrapf(err, "invalid request: %+v", appReqValue.Interface())
				status = http.StatusBadRequest
				return
			}
		}

		appReq := appReqValue.Elem().Interface()
		app.Debugf("appReq: (%T) %+v", appReq, appReq)

		if appRes, err = appFunc(appReq); err != nil {
			err = errors.Wrapf(err, "handler failed")
			status = http.StatusInternalServerError
			return
		}

		//output
		app.Debugf("appRes: (%T) %+v", appRes, appRes)
		jsonRes, _ := json.Marshal(appRes)
		httpRes.Header().Set("Content-Type", "application/json")
		httpRes.Write(jsonRes)
		app.Debugf("JSON response: (%T) %+v", appRes, string(jsonRes))
		status = http.StatusOK
		err = nil
		return
	}
}

func (app app) start(appReq interface{}) (appRes interface{}, err error) {
	startReq := appReq.(msg.StartRequest)
	if startReq.ClientID == "" {
		startReq.ClientID = uuid.NewV1().String()
	}
	session := app.sessions.Get("")
	if session == nil {
		return nil, errors.Errorf("failed to start a new session")
	}
	app.Debugf("session(%s) started", session.ID())
	session.Set("client-id", startReq.ClientID)

	//execute from app start to first content
	userStepID, userStep, err := app.exec(session, "", "")
	if err != nil {
		return nil, errors.Wrapf(err, "failed to start app")
	}

	content, err := userStep.Content(userStepID, session.Data())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to render content")
	}

	return msg.StartResponse{
		ClientID:  startReq.ClientID, //specified or new id
		SessionID: session.ID(),      //new session id
		Content:   content,           //app start content
	}, nil
} //app.start()

func (app app) cont(appReq interface{}) (appRes interface{}, err error) {
	contReq, ok := appReq.(msg.ContinueRequest)
	if !ok {
		return nil, errors.Errorf("%T is not contRequest", appReq)
	}
	session := app.sessions.Get(contReq.SessionID)
	if session == nil {
		return nil, errors.Errorf("session(%s) not found", contReq.SessionID)
	}
	defer session.Save()

	input, ok := contReq.Data["input"].(string)
	if !ok {
		return nil, errors.Errorf("no input")
	}

	userStepID, userStep, err := app.exec(session, contReq.StepID, input)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to continue app")
	}

	content, err := userStep.Content(userStepID, session.Data())
	if err != nil {
		return nil, errors.Wrapf(err, "failed to render content")
	}

	return msg.ContinueResponse{
		SessionID: session.ID(),
		Content:   content,
	}, nil
} //app.cont()

func (app app) end(appReq interface{}) (appRes interface{}, err error) {
	return nil, errors.Errorf("NYI")
} //app.end()

func (app app) exec(session jsessions.ISession, stepID, input string) (userStepID string, userStep config.IUserStep, err error) {
	sessionStepID, _ := session.GetString("app.step")
	//cannot apply input for another step to this step
	if stepID != sessionStepID {
		app.Errorf("session.step(%s) != input.step(%s)", sessionStepID, stepID)
		return "", nil, errors.Errorf("session.step(%s) != input.step(%s)", sessionStepID, stepID)
	}
	if sessionStepID == "" {
		sessionStepID, _ = session.SetString("app.step", app.cfg.Start)
		stepID = sessionStepID
	}
	app.Debugf("exec from stepID=\"%s\"", stepID)

	stepConfig, ok := app.cfg.Steps[stepID]
	if !ok {
		app.Errorf("step(%s) not found", stepID)
		return "", nil, errors.Wrapf(err, "step(%s) not found", stepID)
	}

	step := stepConfig.Step()
	if !ok {
		app.Errorf("step(%s) not a step", stepID)
		return "", nil, errors.Wrapf(err, "step(%s) not a step", stepID)
	}

	visitedStepID := map[string]bool{}
	for {
		app.Debugf("loop step(%s)", stepID)

		nextStepID := ""
		if userStep, ok := step.(config.IUserStep); ok {
			//user step
			if input == "" {
				return stepID, userStep, nil //show to user
			}

			//has input, apply input and proceed
			app.Debugf("exec user step(%s).input(%s)...", stepID, input)
			nextStepID, err = step.Exec(session, map[string]interface{}{"user": input})
			if err != nil {
				return "", nil, errors.Wrapf(err, "user step(%s).input(%s) failed", stepID, input)
			}
			app.Debugf("step(%s).input(%s) -> next(%s)", stepID, input, nextStepID)
			input = "" //clear so need to prompt for next user step
		} else {
			//non-user step - may not exec more than once the same step
			//but user steps are allowed to return to because they will not loop but stop for input
			if alreadyVisited, _ := visitedStepID[stepID]; alreadyVisited {
				return "", nil, errors.Errorf("loop on step(%s) already visited", stepID)
			}
			visitedStepID[stepID] = true

			nextStepID, err = step.Exec(session, map[string]interface{}{})
			if err != nil {
				return "", nil, errors.Wrapf(err, "non-user step(%s) failed", stepID)
			}
			app.Debugf("step(%s) -> next(%s)", stepID, nextStepID)
		}

		if nextStepID == "" {
			return "", nil, errors.Errorf("step(%s) did not return next step", stepID)
		}
		nextStepConfig, ok := app.cfg.Steps[nextStepID]
		if !ok {
			return "", nil, errors.Errorf("step(%s)->unknown next step(%s)", stepID, nextStepID)
		}

		step = nextStepConfig.Step()
		if step == nil {
			return "", nil, errors.Errorf("failed to get next step-config.step")
		}

		stepID, _ = session.SetString("app.step", nextStepID)
	} //step loop
} //app.exec()

func loadConfig(fn string) (config.Config, error) {
	configFile, err := os.Open(fn)
	if err != nil {
		return config.Config{}, errors.Wrapf(err, "failed to open %s", fn)
	}
	defer configFile.Close()

	cfg := config.Config{}
	if err := json.NewDecoder(configFile).Decode(&cfg); err != nil {
		return config.Config{}, errors.Wrapf(err, "failed to read JSON from file")
	}
	if err := cfg.Validate(); err != nil {
		return config.Config{}, errors.Wrapf(err, "invalid config")
	}
	return cfg, nil
} //loadConfig()
