// Package repl IS STILL A WIP. It's meant to provide a REPL like CLI experience for Scribe.
// It's not currently used in the CLI, but hopefully will be soon.
package repl

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	errMsg          error
	onSendCompleted struct {
		results string
		err     error
	}
)

type model struct {
	onSend       func(input string) (string, error)
	viewport     viewport.Model
	latestInput  string
	latestOutput string
	textarea     textarea.Model
	senderStyle  lipgloss.Style
	spinner      spinner.Model
	err          error
}

func New(onSend func(input string) (string, error)) model {
	ta := textarea.New()
	ta.Placeholder = "Compose a prompt..."
	ta.Focus()

	ta.Prompt = "> "
	ta.CharLimit = 280

	ta.SetWidth(100)
	ta.SetHeight(2)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(0, 10)
	vp.SetContent(`Ready for prompts!`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		onSend:      onSend,
		textarea:    ta,
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
		spinner:     spinner.New(spinner.WithSpinner(spinner.Meter)),
		err:         nil,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, m.spinner.Tick)
}

func (m model) execOnSend() tea.Msg {
	res, err := m.onSend(m.latestInput)
	if err != nil {
		return errMsg(err)
	}

	return &onSendCompleted{
		results: res,
		err:     err,
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			newInput := m.textarea.Value()
			if newInput == "" {
				return m, nil
			}

			m.latestInput = newInput
			m.latestOutput = ""

			return m, m.execOnSend
		}

	case onSendCompleted:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}

		m.latestOutput = msg.results

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {

	if m.latestOutput == "" {
		return fmt.Sprintf(
			"%s\n\n%s\n\n",
			m.textarea.View(),
			m.viewport.View(),
		)
	}

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		m.spinner.View(),
		m.textarea.View(),
		m.viewport.View(),
	)
}
