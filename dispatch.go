package main

import (
	"golang.org/x/net/context"
)

// Dispatch manages a dispatcher function. It handles registration.
type Dispatch struct {
	in       chan message
	handlers map[string][]chan interface{}
}

type message struct {
	Tag     string
	Payload interface{}
}

// NewDispatch creates a new empty dispatch and starts it running
func NewDispatch() *Dispatch {
	return &Dispatch{
		in:       make(chan message, 0),
		handlers: map[string][]chan interface{}{},
	}
}

// Send takes an event tag and a payload and dispatches it as soon as possible.
func (d *Dispatch) Send(tag string, payload interface{}) {
	d.in <- message{tag, payload}
}

// Register registers a function for a tag and returns a channel on which events
// will be sent for that tag.
func (d *Dispatch) Register(tag string) chan interface{} {
	out := make(chan interface{}, 1)

	current, ok := d.handlers[tag]
	if ok {
		d.handlers[tag] = append(current, out)
	} else {
		d.handlers[tag] = []chan interface{}{out}
	}

	return out
}

// Run begins the dispatch process
func (d *Dispatch) Run(ctx context.Context) {
	for {
		select {
		case value := <-d.in:
			if targets, ok := d.handlers[value.Tag]; ok {
				for _, t := range targets {
					t <- value.Payload
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
