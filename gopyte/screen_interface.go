package gopyte

// Screen represents a terminal screen
type Screen interface {
	// Basic operations
	Draw(text string)
	Bell()
	Backspace()
	Tab()
	Linefeed()
	CarriageReturn()
	ShiftOut()
	ShiftIn()

	// Cursor movement
	CursorUp(count int)
	CursorDown(count int)
	CursorForward(count int)
	CursorBack(count int)
	CursorUp1(count int)
	CursorDown1(count int)
	CursorPosition(line, column int)
	CursorToColumn(column int)
	CursorToLine(line int)

	// Screen manipulation
	Reset()
	Index()
	ReverseIndex()
	SetTabStop()
	ClearTabStop(how int)
	SaveCursor()
	RestoreCursor()

	// Line operations
	InsertLines(count int)
	DeleteLines(count int)
	InsertCharacters(count int)
	DeleteCharacters(count int)
	EraseCharacters(count int)
	EraseInLine(how int, private bool)
	EraseInDisplay(how int)

	// Modes
	SetMode(modes []int, private bool)
	ResetMode(modes []int, private bool)

	// Attributes
	SelectGraphicRendition(attrs []int)

	// Charset
	DefineCharset(code, mode string)

	// Margins
	SetMargins(top, bottom int)

	// Reports
	ReportDeviceAttributes(mode int, private bool)
	ReportDeviceStatus(mode int)

	// Window operations
	SetTitle(title string)
	SetIconName(name string)

	// Alignment
	AlignmentDisplay()

	// Debug
	Debug(args ...interface{})

	// Process communication
	WriteProcessInput(data string)
}
