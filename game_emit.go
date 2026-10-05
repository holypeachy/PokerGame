package pokergame

import "slices"

func (g *Game) emitGameStarted() {
	e := g.newBaseEvent(GameStarted)
	if e == nil {
		return
	}
	g.emit(e)
}

func (g *Game) emitHandStarted() {
	e := g.newBaseEvent(HandStarted)
	if e == nil {
		return
	}
	e.Board = slices.Clone(g.board)
	g.emit(e)
}

func (g *Game) emitBlindsAdvanced() {
	e := g.newBaseEvent(BlindsAdvanced)
	if e == nil {
		return
	}
	e.BlindIndices = &BlindIndices{
		Dealer:     g.dealerIndex,
		SmallBlind: g.smallBlindIndex,
		BigBlind:   g.bigBlindIndex,
	}
	g.emit(e)
}

func (g *Game) emitBlindsPosted() {
	e := g.newBaseEvent(BlindsPosted)
	if e == nil {
		return
	}
	e.BlindIndices = &BlindIndices{
		Dealer:     g.dealerIndex,
		SmallBlind: g.smallBlindIndex,
		BigBlind:   g.bigBlindIndex,
	}
	g.emit(e)
}

func (g *Game) emitHoleCardsDealt() {
	e := g.newBaseEvent(HoleCardsDealt)
	if e == nil {
		return
	}
	g.emit(e)
}

func (g *Game) emitHandEnded() {
	e := g.newBaseEvent(HandEnded)
	if e == nil {
		return
	}
	e.Board = slices.Clone(g.board)
	// pots?!!!
	g.emit(e)
}

func (g *Game) emitStreetStarted() {
	e := g.newBaseEvent(StreetStarted)
	if e == nil {
		return
	}
	e.Board = slices.Clone(g.board)
	e.Street = g.street
	g.emit(e)
}

func (g *Game) emitRunToShowdown() {
	e := g.newBaseEvent(RunToShowdown)
	if e == nil {
		return
	}
	g.emit(e)
}

func (g *Game) emitActionRequested(request ActionRequest) {
	e := g.newBaseEvent(ActionRequested)
	if e == nil {
		return
	}
	id := g.players[g.actingPlayerIndex].ID
	e.PlayerID = &id
	e.Board = slices.Clone(g.board)
	e.LegalActions = slices.Clone(request.LegalActions)
	e.Amount = &request.ToCall
	g.emit(e)
}

func (g *Game) emitActionInvalid(request ActionRequest, input Action) {
	e := g.newBaseEvent(ActionInvalid)
	if e == nil {
		return
	}
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
	if e == nil {
		return
	}
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
	if e == nil {
		return
	}
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
	if e == nil {
		return
	}
	e.Board = slices.Clone(g.board)
	g.emit(e)
}

func (g *Game) createPotEvent(t EventType, pots []*Pot) *Event {
	e := g.newBaseEvent(t)
	if e == nil {
		return nil
	}
	e.Board = slices.Clone(g.board)
	eventPots := make([]PotState, 0, len(pots))
	for _, p := range pots {
		eligIDs := make([]string, 0, len(p.EligiblePlayers))
		for _, player := range p.EligiblePlayers {
			eligIDs = append(eligIDs, player.ID)
		}
		var winIDs []string
		if t != PotsCreated {
			winIDs = make([]string, 0, len(p.Winners))
			for _, w := range p.Winners {
				winIDs = append(winIDs, w.ID)
			}
		}
		eventPots = append(eventPots, PotState{
			EligiblePlayerIDs: eligIDs,
			Amount:            p.Amount,
			WinnerIDs:         winIDs,
		})
	}
	e.Pots = &eventPots
	return e
}

func (g *Game) emitPotsCreated(pots []*Pot) {
	e := g.createPotEvent(PotsCreated, pots)
	if e == nil {
		return
	}
	g.emit(e)
}

func (g *Game) emitWinnersDetermined(pots []*Pot) {
	e := g.createPotEvent(WinnersDetermined, pots)
	if e == nil {
		return
	}
	g.emit(e)
}

func (g *Game) emitChipsAwarded(pots []*Pot) {
	e := g.createPotEvent(ChipsAwarded, pots)
	if e == nil {
		return
	}
	g.emit(e)
}

func (g *Game) emitChipsOnePlayerLeft(winnerID string, amount int) {
	e := g.newBaseEvent(ChipsAwarded)
	if e == nil {
		return
	}
	e.Board = slices.Clone(g.board)
	e.Pots = &[]PotState{PotState{
		EligiblePlayerIDs: []string{winnerID},
		Amount:            amount,
		WinnerIDs:         []string{winnerID},
	}}
	g.emit(e)
}

func (g *Game) emitPlayerBusted(playerID string) {
	e := g.newBaseEvent(PlayerBusted)
	if e == nil {
		return
	}
	e.PlayerID = &playerID
	g.emit(e)
}

func (g *Game) emitPlayerLeft(playerID string) {
	e := g.newBaseEvent(PlayerLeft)
	if e == nil {
		return
	}
	e.PlayerID = &playerID
	g.emit(e)
}

func (g *Game) emitGameEnded(winnerID string) {
	e := g.newBaseEvent(GameEnded)
	if e == nil {
		return
	}
	e.PlayerID = &winnerID
	g.emit(e)
}

func (g *Game) emitErr(err error) {
	e := &Event{
		Type:       ErrorState,
		HandNumber: g.handNumber,
		Street:     g.street,
		Players:    nil,
		Err:        err,
	}
	g.emit(e)
}
