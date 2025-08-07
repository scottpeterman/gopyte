package gopyte

import (
	// "container/list"
	runewidth "github.com/mattn/go-runewidth"
)

// WideCharScreen adds wide character (CJK, emoji) support to AlternateScreen
type WideCharScreen struct {
	*AlternateScreen

	// Track cell widths (0 = continuation, 1 = normal, 2 = wide start)
	cellWidths     [][]int
	altCellWidths  [][]int
	mainCellWidths [][]int
}

// NewWideCharScreen creates a screen with wide character support
func NewWideCharScreen(columns, lines, maxHistory int) *WideCharScreen {
	alt := NewAlternateScreen(columns, lines, maxHistory)

	w := &WideCharScreen{
		AlternateScreen: alt,
	}

	// Initialize cell width tracking for main buffer
	w.cellWidths = make([][]int, lines)
	for i := 0; i < lines; i++ {
		w.cellWidths[i] = make([]int, columns)
		for j := 0; j < columns; j++ {
			w.cellWidths[i][j] = 1 // Default to normal width
		}
	}

	// Initialize for alternate buffer
	w.altCellWidths = make([][]int, lines)
	for i := 0; i < lines; i++ {
		w.altCellWidths[i] = make([]int, columns)
		for j := 0; j < columns; j++ {
			w.altCellWidths[i][j] = 1
		}
	}

	// Store reference for later use
	w.mainCellWidths = w.cellWidths

	return w
}

// Override Draw to handle wide characters and emojis
func (w *WideCharScreen) Draw(text string) {
	// Exit history mode if in main screen and viewing history
	if !w.usingAlternate && w.viewingHistory {
		w.ScrollToBottom()
	}

	// Process each character with width awareness
	for _, ch := range text {
		w.drawChar(ch)
	}
}

// drawChar handles a single character with width calculation

// drawChar handles a single character with width calculation
func (w *WideCharScreen) drawChar(ch rune) {
	// Get the display width of the character
	charWidth := runewidth.RuneWidth(ch)

	// Handle zero-width characters (combining marks, etc.)
	if charWidth == 0 {
		w.handleZeroWidth(ch)
		return
	}

	// Check if the character fits at current position
	if w.cursor.X+charWidth > w.columns {
		if w.autoWrap {
			// Wide character doesn't fit, wrap to next line
			w.cursor.X = 0
			w.cursor.Y++
			if w.cursor.Y >= w.lines {
				if w.usingAlternate {
					w.scrollUpNoHistory()
				} else {
					w.addToHistory(0)
					w.scrollUpInternal()
				}
				w.cursor.Y = w.lines - 1
			}
		} else {
			// Can't place character at edge without wrapping
			return
		}
	}

	// Now place the character at the (possibly new) cursor position
	if w.cursor.Y < w.lines && w.cursor.X < w.columns {
		// Clear any wide character we're overwriting
		w.clearCellAt(w.cursor.Y, w.cursor.X)

		w.buffer[w.cursor.Y][w.cursor.X] = ch
		w.attrs[w.cursor.Y][w.cursor.X] = w.cursor.Attrs
		w.cellWidths[w.cursor.Y][w.cursor.X] = charWidth

		if charWidth == 2 {
			// Mark the next cell as continuation
			if w.cursor.X+1 < w.columns {
				w.buffer[w.cursor.Y][w.cursor.X+1] = 0 // Null char for continuation
				w.attrs[w.cursor.Y][w.cursor.X+1] = w.cursor.Attrs
				w.cellWidths[w.cursor.Y][w.cursor.X+1] = 0 // Continuation marker
			}
		}

		w.cursor.X += charWidth
	}
}

// handleZeroWidth handles zero-width combining characters
func (w *WideCharScreen) handleZeroWidth(ch rune) {
	// Combining characters attach to the previous character
	if w.cursor.X > 0 {
		// Combine with previous character
		prevX := w.cursor.X - 1
		if w.cellWidths[w.cursor.Y][prevX] == 2 && prevX > 0 {
			// Previous is a wide character, combine with its start
			prevX--
		}

		// Append the combining character
		existing := w.buffer[w.cursor.Y][prevX]
		if existing != 0 && existing != ' ' {
			// In a real implementation, we'd normalize the combination
			// For now, we'll just store the base character
			// A full implementation would need to handle Unicode normalization
		}
	} else if w.cursor.Y > 0 {
		// Combine with last character of previous line
		prevY := w.cursor.Y - 1
		prevX := w.columns - 1

		// Find the last actual character
		for prevX >= 0 && w.cellWidths[prevY][prevX] == 0 {
			prevX--
		}

		if prevX >= 0 && w.buffer[prevY][prevX] != ' ' {
			// Would combine here in full implementation
		}
	}
}

