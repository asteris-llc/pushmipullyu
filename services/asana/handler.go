package asana

import (
	"github.com/Sirupsen/logrus"
	"github.com/asteris-llc/pushmipullyu/dispatch"
	"golang.org/x/net/context"
)

// Handler listens to events from other services and turns them into Asana
// service calls.
type Handler struct {
	Asana  *Asana
	logger *logrus.Entry
}

// New returns a new Asana Handler
func New(token string) *Handler {
	return &Handler{
		Asana:  NewAsana(token),
		logger: logrus.WithField("service", "asana"),
	}
}

// Handle a new source of messages
func (h *Handler) Handle(ctx context.Context, in chan dispatch.Message) {
	h.logger.Debug("handling new events")

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("shutting down")
			return
		}
	}
}
