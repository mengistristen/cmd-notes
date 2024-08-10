package tui

import (
	"cmd-notes/note"
	"cmd-notes/utils"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	textInput textinput.Model
	notes     []note.Note
	path      string
	cursor    int
	editing   bool
}

func InitModel(path string, notes []note.Note) Model {
	ti := textinput.New()
	ti.Placeholder = "Task"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return Model{
		textInput: ti,
		notes:     notes,
		path:      path,
		cursor:    0,
		editing:   false,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		default:
			if m.editing {
				return updateEditing(m, msg)
			} else {
				return updateDefault(m, msg)
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	var result string

	if m.editing {
		result = m.textInput.View()
	} else {
		for i, note := range m.notes {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}

			result += fmt.Sprintf("%s %s - %s\n", cursor, note.FormatPriority(), note.FormatContents())
		}
	}

	return result
}

func updateEditing(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.Type {
	case tea.KeyEnter:
		m.editing = false
		m.notes = append(m.notes, note.Note{
			State:    note.NONE,
			Priority: note.MEDIUM,
			Contents: m.textInput.Value(),
		})
		m.textInput.SetValue("")
		utils.WriteState(m.path, &m.notes)
		m.notes = utils.ReadState(m.path)
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func updateDefault(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.notes)-1 {
			m.cursor++
		}
	case ">":
		if len(m.notes) > 0 {
			m.notes[m.cursor].Promote()

			utils.WriteState(m.path, &m.notes)
			m.notes = utils.ReadState(m.path)
		}
	case "<":
		if len(m.notes) > 0 {
			m.notes[m.cursor].Demote()

			utils.WriteState(m.path, &m.notes)
			m.notes = utils.ReadState(m.path)
		}
	case "+":
		if len(m.notes) > 0 {
			m.notes[m.cursor].IncreasePriority()

			utils.WriteState(m.path, &m.notes)
			m.notes = utils.ReadState(m.path)
		}
	case "-":
		if len(m.notes) > 0 {
			m.notes[m.cursor].DecreasePriority()

			utils.WriteState(m.path, &m.notes)
			m.notes = utils.ReadState(m.path)
		}
	case "x":
		if len(m.notes) > 0 {
			m.notes = append(m.notes[:m.cursor], m.notes[m.cursor+1:]...)

            if m.cursor >= len(m.notes) {
                m.cursor = len(m.notes) - 1 
            }

			utils.WriteState(m.path, &m.notes)
			m.notes = utils.ReadState(m.path)
		}
	case "a":
		m.editing = true
	}

	return m, nil
}
