# lsgo
Just a toy project to get familiar with GO.

Pronounced "les gooooo!"

## Features
- Keyboard navigation through files and directories
- Show/hide hidden files
- Toggle file permission visibility
- Create new files
- Rename files and directories
- Delete files and directories
- Traverse into subdirectories and back up

## Planned Features

- [ ] Ability to copy/paste selected files
- [ ] Search/filter the file list

## Installation

Ensure you have Go installed (version 1.18+ recommended). Clone the repository and build it:

```bash
git clone https://github.com/yourusername/lsgo.git
cd lsgo
go build -o lsgo
```

## Usage
```bash
./lsgo [path] [flags]
```

### Flags
  `-l`: Show file permissions

  `-a`: Show hidden files

  `-la`: Show hidden files and permissions

## Key Bindings

| Key           | Action                              |
|---------------|-------------------------------------|
| `q` / `Ctrl+C`| Quit the program                    |
| `j` / `↓`     | Move cursor down                    |
| `k` / `↑`     | Move cursor up                      |
| `Space`       | Toggle selection for current item   |
| `Enter`       | Enter directory                     |
| `Backspace`   | Go to parent directory              |
| `r`           | Rename selected file or folder      |
| `d`           | Delete file (prompts for confirmation) |
| `a`           | Add new file                        |

