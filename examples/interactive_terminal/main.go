// examples/interactive_terminal/main.go
//
// # Interactive Terminal Example for GoPyte
//
// This example demonstrates GoPyte's terminal emulation capabilities by providing
// an interactive shell where you can:
//   - Execute Windows commands and see parsed output
//   - Test ANSI escape sequence handling
//   - Explore scrollback history
//   - Switch between main and alternate screen buffers
//
// Usage:
//
//	go run main.go
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/scottpeterman/gopyte/gopyte"
)

const (
	screenWidth  = 120
	screenHeight = 40
	historySize  = 5000
)

type Terminal struct {
	screen *gopyte.WideCharScreen
	stream *gopyte.Stream
}

func NewTerminal() *Terminal {
	screen := gopyte.NewWideCharScreen(screenWidth, screenHeight, historySize)
	stream := gopyte.NewStream(screen, false)
	return &Terminal{screen: screen, stream: stream}
}

func main() {
	term := NewTerminal()

	printBanner()
	printHelp()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\n> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		parts := strings.SplitN(input, " ", 2)
		cmd := strings.ToLower(parts[0])
		args := ""
		if len(parts) > 1 {
			args = parts[1]
		}

		switch cmd {
		case "run", "r":
			if args != "" {
				term.runCommand(args, false)
			} else {
				fmt.Println("Usage: run <command>")
				fmt.Println("Example: run dir /w")
			}

		case "ps", "powershell":
			if args != "" {
				term.runCommand(args, true)
			} else {
				fmt.Println("Usage: ps <command>")
				fmt.Println("Example: ps Get-Date")
			}

		case "feed", "f":
			if args != "" {
				term.feedText(args)
			} else {
				fmt.Println("Usage: feed <text>")
				fmt.Println("Example: feed \\x1b[31mRed Text\\x1b[0m")
			}

		case "show", "s":
			term.showDisplay(false)

		case "raw":
			term.showDisplay(true)

		case "clear", "cls":
			term.clear()

		case "reset":
			term.reset()

		case "up", "u":
			term.scrollUp(args)

		case "down", "d":
			term.scrollDown(args)

		case "pos", "cursor":
			term.showCursorPos()

		case "info", "i":
			term.showInfo()

		case "alt", "alternate":
			term.toggleAlternateScreen()

		case "history", "h":
			term.showHistory()

		case "demo":
			term.runDemo()

		case "help", "?":
			printHelp()

		case "examples", "ex":
			printExamples()

		case "quit", "exit", "q":
			fmt.Println("Goodbye!")
			return

		default:
			fmt.Printf("Unknown command: %s (type 'help' for commands)\n", cmd)
		}
	}
}

func (t *Terminal) runCommand(command string, isPowerShell bool) {
	shellName := "CMD"
	if isPowerShell {
		shellName = "PowerShell"
	}

	fmt.Printf("\n=== Running %s Command ===\n", shellName)
	fmt.Printf("Command: %s\n", command)

	var cmd *exec.Cmd
	if isPowerShell {
		cmd = exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", command)
	} else {
		cmd = exec.Command("cmd", "/c", command)
	}

	// Set up environment for better compatibility
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\n" + stderr.String()
	}

	if output != "" {
		t.stream.Feed(output)
		fmt.Printf("Output captured: %d bytes\n", len(output))
	}

	if err != nil {
		fmt.Printf("Command exit code: %v\n", err)
	}

	t.showDisplay(false)
}

func (t *Terminal) feedText(text string) {
	// Replace common escape sequences
	replacements := map[string]string{
		"\\n":   "\n",
		"\\r":   "\r",
		"\\t":   "\t",
		"\\x1b": "\x1b",
		"\\033": "\x1b",
		"\\e":   "\x1b",
		"\\a":   "\a",
		"\\b":   "\b",
		"\\f":   "\f",
		"\\v":   "\v",
	}

	for old, new := range replacements {
		text = strings.ReplaceAll(text, old, new)
	}

	t.stream.Feed(text)
	fmt.Printf("Fed %d bytes to terminal\n", len(text))
}

func (t *Terminal) showDisplay(raw bool) {
	fmt.Println("\n╔══════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                        TERMINAL DISPLAY                              ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════════╝")

	display := t.screen.GetDisplay()
	lineCount := 0

	for i, line := range display {
		if raw {
			// Show with visible whitespace
			visible := strings.ReplaceAll(line, " ", "·")
			visible = strings.ReplaceAll(visible, "\t", "→→→→")
			if visible != strings.Repeat("·", len(line)) || i < 5 {
				fmt.Printf("%02d│ %s│\n", i, visible)
				lineCount++
			}
		} else {
			trimmed := strings.TrimRight(line, " ")
			if trimmed != "" {
				fmt.Printf("%02d│ %s\n", i, trimmed)
				lineCount++
			}
		}
	}

	if lineCount == 0 {
		fmt.Println("   (empty display)")
	}

	t.showStatus()
}

