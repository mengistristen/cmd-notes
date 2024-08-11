package tui

import (
	"cmd-notes/note"
	"cmd-notes/utils"
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Available TUI states
const (
	LIST = iota
	ADD
)

type Model struct {
	keys      keymap
	help      help.Model
	textInput textinput.Model
	notes     []note.Note
	path      string
	cursor    int
	state     int
}

type keymap struct {
	Up               key.Binding
	Down             key.Binding
	Help             key.Binding
	Add              key.Binding
	Remove           key.Binding
	IncreasePriority key.Binding
	DecreasePriority key.Binding
	IncreaseStatus   key.Binding
	DecreaseStatus   key.Binding
	Quit             key.Binding
}

func (k keymap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k keymap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},
		{k.IncreasePriority, k.DecreasePriority},
		{k.IncreaseStatus, k.DecreaseStatus},
		{k.Add, k.Remove},
		{k.Help, k.Quit},
	}
}

var keys = keymap{
	Up:               key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "move up")),
	Down:             key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "move down")),
	Help:             key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "toggle help")),
	Add:              key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add item")),
	Remove:           key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "remove item")),
	IncreasePriority: key.NewBinding(key.WithKeys("+"), key.WithHelp("+", "increase priority")),
	DecreasePriority: key.NewBinding(key.WithKeys("-"), key.WithHelp("-", "decrease priority")),
	IncreaseStatus:   key.NewBinding(key.WithKeys(">"), key.WithHelp(">", "increase status")),
	DecreaseStatus:   key.NewBinding(key.WithKeys("<"), key.WithHelp("<", "decrease status")),
	Quit:             key.NewBinding(key.WithKeys("q", "ctrl-c"), key.WithHelp("q", "quit")),
}

func InitModel(path string, notes []note.Note) Model {
	ti := textinput.New()
	ti.Placeholder = "Task"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return Model{
		keys:      keys,
		help:      help.New(),
		textInput: ti,
		notes:     notes,
		path:      path,
		cursor:    0,
		state:     LIST,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		default:
			switch m.state {
			case LIST:
				return updateList(m, msg)
			case ADD:
				return updateAdd(m, msg)
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	var result string

	switch m.state {
	case LIST:
		for i, note := range m.notes {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}

			result += fmt.Sprintf("%s %s - %s\n", cursor, note.FormatPriority(), note.FormatContents())
		}

		result += "\n" + m.help.View(m.keys)
	case ADD:
		result = fmt.Sprintf("Add item:\n\n%s\n\n%s", m.textInput.View(), "(esc to quit)")
	}

	return result
}

func updateAdd(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.Type {
	case tea.KeyEsc:
		m.state = LIST
		m.textInput.SetValue("")
	case tea.KeyEnter:
		m.state = LIST
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

func updateList(m Model, msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.notes)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keys.IncreaseStatus):
			if len(m.notes) > 0 {
				m.notes[m.cursor].Promote()

				utils.WriteState(m.path, &m.notes)
				m.notes = utils.ReadState(m.path)
			}
		case key.Matches(msg, m.keys.DecreaseStatus):
			if len(m.notes) > 0 {
				m.notes[m.cursor].Demote()

				utils.WriteState(m.path, &m.notes)
				m.notes = utils.ReadState(m.path)
			}
		case key.Matches(msg, m.keys.IncreasePriority):
			if len(m.notes) > 0 {
				m.notes[m.cursor].IncreasePriority()

				utils.WriteState(m.path, &m.notes)
				m.notes = utils.ReadState(m.path)
			}
		case key.Matches(msg, m.keys.DecreasePriority):
			if len(m.notes) > 0 {
				m.notes[m.cursor].DecreasePriority()

				utils.WriteState(m.path, &m.notes)
				m.notes = utils.ReadState(m.path)
			}
		case key.Matches(msg, m.keys.Remove):
			if len(m.notes) > 0 {
				m.notes = append(m.notes[:m.cursor], m.notes[m.cursor+1:]...)

				if m.cursor != 0 && m.cursor >= len(m.notes) {
					m.cursor = len(m.notes) - 1
				}

				utils.WriteState(m.path, &m.notes)
				m.notes = utils.ReadState(m.path)
			}
		case key.Matches(msg, m.keys.Add):
			m.state = ADD
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		}
	}

	return m, nil
}
