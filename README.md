# GoPyte - Native Go Terminal Emulator

A pure Go implementation of a VT100/VT220/ANSI terminal emulator, originally inspired by the Python [pyte](https://github.com/selectel/pyte) library. GoPyte provides complete terminal emulation with no external dependencies.

## Features

- ✅ **Pure Go** - No CGO, no Python, no external dependencies
- ✅ **VT100/VT220 Compatible** - Handles real terminal applications
- ✅ **Fast** - ~26 MB/s processing speed, handles 1000+ screens/second
- ✅ **Complete** - Successfully parses htop, vim, mc, and other complex terminal applications
- ✅ **Well Tested** - Validated against real terminal output captures

## Installation

```bash
go get github.com/pyte/gopyte
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/pyte/gopyte/gopyte"
)

func main() {
    // Create a terminal screen (80x24 is standard)
    screen := gopyte.NewNativeScreen(80, 24)
    stream := gopyte.NewStream(screen, false)
    
    // Feed ANSI sequences
    stream.Feed("\x1b[2J")           // Clear screen
    stream.Feed("\x1b[H")            // Move cursor home
    stream.Feed("\x1b[31mHello, \x1b[32mWorld!\x1b[0m\r\n")
    stream.Feed("Terminal emulation in Go!")
    
    // Get the display
    display := screen.GetDisplay()
    for i, line := range display {
        if line != "" {
            fmt.Printf("Line %d: %s\n", i, line)
        }
    }
}
```

## Architecture

GoPyte consists of two main components:

### 1. Stream Parser (`streams.go`)
A finite state machine that parses ANSI escape sequences:
- Control characters (CR, LF, BS, TAB, etc.)
- ESC sequences (cursor save/restore, index, etc.)
- CSI sequences (cursor movement, colors, erasing)
- OSC sequences (window title, icon name)
- Character set selection (G0/G1)

### 2. Native Screen (`screen.go`)
A complete terminal screen implementation:
- 2D buffer for character storage
- Full cursor management
- Text attributes and colors (SGR)
- Scrolling regions (margins)
- Tab stops
- Line operations (insert/delete)

## Implementation Status

### ✅ Fully Implemented

- **Basic Operations**: Draw, Bell, Backspace, Tab, Linefeed, Carriage Return
- **Cursor Movement**: All directions, positioning, save/restore
- **Screen Manipulation**: Clear, erase (display/line), reset
- **Line Operations**: Insert/delete lines and characters
- **Text Attributes**: Bold, italic, underline, reverse, strikethrough, blink
- **Colors**: ANSI 8-color, AIXterm 16-color, 256-color support
- **Scrolling**: Regions with margins, index/reverse index
- **Modes**: Auto-wrap (DECAWM), newline mode (LNM)
- **Tab Stops**: Set, clear, default positions

### 🚧 Partially Implemented

- **Character Sets**: Basic G0/G1 switching (needs full translation tables)
- **OSC Sequences**: Only window title and icon name (OSC 0/1/2)

### ❌ Not Yet Implemented

- **Wide Characters**: Asian character support (needs wcwidth)
- **Alternative Screen Buffer**: For full vim/less support
- **Mouse Support**: X10, X11, SGR mouse protocols
- **Advanced Modes**: DECOM, DECCOLM, DECSCNM
- **Scrollback Buffer**: History beyond current screen
- **24-bit True Color**: RGB color support

## Performance

Benchmark results on Intel i7-10850H @ 2.70GHz:

| Test Case | Input Size | Time | Throughput | Allocations |
|-----------|------------|------|------------|-------------|
| ls output | 1.6 KB | 50μs | 24K ops/s | 134 allocs |
| top output | 1.8 KB | 56μs | 21K ops/s | 228 allocs |
| vim output | 4.2 KB | 168μs | 7K ops/s | 2,131 allocs |
| htop output | 19 KB | 732μs | 1.5K ops/s | 11,620 allocs |

Processing speed: **~26 MB/second** for complex terminal output

## Testing

```bash
# Run all tests
go test ./gopyte_test -v

# Run native screen tests only
go test ./gopyte_test -v -run TestNativeScreen

# Run with real terminal captures
go test ./gopyte_test -v -run TestNativeScreenWithFixtures

# Benchmarks
go test ./gopyte_test -bench=BenchmarkNativeScreen -benchmem
```

### Test Coverage

Successfully tested with real terminal output from:
- ✅ `ls` - File listings with colors
- ✅ `cat` - Large text files (35KB GPL text)
- ✅ `top` - System monitor with real-time updates
- ✅ `htop` - Complex TUI with colors and bars
- ✅ `vim` - Text editor with positioning
- ✅ `mc` - Midnight Commander file manager

## Project Structure

```
gopyte/
├── charset.go           # Character set mappings (VT100, Latin1, IBM PC)
├── control.go          # Control character constants
├── escape.go           # Escape sequence constants
├── graphics.go         # SGR attributes and color mappings
├── modes.go            # Terminal mode constants
├── streams.go          # ANSI escape sequence parser (FSM)
├── screen_interface.go # Screen interface definition
├── screen.go           # Native Go screen implementation
├── mock_screen.go      # Mock screen for testing
└── python_screen.go    # Python bridge (for validation only)

gopyte_test/
├── native_screen_test.go    # Native implementation tests
├── native_fixtures_test.go  # Tests with real terminal output
├── stream_test.go           # Parser tests
└── testdata/               # Real terminal captures
```

## Comparison with Python pyte

| Feature | Python pyte | GoPyte |
|---------|------------|---------|
| Language | Python | Pure Go |
| Dependencies | wcwidth | None |
| Performance | Interpreted | ~10-50x faster |
| Memory Model | Sparse dict | Dense arrays |
| Wide Char Support | ✅ Full | ❌ Not yet |
| Scrollback | ✅ HistoryScreen | ❌ Not yet |
| API Compatibility | Original | Similar |

## Usage Examples

### Parse Terminal Output

```go
screen := gopyte.NewNativeScreen(80, 24)
stream := gopyte.NewStream(screen, false)

// Read terminal output from file or process
data, _ := os.ReadFile("terminal_output.txt")
stream.Feed(string(data))

// Get the final screen state
display := screen.GetDisplay()
```

### Build a Terminal Emulator

```go
// Create screen
screen := gopyte.NewNativeScreen(cols, rows)
stream := gopyte.NewStream(screen, false)

// In your PTY read loop
for {
    buf := make([]byte, 4096)
    n, err := pty.Read(buf)
    if err != nil {
        break
    }
    
    // Parse the output
    stream.Feed(string(buf[:n]))
    
    // Render the screen
    display := screen.GetDisplay()
    renderToUI(display)
}
```

### Terminal Testing

```go
// Test that your CLI app produces expected output
screen := gopyte.NewNativeScreen(80, 24)
stream := gopyte.NewStream(screen, false)

output := runCommand("myapp --help")
stream.Feed(output)

display := screen.GetDisplay()
assert.Contains(t, display[0], "Usage:")
```

## Contributing

Contributions welcome! Areas that need work:
- Wide character support (integrate `github.com/mattn/go-runewidth`)
- Alternative screen buffer
- Mouse protocol support
- Scrollback/history buffer
- Performance optimizations

## License

[Same as original pyte - LGPL]

## Acknowledgments

- Original [pyte](https://github.com/selectel/pyte) library by Selectel for the inspiration
- Test fixtures from pyte project
- VT100.net and XTerm documentation for specifications