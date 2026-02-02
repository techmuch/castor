package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/techmuch/castor/pkg/agent"
)

type errMsg error

type model struct {
	viewport    viewport.Model
	messages    []string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	err         error
	agent       *agent.Agent
}

func InitialModel(ag *agent.Agent) model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)
	vp.SetContent(`Welcome to Castor TUI!`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		textarea:    ta,
		viewport:    vp,
		messages:    []string{},
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		agent:       ag,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

type agentResponseMsg struct {
	text string
	err  error
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - m.textarea.Height() - 2
		m.textarea.SetWidth(msg.Width)
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			input := m.textarea.Value()
			if input == "" {
				return m, nil
			}
			m.messages = append(m.messages, m.senderStyle.Render("You: ")+input)
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.textarea.Reset()
			m.viewport.GotoBottom()

			// Start agent chat
			return m, func() tea.Msg {
				ctx := context.Background()
				stream, err := m.agent.Chat(ctx, input)
				if err != nil {
					return agentResponseMsg{err: err}
				}
				
				var fullContent strings.Builder
				for event := range stream {
					if event.Error != nil {
						return agentResponseMsg{err: event.Error}
					}
					fullContent.WriteString(event.Delta)
				}
				return agentResponseMsg{text: fullContent.String()}
			}
		}
	case agentResponseMsg:
		if msg.err != nil {
			m.messages = append(m.messages, "Error: "+msg.err.Error())
		} else {
			m.messages = append(m.messages, "Castor: "+msg.text)
		}
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	) + "\n\n"
}

// Run starts the TUI
func Run(ag *agent.Agent) error {
	p := tea.NewProgram(InitialModel(ag), tea.WithAltScreen())
	_, err := p.Run()
	return err
}