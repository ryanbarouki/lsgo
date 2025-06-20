package main

// TODO:
// Features:
// 1. Ability to copy/paste selected files
// 2. Search/Filter list of files

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"lsgo/utils"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Mode int

const (
	NormalMode Mode = iota
	RenameMode
	DeleteMode
	AddMode
)

type model struct {
	fileInfo     []os.FileInfo // items on the to-do list
	displayNames []string
	cursor       int // which to-do list item our cursor is pointing at
	prevCursor   int
	selected     map[int]struct{} // which to-do items are selected
	currEdit     textinput.Model
	renaming     int
	styles       Styles
	opts         Opts
	mode         Mode
}

func isHidden(file os.DirEntry) bool {
	return file.Name()[0] == '.'
}

type Opts struct {
	dir             string
	showPerms       bool
	showHiddenFiles bool
}

func initialModel(opts Opts) *model {
	fnames, fileInfo := loadFileInfo(opts)
	styles := InitStyles()
	ti := textinput.New()
	ti.CharLimit = 156
	ti.Width = 20
	ti.TextStyle = styles.renamingStyle
	ti.Prompt = ""

	return &model{
		fileInfo:     fileInfo,
		displayNames: fnames,
		selected:     make(map[int]struct{}),
		currEdit:     ti,
		renaming:     -1,
		styles:       styles,
		opts:         opts,
		mode:         NormalMode,
	}
}

func (m *model) resetModel(opts Opts) {

	fnames, fileInfo := loadFileInfo(opts)

	m.fileInfo = fileInfo
	m.displayNames = fnames
	m.renaming = -1
	m.mode = NormalMode
	m.opts = opts
	m.cursor = 0
	m.currEdit.Reset()
	m.selected = make(map[int]struct{})
}

func loadFileInfo(opts Opts) ([]string, []os.FileInfo) {

	entries, err := os.ReadDir(opts.dir)
	if err != nil {
		fmt.Println("Error reading directory:", err)
		os.Exit(1)
	}

	fnames := make([]string, 0, len(entries))
	fileInfo := make([]os.FileInfo, 0, len(entries))
	for _, file := range entries {
		info, err := file.Info()
		if err != nil {
			continue
		}
		if !opts.showHiddenFiles && isHidden(file) {
			continue
		}
		fnames = append(fnames, file.Name())
		fileInfo = append(fileInfo, info)
	}
	return fnames, fileInfo
}

func (m *model) absPath(filename string) (string, error) {
	return filepath.Abs(filepath.Join(m.opts.dir, filename))
}

func fatal(err error) tea.Cmd {
	fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
	return tea.Quit
}

func (m *model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m *model) updateNormal(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
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
		m.mode = RenameMode
		m.currEdit.Placeholder = m.displayNames[m.cursor]
		m.currEdit.Focus()
		return m, nil

	case "enter":
		if m.fileInfo[m.cursor].IsDir() {
			absPath, err := m.absPath(m.displayNames[m.cursor])
			if err != nil {
				return m, fatal(err)
			}
			newOpts := m.opts
			newOpts.dir = absPath
			prevCursor := m.cursor
			m.resetModel(newOpts)
			m.prevCursor = prevCursor
			return m, nil
		}

	case "backspace":
		absPath, err := filepath.Abs(filepath.Dir(m.opts.dir))
		if err != nil {
			return m, fatal(err)
		}

		newOpts := m.opts
		newOpts.dir = absPath
		m.resetModel(newOpts)
		m.cursor = m.prevCursor
		return m, nil

	case "d":
		// Delete file
		m.mode = DeleteMode
		return m, nil

	case "a":
		m.mode = AddMode
		m.currEdit.Placeholder = "Create a new file"
		m.currEdit.Focus()
		return m, nil
	}
	return m, nil
}

func (m *model) updateAdd(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "enter":
		newFile := m.currEdit.Value()
		if newFile == "" {
			m.mode = NormalMode
			return m, nil
		}

		newFullPath, err := m.absPath(newFile)
		if err != nil {
			return m, fatal(err)
		}
		//os.Mkdir()
		if _, err := os.Create(newFullPath); err != nil {
			// TODO: Make this a recoverable error
			return m, fatal(err)
		}

		m.displayNames, m.fileInfo = loadFileInfo(m.opts)
		m.mode = NormalMode
		m.currEdit.Reset()
		m.cursor = utils.IndexOf(m.displayNames, newFile)
		return m, nil

	default:
		var cmd tea.Cmd
		m.currEdit, cmd = m.currEdit.Update(msg)
		return m, cmd
	}
}

