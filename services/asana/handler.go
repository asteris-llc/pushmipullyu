package asana

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/asteris-llc/pushmipullyu/dispatch"
	"github.com/asteris-llc/pushmipullyu/services/github"
	"golang.org/x/net/context"
	"strconv"
)

// Handler listens to events from other services and turns them into Asana
// service calls.
type Handler struct {
	Asana        *Asana
	Team         int
	Organization int

	logger *logrus.Entry
}

// New returns a new Asana Handler
func New(token string, teamID int) (*Handler, error) {
	handler := &Handler{
		Asana:  NewAsana(token),
		Team:   teamID,
		logger: logrus.WithField("service", "asana").WithField("team", teamID),
	}

	// get the organization from the team
	_, jsonResponse, errs := handler.Asana.Client().Get(fmt.Sprintf("https://app.asana.com/api/1.0/teams/%d", teamID)).End()

	if errs != nil {
		return nil, errs[0]
	}

	team := &teamResponse{}
	err := json.Unmarshal([]byte(jsonResponse), team)
	if err != nil {
		return nil, err
	}
	handler.Organization = team.Organization.ID

	return handler, nil
}

// Handle a new source of messages
func (h *Handler) Handle(ctx context.Context, in chan dispatch.Message) {
	h.logger.Debug("handling new events")

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("shutting down")
			return

		case event := <-in:
			switch event.Tag {
			case "github:opened":
				issue, ok := event.Payload.(*github.IssueEvent)
				if !ok {
					h.logger.Warn("failed to convert to *github.Issue")
					continue
				}

				// figure out which project to add to
				_, projectsBody, errs := h.Asana.Client().Get(fmt.Sprintf("https://app.asana.com/api/1.0/teams/%d/projects", h.Team)).End()
				if errs != nil {
					for _, err := range errs {
						h.logger.WithField("error", err).Error("error retrieving projects")
					}
					continue
				}

				projects := &WrappedNameIDs{}
				err := json.Unmarshal([]byte(projectsBody), projects)
				if err != nil {
					h.logger.WithField("error", err).Error("error parsing projects")
					continue
				}

				var project *NameID
				for _, potential := range projects.Data {
					if potential.Name == issue.Repository.Name {
						project = &potential
						break
					}
				}
				if project == nil {
					h.logger.WithField("project", issue.Repository.Name).Debug("no matching project")
					continue
				}

				// TODO: get users and assign a random one

				// prepare request
				_, _, errs = h.Asana.Client().
					SetHeader("name", fmt.Sprintf("%s (%d)", issue.Issue.Title, issue.Issue.Number)).
					// SetHeader("assignee", ).
					SetHeader("notes", fmt.Sprintf("%s\n\nOpened by %s at %s\n\n%s", issue.Issue.URL, issue.Issue.User.Login, issue.Issue.CreatedAt, issue.Issue.Body)).
					SetHeader("workspace", strconv.Itoa(h.Organization)).
					SetHeader("projects[0]", strconv.Itoa(project.ID)).
					Post("https://app.asana.com/api/1.0/tasks").
					End()

				for _, err := range errs {
					h.logger.WithField("error", err).Error("error creating task")
				}

			default:
				h.logger.WithField("tag", event.Tag).Debug("not handling tag")
			}
		}
	}
}

// curl -v -H "Authorization: Bearer $ASANA_TKN" https://app.asana.com/api/1.0/tasks -X POST -H "Content-Type: application/json" -d '{"data":{"assignee":44904587765688,"name":"Buy Catnip","workspace":44904596360873,"projects":[68060200860521]}}'

// curl -H "Authorization: Bearer $ASANA_TKN" https://app.asana.com/api/1.0/teams/68060200860520/projects
