package pokergame

import pokeralgo "pokeralgo"

type PlayerInfoDto struct {
	Id string
}

func NewPlayerInfoDto(id string) PlayerInfoDto {
	return PlayerInfoDto{Id: id}
}

type GameState struct {
	PlayerStates   []PlayerState
	CommunityCards []pokeralgo.Card
	OutputType     OutputType

	PlayerToAct   *PlayerState
	PossibleMoves []PlayerMove
	ToCall        int
}

type PlayerAction struct {
	Move   PlayerMove
	Amount int
}

type PlayerState struct {
	Id        string
	Stack     int
	HasFolded bool
	HoleCards *pokeralgo.Pair
	Bet       int
}

type PokerEngineOptions struct {
	BuyIn            int
	BigBlind         int
	AdditionalRaises int
	EnableDebug      bool
	DebugVerbosity   DebugVerbosity
}

type ActionSource interface {
	NextAction(gameState GameState) (PlayerAction, error)
}
