package main

import (
	"go/ast"

	"github.com/marstr/collection"
)

type PackageWalker struct {
	target ast.Node
}

func (pw PackageWalker) Enumerate(cancel <-chan struct{}) collection.Enumerator {
	walker := newDepthBoundPlunger(2)
	go func() {
		defer walker.Dispose()

		ast.Walk(walker, pw.target)
	}()

	return walker.Results()
}

type depthBoundPlunger struct {
	currentDepth uint
	MaxDepth     uint
	results      chan interface{}
	Cancel       <-chan struct{}
}

func newDepthBoundPlunger(maxDepth uint) *depthBoundPlunger {
	return &depthBoundPlunger{
		currentDepth: 0,
		MaxDepth:     maxDepth,
		results:      make(chan interface{}),
	}
}

func (plunger depthBoundPlunger) Results() collection.Enumerator {
	return plunger.results
}

func (plunger depthBoundPlunger) Dispose() {
	close(plunger.results)
}

func (plunger depthBoundPlunger) Visit(node ast.Node) (w ast.Visitor) {
	w = nil
	if plunger.currentDepth > plunger.MaxDepth {
		return
	}

	select {
	case <-plunger.Cancel:
		return
	case plunger.results <- node:
		// Intentionally Left Blank
	}

	w = depthBoundPlunger{
		Cancel:       plunger.Cancel,
		currentDepth: plunger.currentDepth + 1,
		MaxDepth:     plunger.MaxDepth,
		results:      plunger.results,
	}

	return
}
