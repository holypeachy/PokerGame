package pokergame

import (
	"fmt"
	"strings"

	"pokeralgo"
)

type BlindIndices struct {
	Dealer     int
	SmallBlind int
	BigBlind   int
}

type PotState struct {
	EligiblePlayerIDs []string
	Amount            int

	WinnerIDs []string
}

type Street int

const (
	NoStreet Street = iota
	Preflop
	Flop
	Turn
	River
)

func (s Street) String() string {
	switch s {
	case NoStreet:
		return "NoStreet"
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
	// Common
	Type       EventType
	HandNumber int
	Street     Street
	Players    []PlayerState

	BlindIndices *BlindIndices
	Board        []pokeralgo.Card
	PlayerID     *string
	LegalActions []ActionType
	Amount       *int
	ActionType   *ActionType
	Pots         *[]PotState

	Err error
}

func (e Event) String() string {
	var output strings.Builder
	fmt.Fprintf(&output, "%s HandNumber: %d", e.Type, e.HandNumber)
	fmt.Fprintf(&output, "\n   Street: %s", e.Street)
	fmt.Fprintf(&output, "\n   Players: %+v", e.Players)
	if e.BlindIndices != nil {
		fmt.Fprintf(&output, "\n   BlindIndices: %+v", *e.BlindIndices)
	}
	if e.Board != nil {
		fmt.Fprintf(&output, "\n   Board: %v", e.Board)
	}

	var actionDetails []string
	if e.PlayerID != nil {
		actionDetails = append(actionDetails, fmt.Sprintf("PlayerID: %s", *e.PlayerID))
	}
	if e.LegalActions != nil {
		actionDetails = append(actionDetails, fmt.Sprintf("LegalActions: %v", e.LegalActions))
	}
	if e.Amount != nil {
		actionDetails = append(actionDetails, fmt.Sprintf("Amount: %d", *e.Amount))
	}
	if e.ActionType != nil {
		actionDetails = append(actionDetails, fmt.Sprintf("Action: %+v", *e.ActionType))
	}
	if len(actionDetails) > 0 {
		fmt.Fprintf(&output, "\n   %s", strings.Join(actionDetails, " | "))
	}
	if e.Pots != nil {
		fmt.Fprintf(&output, "\n   Pots: %+v", *e.Pots)
	}
	if e.Err != nil {
		fmt.Fprintf(&output, "\n   Err: %v", e.Err)
	}
	fmt.Fprintf(&output, "\n")
	return output.String()
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
	PlayerLeft
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
	case PlayerLeft:
		return "PlayerLeft"
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
