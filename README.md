# GoPyte - Native Go Terminal Emulator

A pure Go implementation of a VT100/VT220/ANSI terminal emulator, originally inspired by the Python [pyte](https://github.com/selectel/pyte) library. GoPyte provides complete terminal emulation with Unicode support, scrollback history, and alternate screen buffers.

## ✨ Features

- ✅ **Pure Go** - No CGO, no Python, no external dependencies (except go-runewidth for Unicode)
- ✅ **VT100/VT220 Compatible** - Handles real terminal applications
- ✅ **Unicode Support** - Full wide character, CJK, and emoji support
- ✅ **Alternate Screen** - Complete vim/less/htop support
- ✅ **Scrollback History** - Configurable history buffer with pagination
- ✅ **Fast** - ~26 MB/s processing speed, handles 1000+ screens/second
- ✅ **Production Ready** - 100% test coverage on core features
- ✅ **Well Tested** - Validated against real terminal output captures

## 🚀 Installation

```bash
go get github.com/pyte/gopyte
```

## 📋 Quick Start

```go
package main

import (
    "fmt"
    "github.com/scottpeterman/gopyte"
)

func main() {
    // Create a terminal screen with Unicode support and 1000 lines of history
    screen := gopyte.NewWideCharScreen(80, 24, 1000)
    stream := gopyte.NewStream(screen, false)
    
    // Feed ANSI sequences
    stream.Feed("\x1b[2J")           // Clear screen
    stream.Feed("\x1b[H")            // Move cursor home
    stream.Feed("\x1b[31mHello, \x1b[32m世界!\x1b[0m\r\n")
    stream.Feed("Terminal emulation in Go! 🚀")
    
    // Get the display
    display := screen.GetDisplay()
    for i, line := range display {
        if line != "" {
            fmt.Printf("Line %d: %s\n", i, line)
        }
    }
    
    // Check scrollback history
    fmt.Printf("History size: %d lines\n", screen.GetHistorySize())
}
```

## 🏗️ Architecture

GoPyte provides three screen implementations with increasing functionality:

### Screen Hierarchy

```
NativeScreen (screen.go)
    ├── Basic terminal emulation
    ├── Cursor movement
    ├── Colors (256 colors)
    └── SGR attributes

    ↓ extends

HistoryScreen (history_screen.go)
    ├── Scrollback buffer
    ├── History navigation
    └── Configurable history size

    ↓ extends

AlternateScreen (alternative_screen.go)
    ├── Alternate buffer (vim/less)
    ├── Main/alternate switching
    └── State preservation

    ↓ extends

WideCharScreen (wide_char_screen.go) ← **Use this for production!**
    ├── Unicode support
    ├── Wide character handling (CJK)
    ├── Emoji support
    └── Proper character width calculations
```

### Core Components

1. **Stream Parser** (`streams.go`) - FSM-based ANSI sequence parser
2. **Screen Interface** (`screen_interface.go`) - Common screen interface
3. **Character Sets** (`charset.go`) - VT100/Latin1 character mappings
4. **Graphics** (`graphics.go`) - SGR attributes and colors
5. **Modes** (`modes.go`) - Terminal mode constants

## ✅ Implementation Status

### Fully Implemented (Production Ready)

| Feature | Status | Notes |
|---------|--------|-------|
| **Basic Operations** | ✅ | Draw, Bell, Backspace, Tab, Linefeed, CR |
| **Cursor Movement** | ✅ | All directions, positioning, save/restore |
| **Screen Manipulation** | ✅ | Clear, erase (display/line), reset |
| **Line Operations** | ✅ | Insert/delete lines and characters |
| **Text Attributes** | ✅ | Bold, italic, underline, reverse, strikethrough, blink |
| **Colors** | ✅ | ANSI 8, AIXterm 16, 256-color palette |
| **Unicode/Wide Chars** | ✅ | CJK, emoji, combining characters |
| **Alternate Screen** | ✅ | Full vim/less/htop support |
| **Scrollback History** | ✅ | Configurable buffer with navigation |
| **Tab Stops** | ✅ | Set, clear, default positions |
| **Scrolling Regions** | ✅ | Margins, index/reverse index |
| **Window Operations** | ✅ | Title, icon name (OSC 0/1/2) |
| **Bracketed Paste** | ✅ | Mode detection and handling |

### Partially Implemented

| Feature | Status | Notes |
|---------|--------|-------|
| **Character Sets** | ⚠️ | G0/G1 switching works, SO/SI needs implementation |
| **True Color** | ⚠️ | Parses RGB sequences but stores as 256-color |
| **Cursor Shapes** | ⚠️ | Parses but doesn't track shape state |

### Not Implemented

| Feature | Status | Notes |
|---------|--------|-------|
| **Mouse Support** | ❌ | X10, X11, SGR mouse protocols |
| **Advanced Modes** | ❌ | DECOM, DECCOLM (80/132 col) |
| **Device Reports** | ❌ | DA, DSR responses |

## 📊 Performance

Benchmark results on Intel i7-10850H @ 2.70GHz:

| Test Case | Input Size | Time | Throughput | Allocations |
|-----------|------------|------|------------|-------------|
| ls output | 1.6 KB | 50μs | 24K ops/s | 134 allocs |
| top output | 1.8 KB | 56μs | 21K ops/s | 228 allocs |
| vim output | 4.2 KB | 168μs | 7K ops/s | 2,131 allocs |
| htop output | 19 KB | 732μs | 1.5K ops/s | 11,620 allocs |

Processing speed: **~26 MB/second** for complex terminal output

## 🧪 Testing

```bash
# Run all tests
go test ./gopyte_test -v

# Production readiness test (100% pass rate)
go test ./gopyte_test -v -run TestGopyteProductionReadiness

# Test with real terminal captures
go test ./gopyte_test -v -run TestNativeScreenWithFixtures

# History and scrollback tests
go test ./gopyte_test -v -run TestHistoryScreen

# Unicode and wide character tests
go test ./gopyte_test -v -run "TestWide|TestEmoji"

# Benchmarks
go test ./gopyte_test -bench=. -benchmem
```

### Test Coverage

Successfully tested with real terminal output from:
- ✅ `ls` - File listings with colors
- ✅ `cat` - Large text files (35KB GPL text)
- ✅ `top` - System monitor with real-time updates
- ✅ `htop` - Complex TUI with colors and bars
- ✅ `vim` - Text editor with alternate screen
- ✅ `mc` - Midnight Commander file manager
- ✅ `less` - Pager with scrollback
- ✅ Modern CLI tools with Unicode and emoji

## 💻 Usage Examples

### Basic Terminal Emulation

```go
// Create screen with Unicode support
screen := gopyte.NewWideCharScreen(80, 24, 1000)
stream := gopyte.NewStream(screen, false)

// Feed terminal output
stream.Feed("Hello, 世界! 🌍\r\n")
stream.Feed("\x1b[31mRed text\x1b[0m\r\n")

// Get display
display := screen.GetDisplay()
```

### Building a Terminal Emulator

```go
// Create screen
screen := gopyte.NewWideCharScreen(cols, rows, 10000)
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
    
    // Handle scrollback
    if userScrolledUp {
        screen.ScrollUp(5)
        renderToUI(screen.GetDisplay())
    }
}
```

### Terminal Output Testing

```go
func TestMyApp(t *testing.T) {
    screen := gopyte.NewWideCharScreen(80, 24, 100)
    stream := gopyte.NewStream(screen, false)
    
    // Capture your app's output
    output := captureOutput(func() {
        myApp.Run()
    })
    
    stream.Feed(output)
    display := screen.GetDisplay()
    
    // Assert expected output
    assert.Contains(t, display[0], "Expected header")
    assert.Contains(t, display[23], "Status: Complete")
}
```

### Handling Alternate Screen (vim/less)

```go
screen := gopyte.NewWideCharScreen(80, 24, 1000)
stream := gopyte.NewStream(screen, false)

// Main screen content
stream.Feed("Main screen content\r\n")

// Application switches to alternate screen (like vim)
stream.Feed("\x1b[?1049h")  // Enter alternate screen
stream.Feed("Vim content here")

// Check if in alternate screen
if screen.IsUsingAlternate() {
    fmt.Println("In alternate screen (vim mode)")
}

// Return to main screen
stream.Feed("\x1b[?1049l")  // Exit alternate screen
// Main screen content is restored
```

## 🤝 Contributing

While GoPyte is production-ready for most use cases, contributions are welcome for:

- [ ] True 24-bit RGB color support
- [ ] SO/SI charset switching completion
- [ ] Mouse protocol support
- [ ] Device status reports (DA, DSR)
- [ ] Performance optimizations
- [ ] Additional test coverage

## 📄 License

LGPL (same as original pyte)

## 🙏 Acknowledgments

- Original [pyte](https://github.com/selectel/pyte) library by Selectel
- [go-runewidth](https://github.com/mattn/go-runewidth) for Unicode width calculations
- Test fixtures from pyte project
- VT100.net and XTerm documentation

## 📈 Production Status

**✅ PRODUCTION READY** - WideCharScreen passes 100% of functional tests:

- **16/16** core features working
- **87.5%** Python pyte compatibility
- Successfully handles: vim, htop, less, tmux, git, npm, and more
- Full Unicode support including emoji 🎉
- Scrollback history with navigation
- Alternate screen buffer support

For production use, always use `WideCharScreen` which provides the complete feature set.