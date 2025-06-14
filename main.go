package main

// TODO:
// 1. Add more detail to files with command line flags for varying levels of detail
// 2. Add delete confirmation in a nice text-box
// 3. Ability to create new files/dirs
// 4. Ability to copy/paste selected files
// 5. Search/Filter list of files

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"lsgo/utils"
)

type Styles struct {
	renamingStyle    lipgloss.Style
	highlightedStyle lipgloss.Style
	cursorStyle      lipgloss.Style
	checkedStyle     lipgloss.Style
}

type model struct {
	fileInfo     []os.FileInfo // items on the to-do list
	displayNames []string
	cursor       int              // which to-do list item our cursor is pointing at
	selected     map[int]struct{} // which to-do items are selected
	currEdit     textinput.Model
	renaming     int
	styles       Styles
	opts         Opts
}

func isHidden(file os.DirEntry) bool {
	return file.Name()[0] == '.'
}

type Opts struct {
	dir       string
	showPerms bool
}

func initialModel(opts Opts) model {
	entries, err := os.ReadDir(opts.dir)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		return model{}
	}
	fnames := make([]string, 0, len(entries))
	fileInfo := make([]os.FileInfo, 0, len(entries))
	for _, file := range entries {
		info, err := file.Info()
		if err != nil {
			continue
		}
		fnames = append(fnames, file.Name())
		fileInfo = append(fileInfo, info)
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
		fileInfo:     fileInfo,
		displayNames: fnames,
		selected:     make(map[int]struct{}),
		currEdit:     ti,
		renaming:     -1,
		styles:       styles,
		opts:         opts,
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
				oldFullPath, err := filepath.Abs(filepath.Join(m.opts.dir, m.displayNames[m.cursor]))

				if err != nil {
					fmt.Println("Invalid path:", err)
					os.Exit(1)
				}

				newFullPath, err := filepath.Abs(filepath.Join(m.opts.dir, newName))

				if err != nil {
					fmt.Println("Invalid path:", err)
					os.Exit(1)
				}

				err = os.Rename(oldFullPath, newFullPath)
				if err != nil {
					fmt.Println("Rename failed:", err)
					return m, tea.Quit
				}
				m.displayNames[m.cursor] = newName
				// NOTE: New name isn't updated in fileinfo unless one moves directory
				// consider changing this
				// currently displayNames may differ
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
			if m.cursor < len(m.fileInfo)-1 {
				m.cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}

		case "r":
			// Rename file or folder
			m.renaming = m.cursor
			m.currEdit.Placeholder = m.displayNames[m.cursor]
			m.currEdit.Focus()
			return m, nil

		case "enter":
			if m.fileInfo[m.cursor].IsDir() {
				newPath := filepath.Join(m.opts.dir, m.displayNames[m.cursor])
				absPath, err := filepath.Abs(newPath)
				if err != nil {
					fmt.Println("Error resolving path:", err)
					return m, nil
				}
				newOpts := m.opts
				newOpts.dir = absPath
				return initialModel(newOpts), nil
			}

		case "backspace":
			parent := filepath.Dir(m.opts.dir)
			absPath, err := filepath.Abs(parent)
			if err != nil {
				fmt.Println("Error resolving parent path:", err)
				return m, nil
			}
			newOpts := m.opts
			newOpts.dir = absPath
			return initialModel(newOpts), nil

		case "d":
			// Delete file
			fileToDelete, err := filepath.Abs(m.displayNames[m.cursor])
			if err != nil {
				fmt.Println("Error resolving parent path:", err)
				return m, nil
			}
			err = os.Remove(fileToDelete)
			if err != nil {
				fmt.Println("Error deleting:", err)
				return m, tea.Quit
			}
			newModel := initialModel(m.opts)
			newModel.cursor = m.cursor
			return newModel, nil
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
	for i, filename := range m.displayNames {

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

		icon := utils.IconFor(m.fileInfo[i])

		permissions := ""
		if m.opts.showPerms {
			permissions = m.fileInfo[i].Mode().Perm().String()
		}

		// Render the row
		if i == m.renaming {
			s += fmt.Sprintf(" %s %s %s %s %s\n", cursor, permissions, checked, icon, m.currEdit.View())
		} else {
			if m.fileInfo[i].IsDir() {
				filename += "/"
			}
			if i == m.cursor {
				filename = m.styles.highlightedStyle.Render(filename)
				permissions = m.styles.highlightedStyle.Render(permissions)
			}
			s += fmt.Sprintf(" %s %s %s %s %s\n", cursor, permissions, checked, icon, filename)
		}
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}

func main() {
	showPerms := flag.Bool("a", false, "Show file permissions")
	flag.Parse()

	args := flag.Args()

	startPath := "."
	if len(args) > 0 {
		startPath = args[0]
	}

	absPath, err := filepath.Abs(startPath)
	if err != nil {
		fmt.Println("Invalid path:", err)
		os.Exit(1)
	}

	opts := Opts{
		dir:       absPath,
		showPerms: *showPerms,
	}

	p := tea.NewProgram(initialModel(opts))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
