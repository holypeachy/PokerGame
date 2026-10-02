package pokergame

import (
	"fmt"
	"pokeralgo"
	"slices"
)

func (g *Game) mapAlgoPlayers(players []pokeralgo.Player) []*Player {
	enginePlayers := append([]*Player(nil), g.players...)
	filtered := []*Player{}
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

func toAlgoPlayers(enginePlayers []*Player) []pokeralgo.Player {
	players := []pokeralgo.Player{}
	for _, ep := range enginePlayers {
		players = append(players, pokeralgo.NewPlayer(ep.ID, ep.HoleCards.First, ep.HoleCards.Second))
	}
	return players
}

func (g *Game) legalActions(player *Player) []ActionType {
	if player.Folded {
		panic("Cannot get possible moves for folded player")
	}

	moves := []ActionType{Fold}
	toCall := g.CurrentBet - player.Bet
	if toCall > 0 {
		moves = append(moves, Call)
	} else if toCall == 0 {
		moves = append(moves, Check)
	}
	// we count players that can act because we don't want to raise if another player is all-in
	if g.AdditionalRaiseCount != g.options.AdditionalRaises && g.playersAbleToAct() > 1 {
		moves = append(moves, Raise)
	}
	return moves
}

func (g *Game) remainingPlayer() (*Player, error) {
	for _, p := range g.players {
		if !p.Folded {
			return p, nil
		}
	}
	return nil, fmt.Errorf("%w: No non-folded player", ErrInternal)
}

func (g *Game) betsSettled() bool {
	for _, p := range g.players {
		if !p.Folded && !isAllIn(p) && g.CurrentBet != p.Bet {
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

func isAllIn(player *Player) bool {
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
		playerStates = append(playerStates, PlayerState{ID: player.ID, Stack: player.Stack, HoleCards: holeCards, Bet: player.Bet, Folded: player.Folded})
	}
	currentPlayer := &playerStates[g.actingPlayerIndex]
	if currentPlayer.Folded {
		panic("Current player cannot be folded")
	}
	toCall := g.CurrentBet - currentPlayer.Bet
	return ActionRequest{
		PlayerStates:   playerStates,
		CommunityCards: slices.Clone(g.board),
		PlayerToAct:    currentPlayer,
		LegalActions:   g.legalActions(g.players[g.actingPlayerIndex]),
		ToCall:         toCall,
	}
}

func (g *Game) advanceBlinds() {
	g.dealerIndex = g.nextPlayerIndex(g.dealerIndex)
	g.smallBlindIndex = g.nextPlayerIndex(g.dealerIndex)
	g.bigBlindIndex = g.nextPlayerIndex(g.smallBlindIndex)
	g.actingPlayerIndex = g.nextPlayerIndex(g.bigBlindIndex)
}

func (g *Game) nextPlayerIndex(index int) int {
	temp := index + 1
	if temp > len(g.players)-1 {
		temp = 0
	}
	return temp
}

func containsAction(moves []ActionType, target ActionType) bool {
	for _, move := range moves {
		if move == target {
			return true
		}
	}
	return false
}

func (g *Game) newEvent(t EventType) Event {
	players := make([]PlayerState, 0, len(g.players))
	for _, p := range g.players {
		players = append(players, PlayerState{
			ID:        p.ID,
			Stack:     p.Stack,
			Folded:    p.Folded,
			HoleCards: p.HoleCards,
			Bet:       p.Bet,
		})
	}

	s := g.street
	return Event{
		Type:       t,
		HandNumber: g.handNumber,
		Street:     &s,
		Players:    players,
	}
}

func (g *Game) emit(e Event) {
	if g.onEvent != nil {
		g.onEvent(e)
	}
}
