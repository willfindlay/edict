package output

// Typer simulates keyboard input to type transcribed text into the focused window.
type Typer interface {
	Type(text string) error
	CheckAvailable() error
}
