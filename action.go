package pokergame

type ActionType int

const (
	Fold ActionType = iota
	Check
	Call
	Raise
	Leave
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
	case Leave:
		return "Leave"
	default:
		return "Unknown"
	}
}
