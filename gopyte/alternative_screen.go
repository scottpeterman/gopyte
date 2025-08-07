package gopyte

import (
	"container/list"
)

// AlternateScreen adds alternative screen buffer support to HistoryScreen
// This is used by applications like vim, less, etc.
type AlternateScreen struct {
	*HistoryScreen

	// Alternative screen state
	mainBuffer   [][]rune
	mainAttrs    [][]Attributes
	mainCursor   Cursor
	mainTabStops map[int]bool
	mainHistory  *list.List

	altBuffer   [][]rune
	altAttrs    [][]Attributes
	altCursor   Cursor
	altTabStops map[int]bool

	usingAlternate bool
}

// NewAlternateScreen creates a screen with both main and alternate buffers
func NewAlternateScreen(columns, lines, maxHistory int) *AlternateScreen {
	a := &AlternateScreen{
		HistoryScreen:  NewHistoryScreen(columns, lines, maxHistory),
		usingAlternate: false,
	}

	// Initialize alternate buffer
	a.altBuffer = make([][]rune, lines)
	a.altAttrs = make([][]Attributes, lines)
	for i := 0; i < lines; i++ {
		a.altBuffer[i] = make([]rune, columns)
		a.altAttrs[i] = make([]Attributes, columns)
		for j := 0; j < columns; j++ {
			a.altBuffer[i][j] = ' '
		}
	}

	// Initialize alternate tab stops
	a.altTabStops = make(map[int]bool)
	for i := 0; i < columns; i += 8 {
		a.altTabStops[i] = true
	}

	return a
}

// Override SetMode to handle alternate screen switching
func (a *AlternateScreen) SetMode(modes []int, private bool) {
	if private {
		for _, mode := range modes {
			switch mode {
			case 1049, 1047, 47: // Alternate screen modes
				if !a.usingAlternate {
					a.switchToAlternate()
				}
			case 1048: // Save cursor
				if a.usingAlternate {
					a.altCursor = a.cursor
				} else {
					a.mainCursor = a.cursor
				}
			}
		}
	}

	// Call parent implementation for other modes
	a.HistoryScreen.SetMode(modes, private)
}

// Override ResetMode to handle alternate screen switching
func (a *AlternateScreen) ResetMode(modes []int, private bool) {
	if private {
		for _, mode := range modes {
			switch mode {
			case 1049, 1047, 47: // Exit alternate screen
				if a.usingAlternate {
					a.switchToMain()
				}
			case 1048: // Restore cursor
				if a.usingAlternate {
					a.cursor = a.altCursor
				} else {
					a.cursor = a.mainCursor
				}
			}
		}
	}

	// Call parent implementation for other modes
	a.HistoryScreen.ResetMode(modes, private)
}

// switchToAlternate switches to the alternate screen buffer
func (a *AlternateScreen) switchToAlternate() {
	// Save main screen state
	a.mainBuffer = a.buffer
	a.mainAttrs = a.attrs
	a.mainCursor = a.cursor
	a.mainTabStops = a.tabStops
	a.mainHistory = a.history

	// Clear alternate buffer before switching
	for i := 0; i < a.lines; i++ {
		for j := 0; j < a.columns; j++ {
			a.altBuffer[i][j] = ' '
			a.altAttrs[i][j] = DefaultAttributes()
		}
	}

	// Switch to alternate
	a.buffer = a.altBuffer
	a.attrs = a.altAttrs
	a.cursor = Cursor{X: 0, Y: 0, Attrs: DefaultAttributes()}
	a.tabStops = a.altTabStops

	// Alternate screen doesn't use history, use empty list
	a.history = list.New()
	a.usingAlternate = true

	// If we were viewing history, exit that mode
	if a.viewingHistory {
		a.viewingHistory = false
		a.historyPos = 0
	}
}

// switchToMain switches back to the main screen buffer
func (a *AlternateScreen) switchToMain() {
	if !a.usingAlternate {
		return
	}

	// Save alternate state (in case we switch back)
	a.altBuffer = a.buffer
	a.altAttrs = a.attrs
	a.altCursor = a.cursor
	a.altTabStops = a.tabStops

	// Restore main screen
	a.buffer = a.mainBuffer
	a.attrs = a.mainAttrs
	a.cursor = a.mainCursor
	a.tabStops = a.mainTabStops
	a.history = a.mainHistory

	a.usingAlternate = false
}

