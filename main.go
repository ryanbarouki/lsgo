package main

// TODO:
// 1. Add more detail to files with command line flags for varying levels of detail
// 2. Ability to delete selected files
// 3. Ability to create new files/dirs
// 4. Ability to move through directories
// 5. Ability to copy/paste selected files

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Styles struct {
	renamingStyle    lipgloss.Style
	highlightedStyle lipgloss.Style
	cursorStyle      lipgloss.Style
	checkedStyle     lipgloss.Style
}

type model struct {
	fnames   []string         // items on the to-do list
	cursor   int              // which to-do list item our cursor is pointing at
	selected map[int]struct{} // which to-do items are selected
	currEdit textinput.Model
	renaming int
	styles   Styles
}

func isHidden(file os.DirEntry) bool {
	return file.Name()[0] == '.'
}

type Opts struct {
	dir string
}

func initialModel(opts Opts) model {
	entries, err := os.ReadDir(opts.dir)
	os.Chdir(opts.dir)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return model{}
	}
	fnames := make([]string, 0, len(entries))
	// fullFnames := make([]string, 0, len(entries))
	for _, file := range entries {
		fnames = append(fnames, file.Name())
		// fullFnames = append(fullFnames, opts.dir+"/"+file.Name())
		// fmt.Println(opts.dir + "/" + file.Name())
	}
	styles := Styles{
		renamingStyle:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("5")),
		highlightedStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#d8b172")).Bold(true),
		checkedStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")),
		cursorStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("#d8b172")),
	}
	ti := textinput.New()
	ti.CharLimit = 156
	ti.Width = 20
	ti.TextStyle = styles.renamingStyle
	ti.Prompt = ""
	return model{
		fnames:   fnames,
		selected: make(map[int]struct{}),
		currEdit: ti,
		renaming: -1,
		styles:   styles,
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.renaming != -1 {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {

			case "enter":
				newName := m.currEdit.Value()
				if newName == "" {
					m.renaming = -1
					return m, nil
				}
				err := os.Rename(m.fnames[m.cursor], newName)
				if err != nil {
					fmt.Println("Rename failed:", err)
					return m, tea.Quit
				}
				m.fnames[m.cursor] = newName
				m.renaming = -1
				m.currEdit.Reset()

			default:
				var cmd tea.Cmd
				m.currEdit, cmd = m.currEdit.Update(msg)
				return m, cmd
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.fnames)-1 {
				m.cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case " ", "enter":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}

		case "r":
			// Rename file or folder
			var cmd tea.Cmd
			m.renaming = m.cursor
			m.currEdit.Placeholder = m.fnames[m.cursor]
			m.currEdit.Focus()
			return m, cmd
		}
	}
	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {
	// The header
	//s := "LSGO\n\n"
	s := "\n"

	// Iterate over our choices
	for i, filename := range m.fnames {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = m.styles.cursorStyle.Render("→")
		}

		// Is this choice selected?
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			checked = m.styles.checkedStyle.Render("\uf42e")
		}

		// Render the row
		if i == m.renaming {
			s += fmt.Sprintf(" %s %s %s\n", cursor, checked, m.currEdit.View())
		} else {
			if i == m.cursor {
				filename = m.styles.highlightedStyle.Render(filename)
			}
			s += fmt.Sprintf(" %s %s %s\n", cursor, checked, filename)
		}
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}

func main() {
	args := os.Args
	opts := Opts{"."}
	if len(args) > 1 {
		opts.dir = args[1]
	}
	p := tea.NewProgram(initialModel(opts))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
