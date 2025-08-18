package main

import (
    "fmt"
    "testing"

    tea "github.com/charmbracelet/bubbletea"
)

type fakeRunner struct{}

func (fakeRunner) Run() (tea.Model, error) { return nil, nil }

func TestMain_WiresProgramWithoutRunningTUI(t *testing.T) {
    prev := newRunner
    newRunner = func(m tea.Model) runner { return fakeRunner{} }
    defer func() { newRunner = prev }()

    // Should not panic or exit
    main()
}

type fakeErrRunner struct{}
func (fakeErrRunner) Run() (tea.Model, error) { return nil, fmt.Errorf("boom") }

func TestMain_ErrorExit(t *testing.T) {
    prevRunner := newRunner
    prevExit := exitMain
    newRunner = func(m tea.Model) runner { return fakeErrRunner{} }
    var code int
    exitMain = func(c int) { code = c }
    defer func() { newRunner = prevRunner; exitMain = prevExit }()

    main()
    if code != 1 {
        t.Fatalf("expected exit code 1, got %d", code)
    }
}
