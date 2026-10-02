package pokergame

import "slices"

func (g *Game) emitGameStarted() {
	e := g.newBaseEvent(GameStarted)
	g.emit(e)
}

func (g *Game) emitHandStarted() {
	e := g.newBaseEvent(HandStarted)
	e.Board = slices.Clone(g.board)
	g.emit(e)
}

func (g *Game) emitBlindsAdvanced() {
	e := g.newBaseEvent(BlindsAdvanced)
	e.BlindIndices = &BlindIndices{
		Dealer:     g.dealerIndex,
		SmallBlind: g.smallBlindIndex,
		BigBlind:   g.bigBlindIndex,
	}
	g.emit(e)
}

func (g *Game) emitBlindsPosted() {
	e := g.newBaseEvent(BlindsPosted)
	e.BlindIndices = &BlindIndices{
		Dealer:     g.dealerIndex,
		SmallBlind: g.smallBlindIndex,
		BigBlind:   g.bigBlindIndex,
	}
	g.emit(e)
}

func (g *Game) emitHoleCardsDealt() {
	e := g.newBaseEvent(HoleCardsDealt)
	g.emit(e)
}

func (g *Game) emitHandEnded() {
	e := g.newBaseEvent(HandEnded)
	e.Board = slices.Clone(g.board)
	// pots?!!!
	g.emit(e)
}

func (g *Game) emitStreetStarted() {
	e := g.newBaseEvent(StreetStarted)
	e.Board = slices.Clone(g.board)
	e.Street = g.street
	g.emit(e)
}

func (g *Game) emitRunToShowdown() {
	e := g.newBaseEvent(RunToShowdown)
	g.emit(e)
}

func (g *Game) emitActionRequested(request ActionRequest) {
	e := g.newBaseEvent(ActionRequested)
	id := g.players[g.actingPlayerIndex].ID
	e.PlayerID = &id
	e.Board = slices.Clone(g.board)
	e.LegalActions = slices.Clone(request.LegalActions)
	e.Amount = &request.ToCall
	g.emit(e)
}

func (g *Game) emitActionInvalid(request ActionRequest, input Action) {
	e := g.newBaseEvent(ActionInvalid)
	id := g.players[g.actingPlayerIndex].ID
	e.PlayerID = &id
	e.Board = slices.Clone(g.board)
	e.LegalActions = slices.Clone(request.LegalActions)
	e.Amount = &request.ToCall
	a := input.Type
	e.ActionType = &a
	g.emit(e)
}

func (g *Game) emitActionValid(playerID string, input Action) {
	e := g.newBaseEvent(ActionValid)
	e.PlayerID = &playerID
	e.Board = slices.Clone(g.board)
	amount := input.Amount
	e.Amount = &amount
	actionType := input.Type
	e.ActionType = &actionType
	g.emit(e)
}

func (g *Game) emitOnePlayerLeft(winnerID string, amount int) {
	e := g.newBaseEvent(OnePlayerLeft)
	e.Board = slices.Clone(g.board)
	e.PlayerID = &winnerID
	e.Pots = &[]PotState{
		{EligiblePlayerIDs: []string{winnerID},
			Amount:    amount,
			WinnerIDs: []string{winnerID}},
	}
	g.emit(e)
}

func (g *Game) emitStreetEnded() {
	e := g.newBaseEvent(StreetEnded)
	e.Board = slices.Clone(g.board)
	g.emit(e)
}

func (g *Game) emitPotsCreated(pots []*Pot) {
	e := g.newBaseEvent(PotsCreated)
	e.Board = slices.Clone(g.board)
	eventPots := make([]PotState, 0, len(pots))
	for _, p := range pots {
		elig := make([]string, 0, len(p.EligiblePlayers))
		for _, player := range p.EligiblePlayers {
			elig = append(elig, player.ID)
		}
		eventPots = append(eventPots, PotState{
			EligiblePlayerIDs: elig,
			Amount:            p.Amount,
		})
	}
	e.Pots = &eventPots
	g.emit(e)
}

func (g *Game) emitWinnersDetermined(pots []*Pot) {
	e := g.newBaseEvent(WinnersDetermined)
	e.Board = slices.Clone(g.board)
	eventPots := make([]PotState, 0, len(pots))
	for _, p := range pots {
		eligIDs := make([]string, 0, len(p.EligiblePlayers))
		for _, player := range p.EligiblePlayers {
			eligIDs = append(eligIDs, player.ID)
		}
		winIDs := make([]string, 0, len(p.Winners))
		for _, w := range p.Winners {
			winIDs = append(winIDs, w.ID)
		}
		eventPots = append(eventPots, PotState{
			EligiblePlayerIDs: eligIDs,
			Amount:            p.Amount,
			WinnerIDs:         winIDs,
		})
	}
	e.Pots = &eventPots
	g.emit(e)
}

func (g *Game) emitChipsAwarded(pots []*Pot) {
	e := g.newBaseEvent(ChipsAwarded)
	e.Board = slices.Clone(g.board)
	eventPots := make([]PotState, 0, len(pots))
	for _, p := range pots {
		eligIDs := make([]string, 0, len(p.EligiblePlayers))
		for _, player := range p.EligiblePlayers {
			eligIDs = append(eligIDs, player.ID)
		}
		winIDs := make([]string, 0, len(p.Winners))
		for _, w := range p.Winners {
			winIDs = append(winIDs, w.ID)
		}
		eventPots = append(eventPots, PotState{
			EligiblePlayerIDs: eligIDs,
			Amount:            p.Amount,
			WinnerIDs:         winIDs,
		})
	}
	e.Pots = &eventPots
	g.emit(e)
}

func (g *Game) emitPlayerBusted(playerID string) {
	e := g.newBaseEvent(PlayerBusted)
	e.PlayerID = &playerID
	g.emit(e)
}

func (g *Game) emitPlayerLeft(playerID string) {
	e := g.newBaseEvent(PlayerLeft)
	e.PlayerID = &playerID
	g.emit(e)
}

func (g *Game) emitGameEnded(winnerID string) {
	e := g.newBaseEvent(GameEnded)
	e.PlayerID = &winnerID
	g.emit(e)
}

func (g *Game) emitErr(err error) {
	e := Event{
		Type:       ErrorState,
		HandNumber: g.handNumber,
		Street:     g.street,
		Players:    nil,
		Err:        err,
	}
	g.emit(e)
}
