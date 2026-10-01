package pokergame

import pokeralgo "pokeralgo"

type PlayerSpec struct {
	ID string
}

func NewPlayerSpec(id string) PlayerSpec {
	return PlayerSpec{ID: id}
}

type ActionRequest struct {
	PlayerStates   []PlayerState
	CommunityCards []pokeralgo.Card

	PlayerToAct  *PlayerState
	LegalActions []ActionType
	ToCall       int
}

type Action struct {
	Type   ActionType
	Amount int
}

type PlayerState struct {
	ID        string
	Stack     int
	Folded    bool
	HoleCards pokeralgo.HoleCards
	Bet       int
}

type Options struct {
	BuyIn            int
	BigBlind         int
	AdditionalRaises int
	EnableDebug      bool
	DebugVerbosity   DebugLevel
}

type ActionSource interface {
	NextAction(gameState ActionRequest) (Action, error)
}
