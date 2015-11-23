package asana

import (
	"encoding/json"
	"fmt"
	"github.com/asteris-llc/pushmipullyu/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/asteris-llc/pushmipullyu/Godeps/_workspace/src/golang.org/x/net/context"
	"github.com/asteris-llc/pushmipullyu/dispatch"
	"github.com/asteris-llc/pushmipullyu/services/github"
	"math/rand"
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
	if team.Data.Organization.ID == 0 {
		return nil, fmt.Errorf("could not get organization: %s", jsonResponse)
	}
	handler.Organization = team.Data.Organization.ID

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
				h.logger.WithField("projects", projects).Debug("got projects")

				var (
					project      NameID
					foundProject = false
				)
				for _, potential := range projects.Data {
					if potential.Name == issue.Repository.Name {
						project = potential
						foundProject = true
						break
					}
				}
				if !foundProject {
					h.logger.WithField("project", issue.Repository.Name).Debug("destination project not found")
					continue
				}

				// get users
				_, usersBody, errs := h.Asana.Client().Get(fmt.Sprintf("https://app.asana.com/api/1.0/teams/%d/users", h.Team)).End()
				if errs != nil {
					for _, err := range errs {
						h.logger.WithField("error", err).Error("error retrieving users")
					}
					continue
				}

				users := &WrappedNameIDs{}
				err = json.Unmarshal([]byte(usersBody), users)
				if err != nil {
					h.logger.WithField("error", err).Error("error parsing users")
					continue
				}
				h.logger.WithField("users", users).Debug("got users")

				user := users.Data[rand.Intn(len(users.Data))]

				// prepare request
				task := createTask{
					Assignee:  user.ID,
					Name:      fmt.Sprintf("%s (#%d)", issue.Issue.Title, issue.Issue.Number),
					Notes:     fmt.Sprintf("%s\n\nOpened by %s at %s\n\n%s", issue.Issue.URL, issue.Issue.User.Login, issue.Issue.CreatedAt, issue.Issue.Body),
					Projects:  []int{project.ID},
					Workspace: h.Organization,
				}

				h.logger.WithField("task", task).Debug("task")

				taskJSON, err := json.Marshal(dataWrapper{task})
				if err != nil {
					h.logger.WithField("error", err).Error("error marshalling task JSON")
					continue
				}

				_, body, errs := h.Asana.Client().
					SendRawBytes(taskJSON).
					Post("https://app.asana.com/api/1.0/tasks").
					End()

				h.logger.WithField("body", body).Debug("created a task")

				for _, err := range errs {
					h.logger.WithField("error", err).Error("error creating task")
				}

			default:
				h.logger.WithField("tag", event.Tag).Debug("not handling tag")
			}
		}
	}
}