// clearCellAt clears a cell, handling wide characters properly
func (w *WideCharScreen) clearCellAt(y, x int) {
	if y >= w.lines || x >= w.columns {
		return
	}

	width := w.cellWidths[y][x]

	// If this is a continuation cell, clear the start cell too
	if width == 0 && x > 0 {
		w.clearCellAt(y, x-1)
		return
	}

	// Clear this cell
	w.buffer[y][x] = ' '
	w.attrs[y][x] = DefaultAttributes()
	w.cellWidths[y][x] = 1

	// If this was a wide character, clear its continuation
	if width == 2 && x+1 < w.columns {
		w.buffer[y][x+1] = ' '
		w.attrs[y][x+1] = DefaultAttributes()
		w.cellWidths[y][x+1] = 1
	}
}

// Override cursor movement to handle wide characters
func (w *WideCharScreen) CursorBack(count int) {
	for i := 0; i < count; i++ {
		if w.cursor.X <= 0 {
			break
		}

		// Skip over continuation cells
		w.cursor.X--
		for w.cursor.X > 0 && w.cellWidths[w.cursor.Y][w.cursor.X] == 0 {
			w.cursor.X--
		}
	}

	if w.cursor.X < 0 {
		w.cursor.X = 0
	}
}

func (w *WideCharScreen) CursorForward(count int) {
	for i := 0; i < count; i++ {
		if w.cursor.X >= w.columns-1 {
			break
		}

		// Skip over continuation cells
		if w.cellWidths[w.cursor.Y][w.cursor.X] == 2 {
			w.cursor.X += 2
		} else {
			w.cursor.X++
		}
	}

	if w.cursor.X >= w.columns {
		w.cursor.X = w.columns - 1
	}
}

// Override EraseCharacters to handle wide characters
func (w *WideCharScreen) EraseCharacters(count int) {
	x := w.cursor.X
	for i := 0; i < count && x < w.columns; i++ {
		w.clearCellAt(w.cursor.Y, x)

		// Move to next character position
		if x < w.columns-1 && w.cellWidths[w.cursor.Y][x+1] == 0 {
			x += 2 // Was a wide character
		} else {
			x++
		}
	}
}

// Override GetDisplay to handle wide characters properly
func (w *WideCharScreen) GetDisplay() []string {
	lines := make([]string, w.lines)
	for y := 0; y < w.lines; y++ {
		runes := make([]rune, 0, w.columns)
		for x := 0; x < w.columns; x++ {
			if w.cellWidths[y][x] == 0 {
				// Skip continuation cells
				continue
			}
			ch := w.buffer[y][x]
			if ch != 0 { // Don't include null characters
				runes = append(runes, ch)
			}
		}
		lines[y] = string(runes)
	}
	return lines
}

// Override switching to handle cell widths
func (w *WideCharScreen) switchToAlternate() {
	// Save main screen cell widths
	w.mainCellWidths = w.cellWidths

	// Call parent
	w.AlternateScreen.switchToAlternate()

	// Switch to alternate cell widths
	w.cellWidths = w.altCellWidths
}

func (w *WideCharScreen) switchToMain() {
	// Save alternate cell widths
	w.altCellWidths = w.cellWidths

	// Call parent
	w.AlternateScreen.switchToMain()

	// Restore main cell widths
	if w.mainCellWidths != nil {
		w.cellWidths = w.mainCellWidths
	}
}

// Helper to check if a rune is an emoji
func isEmoji(r rune) bool {
	// Basic emoji detection - would need a more comprehensive check in production
	return (r >= 0x1F300 && r <= 0x1F9FF) || // Misc symbols and pictographs
		(r >= 0x2600 && r <= 0x27BF) || // Misc symbols
		(r >= 0x1F600 && r <= 0x1F64F) || // Emoticons
		(r >= 0x1F680 && r <= 0x1F6FF) || // Transport and map
		(r == 0x2764) || (r == 0x2665) // Hearts
}
