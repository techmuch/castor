package tui

import (
	"context"
	"fmt"
	"sort"
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
	botStyle    lipgloss.Style
	sysStyle    lipgloss.Style
	err         error
	agent       *agent.Agent
}

func InitialModel(ag *agent.Agent) model {
	ta := textarea.New()
	ta.Placeholder = "Send a message or type /help..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 1000 // Increased limit

	ta.SetWidth(30)
	ta.SetHeight(3)

	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)
	vp.SetContent(lipgloss.NewStyle().Foreground(lipgloss.Color("63")).Render(`Welcome to Castor TUI!
Type /help to see available commands.`))

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		textarea:    ta,
		viewport:    vp,
		messages:    []string{},
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true),
		botStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
		sysStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true),
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
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			input := strings.TrimSpace(m.textarea.Value())
			if input == "" {
				return m, nil
			}

			// Handle Slash Commands
			if strings.HasPrefix(input, "/") {
				m.textarea.Reset()
				return m.handleCommand(input)
			}

			// Regular Chat
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
					// We could stream tool calls here too if we update the event type
				}
				return agentResponseMsg{text: fullContent.String()}
			}
		}
	case agentResponseMsg:
		if msg.err != nil {
			m.messages = append(m.messages, m.sysStyle.Render("Error: "+msg.err.Error()))
		} else {
			m.messages = append(m.messages, m.botStyle.Render("Castor: ")+msg.text)
		}
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) handleCommand(input string) (tea.Model, tea.Cmd) {
	parts := strings.Fields(input)
	cmd := parts[0]
	// args := parts[1:] // For future use

	var output string

	switch cmd {
	case "/quit", "/exit":
		return m, tea.Quit
	case "/clear":
		m.messages = []string{}
		m.viewport.SetContent("Chat cleared.")
		return m, nil
	case "/help":
		output = `Available Commands:
  /tools   - List all available tools
  /sys     - Show current system prompt
  /clear   - Clear chat history
  /help    - Show this help message
  /quit    - Exit the application`
	case "/tools":
		output = "Available Tools:\n"
		var tools []string
		for name, t := range m.agent.Tools {
			tools = append(tools, fmt.Sprintf("• %s: %s", name, t.Description()))
		}
		sort.Strings(tools)
		if len(tools) == 0 {
			output += "  (No tools registered)"
		} else {
			output += strings.Join(tools, "\n")
		}
	case "/sys":
		output = fmt.Sprintf("System Prompt:\n%s", m.agent.SystemPrompt)
	default:
		output = fmt.Sprintf("Unknown command: %s. Type /help for list.", cmd)
	}

	m.messages = append(m.messages, m.sysStyle.Render(output))
	m.viewport.SetContent(strings.Join(m.messages, "\n"))
	m.viewport.GotoBottom()
	return m, nil
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
