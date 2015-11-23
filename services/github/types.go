package github

// Issue is returned from the API
type Issue struct {
	URL       string `json:"html_url"`
	Number    int    `json:"number"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`

	User User `json:"user"`
}

// User is returned from the API
type User struct {
	Login string `json:"login"`
}

// Repository is returned from the API
type Repository struct {
	Name string `json:"name"`
}

// Events

// IssueEvent represents an Issue Event from the API
type IssueEvent struct {
	Action     string     `json:"action"`
	Issue      Issue      `json:"issue"`
	Repository Repository `json:"repository"`
}
