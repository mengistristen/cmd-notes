package note

// Available states
const (
	NONE = iota
	TODO
	IN_PROGRESS
	COMPLETE
)

// Available priorities
const (
	LOW = iota
	MEDIUM
	HIGH
)

// ANSI escape codes for colors
const (
	RED    = "\033[91m"
	YELLOW = "\033[93m"
	GREEN  = "\033[92m"
	BLUE   = "\033[34m"
	RESET  = "\033[0m"
)

type Note struct {
	Priority int
	State    int
	Contents string
}

func (n Note) FormatPriority() string {
	priority := ""

	switch n.Priority {
	case LOW:
		priority = "(\033[34m\u2193\033[0m)"
	case MEDIUM:
		priority = "(-)"
	case HIGH:
		priority = "(\033[31m\u2191\033[0m)"
	}

	return priority
}

func (n Note) FormatContents() string {
	var result string

	switch n.State {
	case TODO:
		result = RED + n.Contents + RESET
	case IN_PROGRESS:
		result = YELLOW + n.Contents + RESET
	case COMPLETE:
		result = GREEN + n.Contents + RESET
	default:
		result = n.Contents
	}

	return result
}

func (n *Note) Promote() {
	switch n.State {
	case NONE:
		n.State = TODO
	case TODO:
		n.State = IN_PROGRESS
	case IN_PROGRESS:
		n.State = COMPLETE
	}
}

func (n *Note) Demote() {
	switch n.State {
	case TODO:
		n.State = NONE
	case IN_PROGRESS:
		n.State = TODO
	case COMPLETE:
		n.State = IN_PROGRESS
	}
}

func (n *Note) IncreasePriority() {
	switch n.Priority {
	case LOW:
		n.Priority = MEDIUM
	case MEDIUM:
		n.Priority = HIGH
	}
}

func (n *Note) DecreasePriority() {
	switch n.Priority {
	case MEDIUM:
		n.Priority = LOW
	case HIGH:
		n.Priority = MEDIUM
	}
}
