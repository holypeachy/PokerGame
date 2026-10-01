package pokergame

type ActionType int

const (
	Fold ActionType = iota
	Check
	Call
	Raise
)

func (m ActionType) String() string {
	switch m {
	case Fold:
		return "Fold"
	case Check:
		return "Check"
	case Call:
		return "Call"
	case Raise:
		return "Raise"
	default:
		return "Unknown"
	}
}
