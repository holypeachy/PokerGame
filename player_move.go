package pokergame

type PlayerMove int

const (
	Fold PlayerMove = iota
	Check
	Call
	Raise
)

func (m PlayerMove) String() string {
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
