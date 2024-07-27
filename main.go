package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	choices        []string
	buttons        []string
	cursor         int
	selected       map[int]struct{}
	terminalWidth  int
	terminalHeight int
	response       string
}

func InitialModel() model {
	return model{
		choices: []string{
			"Tea",
			"Coffee",
			"Milk",
		},
		buttons: []string{
			"Submit Option 1",
			"Submit Option 2",
			"Submit Option 3",
		},
		selected: make(map[int]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)+len(m.buttons)-1 {
				m.cursor++
			}
		case "enter", " ":
			if m.cursor >= len(m.choices) {
				// Обработка нажатия кнопки
				buttonIndex := m.cursor - len(m.choices)
				return m, m.sendPostRequest(buttonIndex)
			} else {
				// Обработка выбора чекбокса
				_, ok := m.selected[m.cursor]
				if ok {
					delete(m.selected, m.cursor)
				} else {
					m.selected[m.cursor] = struct{}{}
				}
			}
		}
	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height
	}
	return m, nil
}

func (m model) sendPostRequest(index int) tea.Cmd {
	endpoints := []string{
		"http://localhost:8000/endpoint1",
		"http://localhost:8000/endpoint2",
		"http://localhost:8000/endpoint3",
	}

	if index < 0 || index >= len(endpoints) {
		return nil
	}

	return func() tea.Msg {
		resp, err := http.Post(endpoints[index], "application/json", nil)
		if err != nil {
			return fmt.Sprintf("POST request failed: %v", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Sprintf("Reading response failed: %v", err)
		}

		return string(body)
	}
}

func (m model) View() string {
	title := " What would you like to do? "
	borderWidth := m.terminalWidth - 2

	if borderWidth < len(title) {
		return "Error: Terminal width is too small"
	}

	borderStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FFB6C1"))

	sideBorder := lipgloss.NewStyle().
		BorderStyle(lipgloss.Border{Left: borderStyle.GetBorderStyle().Left, Right: borderStyle.GetBorderStyle().Right})

	topLeft := borderStyle.GetBorderStyle().TopLeft
	topRight := borderStyle.GetBorderStyle().TopRight
	bottomLeft := borderStyle.GetBorderStyle().BottomLeft
	bottomRight := borderStyle.GetBorderStyle().BottomRight
	topBorder := borderStyle.GetBorderStyle().Top
	bottomBorder := borderStyle.GetBorderStyle().Bottom
	leftBorder := borderStyle.GetBorderStyle().Left
	rightBorder := borderStyle.GetBorderStyle().Right

	leftTitleBorder := strings.Repeat(topBorder, (borderWidth-len(title))/2)
	rightTitleBorder := strings.Repeat(topBorder, (borderWidth-len(title))/2)
	topLine := fmt.Sprintf("%s%s%s%s%s", topLeft, leftTitleBorder, title, rightTitleBorder, topRight)
	bottomLine := fmt.Sprintf("%s%s%s", bottomLeft, strings.Repeat(bottomBorder, borderWidth), bottomRight)

	cursorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
	checkedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#98FB98"))
	choiceStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#87CEFA"))
	buttonStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6347"))

	var leftColumn, rightColumn strings.Builder
	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = cursorStyle.Render(">")
		}
		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = checkedStyle.Render("x")
		}
		line := fmt.Sprintf("%s [%s] %s", cursor, checked, choiceStyle.Render(choice))
		leftColumn.WriteString(line + "\n")
	}

	for i, button := range m.buttons {
		cursor := " "
		if m.cursor == len(m.choices)+i {
			cursor = cursorStyle.Render(">")
		}
		line := fmt.Sprintf("%s %s", cursor, buttonStyle.Render(button))
		rightColumn.WriteString(line + "\n")
	}

	leftColWidth := (borderWidth - 1) / 2
	rightColWidth := borderWidth - leftColWidth - 1

	leftColLines := strings.Split(leftColumn.String(), "\n")
	rightColLines := strings.Split(rightColumn.String(), "\n")

	var combinedContent strings.Builder
	maxLines := max(len(leftColLines), len(rightColLines))

	for i := 0; i < maxLines; i++ {
		var leftLine, rightLine string
		if i < len(leftColLines) {
			leftLine = leftColLines[i]
		}
		if i < len(rightColLines) {
			rightLine = rightColLines[i]
		}
		leftLine = padRight(leftLine, leftColWidth)
		rightLine = padRight(rightLine, rightColWidth)
		combinedContent.WriteString(fmt.Sprintf("%s%s %s%s\n", leftBorder, leftLine, rightLine, rightBorder))
	}

	// responseLine := fmt.Sprintf("%s%s%s", leftBorder, m.response, strings.Repeat(" ", borderWidth-len(m.response)+1)+rightBorder)
	combinedContent.WriteString(fmt.Sprintf("%s\n", sideBorder.GetBorderStyle().Left))

	return fmt.Sprintf("%s\n%s%s", topLine, combinedContent.String(), bottomLine)
}

func padRight(str string, length int) string {
	return str + strings.Repeat(" ", length-lipgloss.Width(str))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	p := tea.NewProgram(InitialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting program: %v\n", err)
	}
}
