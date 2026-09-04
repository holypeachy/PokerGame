package pokergame

type OutputType int

const (
	InputRequest OutputType = iota
	InvalidInput
	HandEnd
	GameEnd
)

func (o OutputType) String() string {
	switch o {
	case InputRequest:
		return "InputRequest"
	case InvalidInput:
		return "InvalidInput"
	case HandEnd:
		return "HandEnd"
	case GameEnd:
		return "GameEnd"
	default:
		return "Unknown"
	}
}
