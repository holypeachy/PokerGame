package pokergame

import pokeralgo "pokeralgo"

type BlindIndices struct {
	Dealer     int
	BigBlind   int
	SmallBlind int
}

type PotState struct {
	EligiblePlayerIDs []string
	Amount            int

	WinnerIDs []string
}

type Street int

const (
	Preflop Street = iota
	Flop
	Turn
	River
)

func (s Street) String() string {
	switch s {
	case Preflop:
		return "Preflop"
	case Flop:
		return "Flop"
	case Turn:
		return "Turn"
	case River:
		return "River"
	default:
		return "Unknown"
	}
}

type Event struct {
	HandNumber   int
	Street       Street
	Type         EventType
	Players      []PlayerState
	BlindIndices *BlindIndices
	Board        []pokeralgo.Card
	PlayerID     *string
	Action       *Action
	Pots         []PotState
	Err          error
}

type EventType int

const (
	GameStarted EventType = iota
	BlindsAdvanced
	HandStarted
	BlindsPosted
	HoleCardsDealt
	StreetStarted
	ActionRequested
	ActionValid
	ActionInvalid
	StreetEnded
	OnePlayerLeft
	RunToShowdown
	ShowdownStarted
	PotsCreated
	WinnersDetermined
	ChipsAwarded
	PlayerBusted
	HandEnded
	GameEnded
	ErrorState
)

func (e EventType) String() string {
	switch e {
	case GameStarted:
		return "GameStarted"
	case BlindsAdvanced:
		return "BlindsAdvanced"
	case HandStarted:
		return "HandStarted"
	case BlindsPosted:
		return "BlindsPosted"
	case HoleCardsDealt:
		return "HoleCardsDealt"
	case StreetStarted:
		return "StreetStarted"
	case ActionRequested:
		return "ActionRequested"
	case ActionValid:
		return "ActionValid"
	case ActionInvalid:
		return "ActionInvalid"
	case StreetEnded:
		return "StreetEnded"
	case OnePlayerLeft:
		return "OnePlayerLeft"
	case RunToShowdown:
		return "RunToShowdown"
	case ShowdownStarted:
		return "ShowdownStarted"
	case PotsCreated:
		return "PotsCreated"
	case WinnersDetermined:
		return "WinnersDetermined"
	case ChipsAwarded:
		return "ChipsAwarded"
	case PlayerBusted:
		return "PlayerBusted"
	case HandEnded:
		return "HandEnded"
	case GameEnded:
		return "GameEnded"
	case ErrorState:
		return "ErrorState"
	default:
		return "Unknown"
	}
}
