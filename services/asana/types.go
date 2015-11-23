package asana

// NameID is an representation of the fact that tons of Asana resources return a
// list of objects with name and ID
type NameID struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// WrappedNameIDs is a representation of another common pattern: to return a
// list of NameIDs in a "data" key.
type WrappedNameIDs struct {
	Data []NameID `json:"data"`
}

// Responses

type teamResponse struct {
	Data struct {
		Organization NameID `json:"organization"`
	} `json:"data"`
}

// Requests

type dataWrapper struct {
	Data interface{} `json:"data"`
}

type createTask struct {
	Assignee  int    `json:"assignee"`
	Name      string `json:"name"`
	Notes     string `json:"notes"`
	Projects  []int  `json:"projects"`
	Workspace int    `json:"workspace"`
}
