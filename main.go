package main

import (
	"encoding/json"
	"fmt"
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
	"time"
)

type Clip struct {
	Content string
	Copied  bool
	Id      int
}

type model struct {
	clipboard_history []Clip
	cursor            int
	copied            map[int]struct{}
	highlighted       bool
}

type tickMsg time.Time

func uploadClips(clips []Clip) string {
	output, _ := json.MarshalIndent(clips, "", " ")
	message := ""
	err := os.WriteFile("clipboard_history.json", output, 0644)
	if err != nil {
		message = "Error adding clips"
	} else {
		message = "Json updated"
	}
	return message
}

func tick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func initialModel(clipboard_history []Clip) model {

	return model{
		clipboard_history: clipboard_history,
	}
}

func clearScreenCmd() tea.Cmd {
	return func() tea.Msg {
		return tea.ClearScreen()
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tick(),
		clearScreenCmd(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			for i := range m.clipboard_history {
				if i == m.cursor {
					m.clipboard_history[i].Copied = true
					clipboard.WriteAll(m.clipboard_history[m.cursor].Content)
				} else {
					m.clipboard_history[i].Copied = false
				}
			}
		case "down", "j":
			if m.cursor < len(m.clipboard_history)-1 {
				m.cursor++
			}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		}
	case tickMsg:
		clipboard_text, _ := clipboard.ReadAll()
		clip_already_exist := false

		for _, clip := range m.clipboard_history {
			if clip.Content == clipboard_text {
				clip_already_exist = true
			}
		}
		newId := 1
		if len(m.clipboard_history) > 0 {
			newId = m.clipboard_history[len(m.clipboard_history)-1].Id + 1
		}
		if !clip_already_exist {
			newClip := Clip{Id: newId, Content: clipboard_text, Copied: true}
			for i := range m.clipboard_history {
				m.clipboard_history[i].Copied = false
			}
			m.clipboard_history = append(m.clipboard_history, newClip)
			uploadClips(m.clipboard_history)
		}
		return m, tick()
	}
	return m, nil
}

var normalStyle = lipgloss.NewStyle()
var greenStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))

func (m model) View() string {
	s := "Tu historial de copiados!\n\n"
	for i, clipboard := range m.clipboard_history {

		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		var text string
		if clipboard.Copied {
			text = greenStyle.Render(clipboard.Content)
		} else {
			text = normalStyle.Render(clipboard.Content)
		}
		s += fmt.Sprintf("%s %s \n", cursor, text)
	}

	// s += "\nPress q to quit. \n"

	return s
}

func main() {
	var clips []Clip
	file, _ := os.ReadFile("clipboard_history.json")
	if len(file) > 0 {
		json.Unmarshal(file, &clips)
	}
	p := tea.NewProgram(initialModel(clips))

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
