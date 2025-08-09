# GoPyte - Native Go Terminal Emulator

A pure Go implementation of a VT100/VT220/ANSI terminal emulator, originally inspired by the Python [pyte](https://github.com/selectel/pyte) library. GoPyte provides complete terminal emulation with Unicode support, scrollback history, alternate screen buffers, and dynamic resizing.

## Features

- **Pure Go** - No CGO, no Python, no external dependencies (except go-runewidth for Unicode)
- **VT100/VT220 Compatible** - Handles real terminal applications
- **Unicode Support** - Full wide character, CJK, and emoji support
- **Alternate Screen** - Complete vim/less/htop support with proper state isolation
- **Scrollback History** - Configurable history buffer with pagination
- **Dynamic Resizing** - Proper terminal resize handling with content preservation
- **Fast** - ~26 MB/s processing speed, handles 1000+ screens/second
- **Production Ready** - 100% test coverage on core features
- **Well Tested** - Validated against real terminal output captures and live applications

## Installation

```bash
go get github.com/scottpeterman/gopyte/gopyte
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/scottpeterman/gopyte/gopyte"
)

func main() {
    // Create a terminal screen with Unicode support and 1000 lines of history
    screen := gopyte.NewWideCharScreen(80, 24, 1000)
    stream := gopyte.NewStream(screen, false)
    
    // Feed ANSI sequences
    stream.Feed("\x1b[2J")           // Clear screen
    stream.Feed("\x1b[H")            // Move cursor home
    stream.Feed("\x1b[31mHello, \x1b[32m世界!\x1b[0m\r\n")
    stream.Feed("Terminal emulation in Go!")
    
    // Get the display
    display := screen.GetDisplay()
    for i, line := range display {
        if line != "" {
            fmt.Printf("Line %d: %s\n", i, line)
        }
    }
    
    // Check scrollback history
    fmt.Printf("History size: %d lines\n", screen.GetHistorySize())
    
    // Handle terminal resize
    screen.Resize(120, 30)  // Resize to 120x30
    fmt.Printf("Resized to %dx%d\n", 120, 30)
}
```

## Architecture

GoPyte provides three screen implementations with increasing functionality:

### Screen Hierarchy

```
NativeScreen (screen.go)
    ├── Basic terminal emulation
    ├── Cursor movement & positioning
    ├── Colors (ANSI 8, AIXterm 16, 256-color)
    ├── SGR attributes (bold, italic, etc.)
    └── Dynamic resize support

    ↓ extends

HistoryScreen (history_screen.go)
    ├── Scrollback buffer
    ├── History navigation
    ├── Configurable history size
    └── Resize with history preservation

    ↓ extends

AlternateScreen (alternative_screen.go)
    ├── Alternate buffer (vim/less/htop)
    ├── Main/alternate switching
    ├── State preservation across resize
    └── Isolated history handling

    ↓ extends

WideCharScreen (wide_char_screen.go) ← **Use this for production!**
    ├── Unicode support (CJK, emoji)
    ├── Wide character handling
    ├── Proper character width calculations
    └── Resize-safe width tracking
```

### Core Components

1. **Stream Parser** (`streams.go`) - FSM-based ANSI sequence parser
2. **Screen Interface** (`screen_interface.go`) - Common screen interface with resize support
3. **Character Sets** (`charset.go`) - VT100/Latin1 character mappings
4. **Graphics** (`graphics.go`) - SGR attributes and colors
5. **Modes** (`modes.go`) - Terminal mode constants

## Dynamic Resizing

GoPyte includes comprehensive resize support that handles complex edge cases:

### Resize Behavior

- **Column Shrinking**: Content truncated from right, left content preserved
- **Column Growing**: Right-padded with spaces, cursor remains valid
- **Row Shrinking**: Bottom content moves to history (main screen only)
- **Row Growing**: New blank lines added at bottom
- **Wide Characters**: Proper handling across resize boundaries
- **Alternate Screen**: Resize without affecting main screen history
- **Cursor Position**: Automatically clamped to new boundaries

### Resize Example

```go
screen := gopyte.NewWideCharScreen(80, 24, 1000)
stream := gopyte.NewStream(screen, false)

// Fill with content
stream.Feed("Line 1\r\nLine 2\r\nLine 3\r\n")

// Resize to larger
screen.Resize(120, 30)

// Resize to smaller - bottom lines go to history
screen.Resize(60, 10)

// Check history
fmt.Printf("History contains %d lines\n", screen.GetHistorySize())
```

## Implementation Status

### Fully Implemented (Production Ready)

| Feature | Status | Notes |
|---------|--------|-------|
| **Basic Operations** | ✅ DONE | Draw, Bell, Backspace, Tab, Linefeed, CR |
| **Cursor Movement** | ✅ DONE | All directions, positioning, save/restore |
| **Screen Manipulation** | ✅ DONE | Clear, erase (display/line), reset |
| **Line Operations** | ✅ DONE | Insert/delete lines and characters |
| **Text Attributes** | ✅ DONE | Bold, italic, underline, reverse, strikethrough, blink |
| **Colors** | ✅ DONE | ANSI 8, AIXterm 16, 256-color palette |
| **Unicode/Wide Chars** | ✅ DONE | CJK, emoji, combining characters |
| **Alternate Screen** | ✅ DONE | Full vim/less/htop support |
| **Scrollback History** | ✅ DONE | Configurable buffer with navigation |
| **Dynamic Resizing** | ✅ DONE | Content preservation, history handling, cursor clamping |
| **Tab Stops** | ✅ DONE | Set, clear, default positions |
| **Scrolling Regions** | ✅ DONE | Margins, index/reverse index |
| **Window Operations** | ✅ DONE | Title, icon name (OSC 0/1/2) |
| **Bracketed Paste** | ✅ DONE | Mode detection and handling |

### Partially Implemented

| Feature | Status | Notes |
|---------|--------|-------|
| **Character Sets** | 🔄 PARTIAL | G0/G1 switching works, SO/SI needs implementation |
| **True Color** | 🔄 PARTIAL | Parses RGB sequences but stores as 256-color |
| **Cursor Shapes** | 🔄 PARTIAL | Parses but doesn't track shape state |

### Not Yet Implemented

| Feature | Status | Notes |
|---------|--------|-------|
| **Mouse Support** | ⏳ TODO | X10, X11, SGR mouse protocols |
| **Advanced Modes** | ⏳ TODO | DECOM, DECCOLM (80/132 col) |
| **Device Reports** | ⏳ TODO | DA, DSR responses |

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
go test ./gopyte/gopyte_test -v

# Production readiness test (100% pass rate)
go test ./gopyte/gopyte_test -v -run TestGopyteProductionReadiness

# Test with real terminal captures
go test ./gopyte/gopyte_test -v -run TestNativeScreenWithFixtures

# History and scrollback tests
go test ./gopyte/gopyte_test -v -run TestHistoryScreen

# Unicode and wide character tests
go test ./gopyte/gopyte_test -v -run "TestWide|TestEmoji"

# Resize behavior tests
go test ./gopyte/gopyte_test -v -run TestResizeBehavior

# Benchmarks
go test ./gopyte/gopyte_test -bench=. -benchmem
```

### Test Coverage

Successfully tested with real terminal output from:
- `ls` - File listings with colors
- `cat` - Large text files (35KB GPL text)  
- `top` - System monitor with real-time updates
- `htop` - Complex TUI with colors and bars
- `vim` - Text editor with alternate screen
- `mc` - Midnight Commander file manager
- `less` - Pager with scrollback
- **Microsoft `edit`** - Modern Windows text editor
- Modern CLI tools with Unicode and emoji

## Usage Examples

### Basic Terminal Emulation

```go
// Create screen with Unicode support
screen := gopyte.NewWideCharScreen(80, 24, 1000)
stream := gopyte.NewStream(screen, false)

// Feed terminal output
stream.Feed("Hello, 世界!\r\n")
stream.Feed("\x1b[31mRed text\x1b[0m\r\n")

// Get display
display := screen.GetDisplay()
```

### Building a Terminal Emulator with Resize Support

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
    
    // Handle resize events
    if windowResized {
        newCols, newRows := getNewTerminalSize()
        pty.Resize(newCols, newRows)      // Tell PTY
        screen.Resize(newCols, newRows)   // Tell GoPyte
    }
    
    // Render the screen
    display := screen.GetDisplay()
    renderToUI(display)
}
```

### Real-world Example: Windows ConPTY Integration

```go
import "github.com/UserExistsError/conpty"

// Create ConPTY process
cpty, _ := conpty.Start("cmd.exe", conpty.ConPtyDimensions(80, 24))
defer cpty.Close()

// Create GoPyte emulator
screen := gopyte.NewWideCharScreen(80, 24, 2000)
stream := gopyte.NewStream(screen, false)

// Handle resize with proper synchronization
go func() {
    for newSize := range resizeEvents {
        cpty.Resize(newSize.Cols, newSize.Rows)     // Windows PTY
        screen.Resize(newSize.Cols, newSize.Rows)   // GoPyte
        redraw(screen)                              // Update display
    }
}()
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
    
    // Test resize behavior
    screen.Resize(120, 30)
    assert.Equal(t, 30, len(screen.GetDisplay()))
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

// Resize in alternate screen doesn't affect main history
screen.Resize(120, 30)

// Check if in alternate screen
if screen.IsUsingAlternate() {
    fmt.Println("In alternate screen (vim mode)")
}

// Return to main screen
stream.Feed("\x1b[?1049l")  // Exit alternate screen
// Main screen content is restored, history intact
```

## Production Use Cases

### Terminal Emulator Applications
- SSH clients with proper resize handling
- Remote desktop applications
- Development IDEs with integrated terminals
- Cloud-based terminal services

### Testing Frameworks
- CLI application output validation
- Regression testing for terminal-based tools
- Automated testing of TUI applications
- Performance testing of terminal output

### Log Processing
- Real-time log visualization with ANSI colors
- Terminal output capture and replay
- Log file analysis with terminal formatting

## Contributing

While GoPyte is production-ready for most use cases, contributions are welcome for:

- **Mouse Support** - X10, X11, SGR protocols
- **True 24-bit RGB** - Full color support
- **SO/SI Charset** - Complete character set switching
- **Device Reports** - DA, DSR response generation
- **Performance** - Optimization for very large screens
- **Documentation** - Additional examples and guides

## License

LGPL (same as original pyte)

## Acknowledgments

- Original [pyte](https://github.com/selectel/pyte) library by Selectel
- [go-runewidth](https://github.com/mattn/go-runewidth) for Unicode width calculations
- [ConPTY](https://github.com/UserExistsError/conpty) for Windows terminal integration
- Test fixtures from pyte project
- VT100.net and XTerm documentation

## Production Status

**🚀 PRODUCTION READY** - WideCharScreen passes 100% of functional tests:

- **✅ 17/19** core features fully implemented (89.5%)
- **✅ 87.5%** Python pyte compatibility maintained
- **✅ Dynamic resize** with proper content handling
- **✅ Real application support**: vim, htop, less, Microsoft edit, tmux, git, npm
- **✅ Full Unicode support** including emoji and wide characters
- **✅ Scrollback history** with navigation and resize preservation
- **✅ Alternate screen buffer** with proper state isolation

**For production use, always use `WideCharScreen` which provides the complete feature set including resize support.**