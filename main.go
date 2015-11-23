package main

import (
	"golang.org/x/net/context"
)

func main() {
	ctx, shutdown := context.WithCancel(context.Background())
	dispatch := NewDispatch(ctx)
}
