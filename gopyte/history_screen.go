package gopyte

import (
	"container/list"
)

// HistoryScreen extends NativeScreen with scrollback buffer support
type HistoryScreen struct {
	NativeScreen // Embedded, not pointer

	// History management
	history    *list.List // Doubly-linked list of historical lines
	maxHistory int        // Maximum lines to keep in history
	historyPos int        // Current position in history (0 = bottom/current)

	// Saved screen state for viewing history
	savedBuffer    [][]rune
	savedAttrs     [][]Attributes
	savedCursor    Cursor
	viewingHistory bool
}

// HistoryLine stores a line that scrolled off the top
type HistoryLine struct {
	Chars []rune
	Attrs []Attributes
}

// NewHistoryScreen creates a screen with scrollback buffer
func NewHistoryScreen(columns, lines, maxHistory int) *HistoryScreen {
	h := &HistoryScreen{
		NativeScreen:   *NewNativeScreen(columns, lines),
		history:        list.New(),
		maxHistory:     maxHistory,
		historyPos:     0,
		viewingHistory: false,
	}
	return h
}

// Override Linefeed to capture scrolling
func (h *HistoryScreen) Linefeed() {
	// Check if at bottom BEFORE incrementing
	if h.cursor.Y == h.lines-1 {
		// At bottom, scroll
		h.addToHistory(0)
		h.scrollUpInternal()
		// Stay at bottom
	} else {
		// Not at bottom, move down
		h.cursor.Y++
	}

	// In newline mode, also do CR
	if h.newlineMode {
		h.cursor.X = 0
	}
}

func (h *HistoryScreen) Index() {
	// Check if at bottom BEFORE incrementing
	if h.cursor.Y == h.lines-1 {
		// At bottom, scroll
		h.addToHistory(0)
		h.scrollUpInternal()
		// Stay at bottom
	} else {
		// Not at bottom, move down
		h.cursor.Y++
	}
}

// scrollUpInternal performs the actual scroll without calling parent
func (h *HistoryScreen) scrollUpInternal() {
	// Move all lines up by one
	copy(h.buffer[0:], h.buffer[1:])
	copy(h.attrs[0:], h.attrs[1:])

	// Clear the last line
	lastLine := h.lines - 1
	h.buffer[lastLine] = make([]rune, h.columns)
	h.attrs[lastLine] = make([]Attributes, h.columns)
	for i := 0; i < h.columns; i++ {
		h.buffer[lastLine][i] = ' '
	}
}

// addToHistory saves a line to the scrollback buffer
func (h *HistoryScreen) addToHistory(lineNum int) {
	if lineNum >= 0 && lineNum < h.lines {
		// Create a copy of the line
		line := HistoryLine{
			Chars: make([]rune, h.columns),
			Attrs: make([]Attributes, h.columns),
		}
		copy(line.Chars, h.buffer[lineNum])
		copy(line.Attrs, h.attrs[lineNum])

		// Add to history
		h.history.PushBack(line)

		// Trim history if it exceeds max
		if h.history.Len() > h.maxHistory {
			h.history.Remove(h.history.Front())
		}
	}
}

// ScrollUp scrolls the view up into history (like PageUp)
func (h *HistoryScreen) ScrollUp(lines int) {
	// Save current screen if we're not already viewing history
	if !h.viewingHistory {
		h.saveCurrentScreen()
		h.viewingHistory = true
	}

	// Calculate how many lines we can actually scroll
	maxScroll := h.history.Len() - h.historyPos
	if lines > maxScroll {
		lines = maxScroll
	}

	if lines <= 0 {
		return
	}

	h.historyPos += lines
	h.renderHistoryView()
}

// ScrollDown scrolls the view down towards current (like PageDown)
func (h *HistoryScreen) ScrollDown(lines int) {
	if !h.viewingHistory {
		return
	}

	h.historyPos -= lines
	if h.historyPos <= 0 {
		// Return to live view
		h.historyPos = 0
		h.restoreCurrentScreen()
		h.viewingHistory = false
	} else {
		h.renderHistoryView()
	}
}

// ScrollToBottom returns to the live terminal view
func (h *HistoryScreen) ScrollToBottom() {
	if h.viewingHistory {
		h.historyPos = 0
		h.restoreCurrentScreen()
		h.viewingHistory = false
	}
}

// saveCurrentScreen saves the current display for later restoration
func (h *HistoryScreen) saveCurrentScreen() {
	h.savedBuffer = make([][]rune, h.lines)
	h.savedAttrs = make([][]Attributes, h.lines)
	for i := 0; i < h.lines; i++ {
		h.savedBuffer[i] = make([]rune, h.columns)
		h.savedAttrs[i] = make([]Attributes, h.columns)
		copy(h.savedBuffer[i], h.buffer[i])
		copy(h.savedAttrs[i], h.attrs[i])
	}
	h.savedCursor = h.cursor
}

// restoreCurrentScreen restores the saved display
func (h *HistoryScreen) restoreCurrentScreen() {
	if h.savedBuffer != nil {
		h.buffer = h.savedBuffer
		h.attrs = h.savedAttrs
		h.cursor = h.savedCursor
		h.savedBuffer = nil
		h.savedAttrs = nil
		// Restore cursor visibility
		h.cursor.Hidden = false
	}
}