func (m *model) updateRename(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "enter":
		newName := m.currEdit.Value()
		if newName == "" {
			m.renaming = -1
			m.mode = NormalMode
			return m, nil
		}
		oldFullPath, err := m.absPath(m.displayNames[m.cursor])
		if err != nil {
			return m, fatal(err)
		}

		newFullPath, err := m.absPath(newName)
		if err != nil {
			return m, fatal(err)
		}

		if err := os.Rename(oldFullPath, newFullPath); err != nil {
			// TODO: Make this a recoverable error
			return m, fatal(err)
		}

		m.displayNames[m.cursor] = newName
		// NOTE: New name isn't updated in fileinfo unless one moves directory
		// consider changing this
		// currently displayNames may differ
		m.renaming = -1
		m.mode = NormalMode
		m.currEdit.Reset()

	default:
		var cmd tea.Cmd
		m.currEdit, cmd = m.currEdit.Update(msg)
		return m, cmd

	}
	return m, nil
}

func (m *model) boundedCursor(cursor int) int {
	if cursor >= len(m.displayNames) {
		return len(m.displayNames) - 1
	}
	return cursor
}

func (m *model) updateDelete(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch key.String() {
	case "y":
		fileToDelete, err := m.absPath(m.displayNames[m.cursor])
		if err != nil {
			return m, fatal(err)
		}

		if err = os.Remove(fileToDelete); err != nil {
			return m, fatal(err)
		}

		savedCursor := m.cursor
		m.resetModel(m.opts)
		m.cursor = m.boundedCursor(savedCursor)
		return m, nil
	case "n":
		m.mode = NormalMode
		return m, nil

	}
	return m, nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case NormalMode:
		return m.updateNormal(msg)
	case RenameMode:
		return m.updateRename(msg)
	case DeleteMode:
		return m.updateDelete(msg)
	case AddMode:
		return m.updateAdd(msg)
	}
	return m, nil
}

func (m *model) rowBuilder(i int, s *strings.Builder) {

	// Is the cursor pointing at this choice?
	cursor := " " // no cursor
	if m.cursor == i {
		cursor = m.styles.cursorStyle.Render("→")
	}

	s.WriteString(fmt.Sprintf(" %s", cursor))

	if m.opts.showPerms {
		permissions := m.fileInfo[i].Mode().Perm().String()
		if m.cursor == i {
			permissions = m.styles.highlightedStyle.Render(permissions)
		}
		s.WriteString(fmt.Sprintf(" %s", permissions))
	}

	// Is this choice selected?
	checked := " " // not selected
	if _, ok := m.selected[i]; ok {
		checked = m.styles.checkedStyle.Render("\uf42e")
	}

	icon := utils.IconFor(m.fileInfo[i])

	s.WriteString(fmt.Sprintf(" %s %s", checked, icon))

	filename := m.displayNames[i]
	if m.fileInfo[i].IsDir() {
		filename += "/"
	}

	if i != m.cursor {
		s.WriteString(fmt.Sprintf(" %s", filename))
		return
	}

	switch m.mode {
	case NormalMode:
		filename = m.styles.highlightedStyle.Render(filename)
		s.WriteString(fmt.Sprintf(" %s", filename))
	case RenameMode:
		s.WriteString(fmt.Sprintf(" %s", m.currEdit.View()))
	case DeleteMode:
		message := m.styles.confirmDelStyle.Render("Confirm delete (y/n)")
		s.WriteString(fmt.Sprintf(" %s", message))
	}
}

func (m *model) addFileView(s *strings.Builder) {
	cursor := m.styles.cursorStyle.Render("→")
	s.WriteString(fmt.Sprintf(" %s", cursor))
	if m.opts.showPerms {
		var mode os.FileMode = 0644
		permissions := mode.Perm().String()
		permissions = m.styles.highlightedStyle.Render(permissions)
		s.WriteString(fmt.Sprintf(" %s", permissions))
	}
	checked := " " // not selected
	icon := "📄"
	s.WriteString(fmt.Sprintf(" %s %s %s", checked, icon, m.currEdit.View()))
}

func (m *model) View() string {
	var s strings.Builder
	s.WriteString("\n")

	if m.mode == AddMode {
		m.cursor = len(m.displayNames)
	}

	for i, _ := range m.displayNames {
		m.rowBuilder(i, &s)
		s.WriteString("\n")
	}

	if m.mode == AddMode {
		m.addFileView(&s)
		s.WriteString("\n")
	}

	styled := m.styles.keyStyle.Render

	instructions := fmt.Sprintf(
		"%s Navigate   %s Select   %s Open   %s Rename   %s Delete   %s Add   %s Quit\n",
		styled("[j/k]"),
		styled("[Space]"),
		styled("[Enter]"),
		styled("[r]"),
		styled("[d]"),
		styled("[a]"),
		styled("[q]"),
	)

	s.WriteString("\n" + instructions)
	return s.String()
}

func main() {
	showPerms := flag.Bool("l", false, "Show file permissions")
	showHiddenFiles := flag.Bool("a", false, "Show hidden files")
	showHiddenAndPerms := flag.Bool("la", false, "Show hidden files and all permissions")
	flag.Parse()

	if *showHiddenAndPerms {
		*showPerms = true
		*showHiddenFiles = true
	}

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
		dir:             absPath,
		showPerms:       *showPerms,
		showHiddenFiles: *showHiddenFiles,
	}

	p := tea.NewProgram(initialModel(opts))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
