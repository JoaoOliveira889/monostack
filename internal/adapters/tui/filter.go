package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
)

type filterState struct {
	input  textinput.Model
	active bool
	query  string
}

func newFilterState(input textinput.Model) filterState {
	return filterState{input: input}
}

func (f *filterState) Activate() {
	f.active = true
	f.input.SetValue("")
	f.input.Focus()
}

func (f *filterState) Deactivate() {
	f.active = false
	f.query = ""
	f.input.Blur()
	f.input.SetValue("")
}

func (f *filterState) Commit() {
	f.query = f.input.Value()
	f.input.Blur()
}

func (f *filterState) Update(msg tea.KeyMsg) {
	f.input, _ = f.input.Update(msg)
	f.query = f.input.Value()
}

func (f *filterState) IsActive() bool { return f.active }

func (f *filterState) HasQuery() bool { return f.query != "" }

func (f *filterState) Query() string { return f.query }

func (f *filterState) Input() textinput.Model { return f.input }