// renderHistoryView renders the history at the current position
func (h *HistoryScreen) renderHistoryView() {
	// Clear the buffer first
	for i := 0; i < h.lines; i++ {
		for j := 0; j < h.columns; j++ {
			h.buffer[i][j] = ' '
			h.attrs[i][j] = Attributes{}
		}
	}

	// We need to show historyPos lines from the end of history
	// If historyPos = 1, show the last line of history and rest from saved
	// If historyPos = history.Len(), show all history that fits

	totalLines := h.history.Len() + h.lines // history + current screen
	startLine := totalLines - h.historyPos - h.lines

	if startLine < 0 {
		startLine = 0
	}

	lineIdx := 0
	elem := h.history.Front()

	// Skip to the start position in history
	for i := 0; i < startLine && elem != nil; i++ {
		elem = elem.Next()
	}

	// Fill from history
	for elem != nil && lineIdx < h.lines {
		histLine := elem.Value.(HistoryLine)
		copy(h.buffer[lineIdx], histLine.Chars)
		copy(h.attrs[lineIdx], histLine.Attrs)
		elem = elem.Next()
		lineIdx++
	}

	// Fill remaining lines from saved buffer
	if lineIdx < h.lines && h.savedBuffer != nil {
		savedStart := 0
		if h.historyPos <= h.lines {
			// We're showing part of the current screen
			savedStart = h.lines - h.historyPos
		}

		for i := savedStart; i < h.lines && lineIdx < h.lines; i++ {
			copy(h.buffer[lineIdx], h.savedBuffer[i])
			copy(h.attrs[lineIdx], h.savedAttrs[i])
			lineIdx++
		}
	}

	// Hide cursor when viewing history
	h.cursor.Hidden = true
}

// Override Draw to exit history mode when new content arrives
func (h *HistoryScreen) Draw(text string) {
	// Exit history mode if we're in it
	if h.viewingHistory {
		h.ScrollToBottom()
	}

	// Now draw using embedded NativeScreen's implementation
	for _, ch := range text {
		// Check if we need to wrap
		if h.cursor.X >= h.columns {
			if h.autoWrap {
				h.cursor.X = 0
				// FIX: Check BEFORE incrementing
				if h.cursor.Y >= h.lines-1 {
					h.addToHistory(0)
					h.scrollUpInternal()
					// Stay at bottom line
				} else {
					h.cursor.Y++
				}
			} else {
				h.cursor.X = h.columns - 1
			}
		}

		// Place character
		if h.cursor.Y < h.lines && h.cursor.X < h.columns {
			h.buffer[h.cursor.Y][h.cursor.X] = ch
			h.attrs[h.cursor.Y][h.cursor.X] = h.cursor.Attrs
			h.cursor.X++
		}
	}
}

// Override EraseInDisplay to handle history clearing
func (h *HistoryScreen) EraseInDisplay(how int) {
	if h.viewingHistory {
		h.ScrollToBottom()
	}

	// Call embedded implementation
	h.NativeScreen.EraseInDisplay(how)

	// Clear history on full clear (ESC[2J or ESC[3J)
	if how == 2 || how == 3 {
		h.history.Init() // Clear the list
		h.historyPos = 0
	}
}

// Override Reset to clear history
func (h *HistoryScreen) Reset() {
	h.NativeScreen.Reset()
	h.history.Init() // Clear history
	h.historyPos = 0
	h.viewingHistory = false
	h.savedBuffer = nil
	h.savedAttrs = nil
}

// GetHistorySize returns the current number of lines in history
func (h *HistoryScreen) GetHistorySize() int {
	if h.history == nil {
		return 0
	}
	return h.history.Len()
}

// IsViewingHistory returns true if currently scrolled back in history
func (h *HistoryScreen) IsViewingHistory() bool {
	return h.viewingHistory
}

// GetDisplay returns the current display as strings (from embedded NativeScreen)
func (h *HistoryScreen) GetDisplay() []string {
	return h.NativeScreen.GetDisplay()
}

// GetCursor returns the current cursor position
func (h *HistoryScreen) GetCursor() (int, int) {
	return h.cursor.X, h.cursor.Y
}

// GetCursorObject returns the cursor object for testing
func (h *HistoryScreen) GetCursorObject() *Cursor {
	return &h.cursor
}

// Resize on HistoryScreen adds policy for scrollback when SHRINKING rows.
// We preserve the TOP..(newLines-1) region and PUSH cut bottom rows into history.
// Growing rows delegates to base then pads as usual.
func (h *HistoryScreen) Resize(newCols, newLines int) {
	if newCols <= 0 || newLines <= 0 {
		return
	}

	// If we are viewing history, jump back to live view first.
	if h.viewingHistory {
		h.ScrollToBottom()
	}

	oldLines := h.lines
	oldCols := h.columns

	// If rows will shrink and we’re not in alternate (alt handled elsewhere),
	// push the bottom lines that would be lost into history so they remain reachable.
	if newLines < oldLines {
		cut := oldLines - newLines
		start := oldLines - cut
		for i := start; i < oldLines; i++ {
			h.addToHistory(i)
		}
	}

	// Resize underlying NativeScreen buffers/attrs first with column logic.
	// Temporarily set base geometry so base Resize sees the old size.
	h.NativeScreen.Resize(newCols, newLines)

	// Cursor already clamped by base Resize
	_ = oldCols
}
