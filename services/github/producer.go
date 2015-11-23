package github

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

// Producer produces events from the event hook
type Producer struct {
	Port   string
	logger *logrus.Entry
}

// New returns a new producer
func New(listen string) (*Producer, error) {
	return &Producer{
		Port:   listen,
		logger: logrus.WithField("service", "github"),
	}, nil
}

// Produce starts the production
func (p *Producer) Produce(emit func(string, interface{})) error {
	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			p.logger.WithField("error", err).Error("could not read body")
			return
		}
		defer r.Body.Close()

		response := &IssueEvent{}
		err = json.Unmarshal(body, response)
		if err != nil {
			p.logger.WithField("error", err).Error("could not decode body JSON")
			return
		}

		if response.Action == "" {
			p.logger.Error("no action set")
			return
		}

		p.logger.WithField("tag", "github:"+response.Action).Debug("emitting action")
		emit("github:"+response.Action, response)
	})

	return http.ListenAndServe(
		":"+p.Port,
		nil,
	)
}
