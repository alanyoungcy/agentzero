package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type role string

const (
	roleUser      role = "You"
	roleAssistant role = "Agent"
)

type chatMessage struct {
	role    role
	content string
}

type responseMsg struct {
	content string
	err     error
}

type model struct {
	viewport    viewport.Model
	textarea    textarea.Model
	messages    []chatMessage
	history     []llms.MessageContent
	llm         llms.Model
	waiting     bool
	err         error
	width       int
	height      int
	initialized bool
}

var (
	userStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	agentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true)
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Bold(true)
	helpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("5")).
			Bold(true).
			Padding(0, 1)
)

func initLLM() (llms.Model, error) {
	opts := []openai.Option{}

	if baseURL := os.Getenv("OPENAI_BASE_URL"); baseURL != "" {
		opts = append(opts, openai.WithBaseURL(baseURL))
	}
	if model := os.Getenv("OPENAI_MODEL"); model != "" {
		opts = append(opts, openai.WithModel(model))
	}

	return openai.New(opts...)
}

func initialModel(llm llms.Model) model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()
	ta.CharLimit = 0
	ta.SetWidth(80)
	ta.SetHeight(3)
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.SetEnabled(false)

	vp := viewport.New(80, 20)

	return model{
		textarea: ta,
		viewport: vp,
		messages: []chatMessage{},
		history:  []llms.MessageContent{},
		llm:      llm,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		taCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, taCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.initialized = true

		headerHeight := 1
		footerHeight := 1
		inputHeight := 5
		verticalMargin := headerHeight + footerHeight + inputHeight

		m.textarea.SetWidth(msg.Width - 2)
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - verticalMargin
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.waiting {
				return m, nil
			}
			input := strings.TrimSpace(m.textarea.Value())
			if input == "" {
				return m, nil
			}
			if input == "/quit" || input == "/exit" {
				return m, tea.Quit
			}
			if input == "/clear" {
				m.messages = []chatMessage{}
				m.history = []llms.MessageContent{}
				m.textarea.Reset()
				m.viewport.SetContent(m.renderMessages())
				return m, nil
			}

			m.messages = append(m.messages, chatMessage{role: roleUser, content: input})
			m.history = append(m.history, llms.TextParts(llms.ChatMessageTypeHuman, input))
			m.textarea.Reset()
			m.waiting = true
			m.viewport.SetContent(m.renderMessages())
			m.viewport.GotoBottom()

			return m, m.sendMessage()
		}

	case responseMsg:
		m.waiting = false
		if msg.err != nil {
			m.messages = append(m.messages, chatMessage{
				role:    roleAssistant,
				content: fmt.Sprintf("[Error: %v]", msg.err),
			})
		} else {
			m.messages = append(m.messages, chatMessage{
				role:    roleAssistant,
				content: msg.content,
			})
			m.history = append(m.history, llms.TextParts(llms.ChatMessageTypeAI, msg.content))
		}
		m.viewport.SetContent(m.renderMessages())
		m.viewport.GotoBottom()
	}

	return m, tea.Batch(taCmd, vpCmd)
}

func (m model) sendMessage() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		resp, err := m.llm.GenerateContent(ctx, m.history)
		if err != nil {
			return responseMsg{err: err}
		}
		if len(resp.Choices) == 0 {
			return responseMsg{err: fmt.Errorf("no response from model")}
		}
		return responseMsg{content: resp.Choices[0].Content}
	}
}

func (m model) renderMessages() string {
	if len(m.messages) == 0 {
		return helpStyle.Render("  No messages yet. Start chatting!")
	}

	var sb strings.Builder
	for _, msg := range m.messages {
		switch msg.role {
		case roleUser:
			sb.WriteString(userStyle.Render("You: "))
		case roleAssistant:
			sb.WriteString(agentStyle.Render("Agent: "))
		}
		sb.WriteString(msg.content)
		sb.WriteString("\n\n")
	}

	if m.waiting {
		sb.WriteString(agentStyle.Render("Agent: "))
		sb.WriteString(helpStyle.Render("thinking..."))
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m model) View() string {
	if !m.initialized {
		return "Initializing..."
	}

	header := titleStyle.Render(" AgentZero Chat ")
	footer := helpStyle.Render(" enter: send | /clear: reset | /quit: exit | esc: quit ")

	return fmt.Sprintf(
		"%s\n%s\n%s\n%s",
		header,
		m.viewport.View(),
		m.textarea.View(),
		footer,
	)
}

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: no .env file found")
	}

	llm, err := initLLM()
	if err != nil {
		log.Fatalf("Failed to initialize LLM: %v", err)
	}

	p := tea.NewProgram(
		initialModel(llm),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
