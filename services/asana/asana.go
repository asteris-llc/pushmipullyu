package asana

import (
	"github.com/asteris-llc/pushmipullyu/Godeps/_workspace/src/github.com/smallnest/goreq"
)

// Asana works with the Asana API
type Asana struct {
	Token string
}

// NewAsana returns a new instance of the Asana client
func NewAsana(token string) *Asana {
	return &Asana{token}
}

func (a *Asana) Client() *goreq.GoReq {
	return goreq.New().SetHeader("Authorization", "Bearer "+a.Token)
}