func (t *Terminal) showStatus() {
	x, y := t.screen.GetCursor()

	fmt.Println("\n╔══════════════════════════════════════════════════════════════════════╗")
	fmt.Printf("║ Cursor: (%3d,%3d)  History: %4d lines  ", x, y, t.screen.GetHistorySize())

	if t.screen.IsUsingAlternate() {
		fmt.Printf("Screen: ALTERNATE          ║\n")
	} else if t.screen.IsViewingHistory() {
		fmt.Printf("Screen: MAIN (in history)  ║\n")
	} else {
		fmt.Printf("Screen: MAIN               ║\n")
	}
	fmt.Println("╚══════════════════════════════════════════════════════════════════════╝")
}

func (t *Terminal) clear() {
	t.stream.Feed("\x1b[2J\x1b[H")
	fmt.Println("Screen cleared")
}

func (t *Terminal) reset() {
	t.stream.Feed("\x1bc")
	fmt.Println("Terminal reset")
}

func (t *Terminal) scrollUp(args string) {
	lines := 5
	if args != "" {
		if n, err := strconv.Atoi(args); err == nil {
			lines = n
		}
	}
	t.screen.ScrollUp(lines)
	fmt.Printf("Scrolled up %d lines\n", lines)
	t.showDisplay(false)
}

func (t *Terminal) scrollDown(args string) {
	lines := 5
	if args != "" {
		if n, err := strconv.Atoi(args); err == nil {
			lines = n
		}
	}
	t.screen.ScrollDown(lines)
	fmt.Printf("Scrolled down %d lines\n", lines)
	t.showDisplay(false)
}

func (t *Terminal) showCursorPos() {
	x, y := t.screen.GetCursor()
	fmt.Printf("Cursor position: column=%d, row=%d (0-indexed)\n", x, y)
}

func (t *Terminal) showInfo() {
	fmt.Printf("\nTerminal Information:\n")
	fmt.Printf("  Screen dimensions: %dx%d\n", screenWidth, screenHeight)
	fmt.Printf("  History buffer: %d lines maximum\n", historySize)
	fmt.Printf("  Current history: %d lines\n", t.screen.GetHistorySize())
	fmt.Printf("  Viewing history: %v\n", t.screen.IsViewingHistory())
	fmt.Printf("  Alternate screen: %v\n", t.screen.IsUsingAlternate())

	x, y := t.screen.GetCursor()
	fmt.Printf("  Cursor position: (%d, %d)\n", x, y)
	fmt.Printf("  Operating System: %s\n", runtime.GOOS)
}

func (t *Terminal) toggleAlternateScreen() {
	if t.screen.IsUsingAlternate() {
		t.stream.Feed("\x1b[?1049l")
		fmt.Println("Switched to MAIN screen")
	} else {
		t.stream.Feed("\x1b[?1049h")
		fmt.Println("Switched to ALTERNATE screen (vim/less mode)")
	}
}

func (t *Terminal) showHistory() {
	fmt.Printf("\nHistory Buffer Status:\n")
	fmt.Printf("  Total lines in history: %d\n", t.screen.GetHistorySize())
	fmt.Printf("  Maximum history size: %d\n", historySize)
	fmt.Printf("  Currently viewing history: %v\n", t.screen.IsViewingHistory())

	if t.screen.GetHistorySize() > 0 {
		fmt.Println("\nUse 'up' and 'down' commands to navigate history")
	}
}