// Override methods that shouldn't save to history in alternate mode

func (a *AlternateScreen) Linefeed() {
	if a.usingAlternate {
		// Check if at bottom BEFORE incrementing
		if a.cursor.Y == a.lines-1 {
			// At bottom, scroll without history
			a.scrollUpNoHistory()
			// Stay at bottom
		} else {
			// Not at bottom, move down
			a.cursor.Y++
		}

		if a.newlineMode {
			a.cursor.X = 0
		}
	} else {
		// Use parent implementation with history
		a.HistoryScreen.Linefeed()
	}
}

func (a *AlternateScreen) Index() {
	if a.usingAlternate {
		// Check if at bottom BEFORE incrementing
		if a.cursor.Y == a.lines-1 {
			// At bottom, scroll without history
			a.scrollUpNoHistory()
			// Stay at bottom
		} else {
			// Not at bottom, move down
			a.cursor.Y++
		}
	} else {
		a.HistoryScreen.Index()
	}
}

// scrollUpNoHistory scrolls without saving to history (for alternate screen)
func (a *AlternateScreen) scrollUpNoHistory() {
	// Move all lines up by one
	copy(a.buffer[0:], a.buffer[1:])
	copy(a.attrs[0:], a.attrs[1:])

	// Clear the last line
	lastLine := a.lines - 1
	a.buffer[lastLine] = make([]rune, a.columns)
	a.attrs[lastLine] = make([]Attributes, a.columns)
	for i := 0; i < a.columns; i++ {
		a.buffer[lastLine][i] = ' '
	}
}

// Override Draw to handle alternate screen
func (a *AlternateScreen) Draw(text string) {
	if a.usingAlternate {
		// Don't exit history mode in alternate screen (there is no history)
		// Just draw normally using the base implementation
		a.drawTextDirect(text)
	} else {
		a.HistoryScreen.Draw(text)
	}
}

// drawTextDirect draws text without history handling
func (a *AlternateScreen) drawTextDirect(text string) {
	for _, ch := range text {
		// Check if we need to wrap
		if a.cursor.X >= a.columns {
			if a.autoWrap {
				a.cursor.X = 0
				a.cursor.Y++
				if a.cursor.Y >= a.lines {
					a.scrollUpNoHistory()
					a.cursor.Y = a.lines - 1
				}
			} else {
				a.cursor.X = a.columns - 1
			}
		}

		// Place character
		if a.cursor.Y < a.lines && a.cursor.X < a.columns {
			a.buffer[a.cursor.Y][a.cursor.X] = ch
			a.attrs[a.cursor.Y][a.cursor.X] = a.cursor.Attrs
			a.cursor.X++
		}
	}
}

// Override Reset to handle alternate screen
func (a *AlternateScreen) Reset() {
	// Exit alternate screen if we're in it
	if a.usingAlternate {
		a.switchToMain()
	}

	// Reset both buffers
	a.HistoryScreen.Reset()

	// Also reset alternate buffer
	for i := 0; i < a.lines; i++ {
		for j := 0; j < a.columns; j++ {
			a.altBuffer[i][j] = ' '
			a.altAttrs[i][j] = DefaultAttributes()
		}
	}
	a.altCursor = Cursor{X: 0, Y: 0, Attrs: DefaultAttributes()}
	a.altTabStops = make(map[int]bool)
	for i := 0; i < a.columns; i += 8 {
		a.altTabStops[i] = true
	}
}

// IsUsingAlternate returns true if using alternate screen buffer
func (a *AlternateScreen) IsUsingAlternate() bool {
	return a.usingAlternate
}

// Override history methods to disable in alternate screen
func (a *AlternateScreen) ScrollUp(lines int) {
	if !a.usingAlternate {
		a.HistoryScreen.ScrollUp(lines)
	}
	// No-op in alternate screen
}

func (a *AlternateScreen) ScrollDown(lines int) {
	if !a.usingAlternate {
		a.HistoryScreen.ScrollDown(lines)
	}
	// No-op in alternate screen
}

func (a *AlternateScreen) ScrollToBottom() {
	if !a.usingAlternate {
		a.HistoryScreen.ScrollToBottom()
	}
	// No-op in alternate screen
}
