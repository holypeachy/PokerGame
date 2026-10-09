package pokergame

import (
	"fmt"
	"pokeralgo"
	"slices"
)

type EventVerbosity uint8

const (
	EventsOff  EventVerbosity = iota // Errors only
	EventsCore                       // Primary Game Info
	EventsAll                        // Every Step Logged
)

var eventLevelLookup = map[EventType]EventVerbosity{
	GameStarted:       EventsCore,
	HandStarted:       EventsAll,
	BlindsAdvanced:    EventsAll,
	BlindsPosted:      EventsCore,
	HoleCardsDealt:    EventsAll,
	StreetStarted:     EventsCore,
	ActionRequested:   EventsCore,
	ActionValid:       EventsCore,
	ActionInvalid:     EventsCore,
	StreetEnded:       EventsAll,
	OnePlayerLeft:     EventsAll,
	RunToShowdown:     EventsAll,
	ShowdownStarted:   EventsAll,
	PotsCreated:       EventsAll,
	WinnersDetermined: EventsAll,
	ChipsAwarded:      EventsCore,
	HandEnded:         EventsAll,
	PlayerBusted:      EventsCore,
	PlayerLeft:        EventsCore,
	GameEnded:         EventsCore,
}

func (g *Game) mapAlgoPlayers(players []pokeralgo.Player) []*player {
	enginePlayers := slices.Clone(g.players)
	filtered := []*player{}
	for _, p := range enginePlayers {
		for _, p2 := range players {
			if p2.Name == p.ID {
				filtered = append(filtered, p)
				break
			}
		}
	}
	if len(filtered) == 0 {
		panic("No PokerAlgo players match any engine players")
	} else if len(filtered) != len(players) {
		panic("len filtered and players don't match")
	}
	return filtered
}

func toAlgoPlayers(enginePlayers []*player) []pokeralgo.Player {
	players := []pokeralgo.Player{}
	for _, ep := range enginePlayers {
		players = append(players, pokeralgo.NewPlayer(ep.ID, ep.HoleCards.First, ep.HoleCards.Second))
	}
	return players
}

func (g *Game) legalActions(player *player) []ActionType {
	if player.Folded {
		panic("Cannot get possible moves for folded player")
	}

	moves := []ActionType{Fold}
	toCall := g.currentBet - player.Bet
	if toCall > 0 {
		moves = append(moves, Call)
	} else if toCall == 0 {
		moves = append(moves, Check)
	}
	// we count players that can act because we don't want to raise if another player is all-in
	if g.additionalRaiseCount != g.options.AdditionalRaises && g.playersAbleToAct() > 1 && toCall < player.Stack {
		moves = append(moves, Raise)
	}
	moves = append(moves, Leave)
	return moves
}

func (g *Game) remainingPlayer() (*player, error) {
	for _, p := range g.players {
		if !p.Folded {
			return p, nil
		}
	}
	return nil, fmt.Errorf("%w: No non-folded player", ErrInternal)
}

func (g *Game) betsSettled() bool {
	for _, p := range g.players {
		if !p.Folded && !isAllIn(p) && g.currentBet != p.Bet {
			return false
		}
	}
	return true
}

func (g *Game) allPlayersActed() bool {
	for _, p := range g.players {
		if !p.Folded && !p.Acted && !isAllIn(p) {
			return false
		}
	}
	return true
}

func (g *Game) playersAbleToAct() int {
	count := 0
	for _, p := range g.players {
		if !p.Folded && !isAllIn(p) {
			count++
		}
	}
	return count
}

func (g *Game) remainingPlayerCount() int {
	count := 0
	for _, p := range g.players {
		if !p.Folded {
			count++
		}
	}
	return count
}

func isAllIn(player *player) bool {
	if player.Stack == 0 && player.Bet == 0 {
		panic(fmt.Errorf("%w: Player has 0 stack and 0 bet. Busted players should never be in play.", ErrInternal))
	}

	if player.Stack == 0 && player.Bet > 0 {
		return true
	}
	return false
}

func (g *Game) actionRequest() ActionRequest {
	playerStates := []PlayerState{}
	for _, player := range g.players {
		holeCards := player.HoleCards
		playerStates = append(playerStates, PlayerState{ID: player.ID, Stack: player.Stack, HoleCards: holeCards, Bet: player.Bet, Folded: player.Folded, Left: player.Left})
	}
	currentPlayer := &playerStates[g.actingPlayerIndex]
	if currentPlayer.Folded {
		panic("Current player cannot be folded")
	}
	toCall := g.currentBet - currentPlayer.Bet
	return ActionRequest{
		PlayerStates:   playerStates,
		CommunityCards: slices.Clone(g.board),
		PlayerToAct:    currentPlayer,
		LegalActions:   g.legalActions(g.players[g.actingPlayerIndex]),
		ToCall:         toCall,
	}
}

func (g *Game) advanceBlinds() {
	if len(g.players) == 2 {
		g.dealerIndex = g.nextPlayerIndex(g.dealerIndex)
		g.smallBlindIndex = g.dealerIndex
		g.bigBlindIndex = g.nextPlayerIndex(g.dealerIndex)
		g.actingPlayerIndex = g.dealerIndex
	} else {
		g.dealerIndex = g.nextPlayerIndex(g.dealerIndex)
		g.smallBlindIndex = g.nextPlayerIndex(g.dealerIndex)
		g.bigBlindIndex = g.nextPlayerIndex(g.smallBlindIndex)
		g.actingPlayerIndex = g.nextPlayerIndex(g.bigBlindIndex)
	}
}

func (g *Game) adjustDealer(removed int) {
	g.dealerIndex -= removed
}

func (g *Game) nextPlayerIndex(index int) int {
	temp := index + 1
	if temp > len(g.players)-1 {
		temp = 0
	}
	return temp
}

func containsAction(moves []ActionType, target ActionType) bool {
	return slices.Contains(moves, target)
}

func (g *Game) checkGameOver() error {
	if len(g.players) < 2 {
		if len(g.players) != 1 {
			return fmt.Errorf("%w: No Winner, invariant violation", ErrInternal)
		}
		g.emitGameEnded(g.players[0].ID)
		g.gameOver = true
	}
	return nil
}

func (g *Game) newBaseEvent(t EventType) *Event {
	if eventLevelLookup[t] > g.options.EventVerbosity {
		return nil
	}
	players := make([]PlayerState, 0, len(g.players))
	for _, p := range g.players {
		players = append(players, PlayerState{
			ID:        p.ID,
			Stack:     p.Stack,
			Folded:    p.Folded,
			Left:      p.Left,
			HoleCards: p.HoleCards,
			Bet:       p.Bet,
		})
	}

	return &Event{
		Type:       t,
		HandNumber: g.handNumber,
		Street:     g.street,
		Players:    players,
	}
}

func (g *Game) emit(e *Event) {
	if g.onEvent != nil {
		g.onEvent(*e)
	}
}

func (g *Game) SetEventsVerbosity(lvl EventVerbosity) {
	g.options.EventVerbosity = lvl
}