func (t *Terminal) runDemo() {
	fmt.Println("\n=== Running Demo Sequence ===\n")

	demos := []struct {
		name  string
		feed  string
		pause bool
	}{
		{"Clear screen", "\x1b[2J\x1b[H", false},
		{"Header", "=== GoPyte Demo ===\n\n", false},
		{"Simple text", "Hello from GoPyte Terminal Emulator!\n", false},
		{"ANSI Colors", "\x1b[31mRed \x1b[32mGreen \x1b[34mBlue \x1b[33mYellow \x1b[35mMagenta \x1b[36mCyan \x1b[0mReset\n", false},
		{"Text styles", "\x1b[1mBold \x1b[3mItalic \x1b[4mUnderline \x1b[7mReverse \x1b[0mNormal\n", false},
		{"Background colors", "\x1b[41m Red BG \x1b[42m Green BG \x1b[44m Blue BG \x1b[0m Normal\n", false},
		{"256 colors", "\x1b[38;5;196mColor 196 \x1b[38;5;46mColor 46 \x1b[38;5;21mColor 21\x1b[0m\n", false},
		{"Progress bar start", "\nDownload Progress:\n", false},
		{"Progress 0%", "\r[          ] 0%", true},
		{"Progress 25%", "\r[██        ] 25%", true},
		{"Progress 50%", "\r[█████     ] 50%", true},
		{"Progress 75%", "\r[███████   ] 75%", true},
		{"Progress 100%", "\r[██████████] 100% Complete!\n", false},
		{"Unicode support", "\nUnicode: 你好世界 • Emoji: 🚀 ✨ 🎉 • Math: ∑ ∏ ∫ √ ∞\n", false},
		{"Box drawing", "\n┌──────────────────┐\n│  Box Drawing     │\n├──────────────────┤\n│  Works Great!    │\n└──────────────────┘\n", false},
		{"Tab stops", "Col1\tCol2\tCol3\tCol4\n────\t────\t────\t────\nData\tMore\tStuff\tHere\n", false},
		{"Cursor movement", "\nLine 1\nLine 2\x1b[A←moved up\x1b[B\nLine 3\n", false},
		{"Footer", "\n=== Demo Complete ===\n", false},
	}

	for _, demo := range demos {
		if demo.pause {
			fmt.Printf("  %s", demo.name)
		} else {
			fmt.Printf("  %s\n", demo.name)
		}
		t.stream.Feed(demo.feed)
	}

	fmt.Println("\nDemo complete! Type 'show' to see the full result")
}

func printBanner() {
	banner := `
╔════════════════════════════════════════════════════════════════════════╗
║                                                                        ║
║                    GoPyte Interactive Terminal                        ║
║                 Terminal Emulator Testing Tool                        ║
║                                                                        ║
╚════════════════════════════════════════════════════════════════════════╝
`
	fmt.Print(banner)
}

func printHelp() {
	fmt.Println("\n COMMANDS:")
	fmt.Println("  run <cmd>      (r)   Execute Windows command")
	fmt.Println("  ps <cmd>            Execute PowerShell command")
	fmt.Println("  feed <text>    (f)   Feed text/ANSI sequences to terminal")
	fmt.Println("  show          (s)   Display current screen")
	fmt.Println("  raw                 Display screen with visible whitespace")
	fmt.Println("  clear         (cls) Clear the screen")
	fmt.Println("  reset              Full terminal reset")
	fmt.Println("  up [n]        (u)   Scroll up n lines (default: 5)")
	fmt.Println("  down [n]      (d)   Scroll down n lines (default: 5)")
	fmt.Println("  pos                Show cursor position")
	fmt.Println("  info          (i)   Show terminal information")
	fmt.Println("  alt                Toggle alternate screen buffer")
	fmt.Println("  history       (h)   Show history buffer status")
	fmt.Println("  demo               Run demonstration")
	fmt.Println("  examples      (ex)  Show example commands")
	fmt.Println("  help          (?)   Show this help")
	fmt.Println("  quit          (q)   Exit the program")
}

func printExamples() {
	fmt.Println("\n EXAMPLE COMMANDS:")
	fmt.Println("\n Windows Commands:")
	fmt.Println("  run dir                    # Directory listing")
	fmt.Println("  run dir /w                 # Wide format directory")
	fmt.Println("  run type file.txt          # Display file contents")
	fmt.Println("  run echo Hello World       # Simple echo")
	fmt.Println("  run ipconfig              # Network configuration")
	fmt.Println("  run ping 8.8.8.8 -n 5     # Ping with updates")
	fmt.Println("  run tree /F               # Directory tree")

	fmt.Println("\n PowerShell Commands:")
	fmt.Println("  ps Get-Date                # Current date/time")
	fmt.Println("  ps Get-Process | Select-Object -First 5")
	fmt.Println("  ps Write-Host 'Colored' -ForegroundColor Red")
	fmt.Println("  ps Get-ChildItem | Format-Table")

	fmt.Println("\n ANSI Sequences:")
	fmt.Println("  feed \\x1b[31mRed Text\\x1b[0m     # Colored text")
	fmt.Println("  feed \\x1b[1mBold\\x1b[0m          # Bold text")
	fmt.Println("  feed Line1\\nLine2\\nLine3       # Multiple lines")
	fmt.Println("  feed \\x1b[2J\\x1b[H             # Clear and home")
	fmt.Println("  feed Progress:\\r[####    ] 40% # Progress bar")
}
