package tui

// verboseDepth controls how deeply the verbose view renders nested
// attributes. Currently a placeholder: the view always renders the full
// terraform-show text. A follow-up patch can use this flag to limit depth
// when the user opts into "compact".
var verboseDepth = "compact"

// SetVerboseDepth lets cmd/root.go push the configured value at startup.
// Only "compact" and "full" are accepted; anything else is ignored.
func SetVerboseDepth(d string) {
	if d == "full" || d == "compact" {
		verboseDepth = d
	}
}

// VerboseDepth exposes the current setting (mostly useful for tests).
func VerboseDepth() string { return verboseDepth }
