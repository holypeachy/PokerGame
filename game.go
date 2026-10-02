package pokergame

import (
	"fmt"
	"slices"

	pokeralgo "pokeralgo"
)

type Game struct {
	// Engine
	options      Options
	actionSource ActionSource
	onEvent      func(Event)

	// Table
	deck    *pokeralgo.Deck
	players []*Player

	// Hand
	handNumber int
	street     Street
	board      []pokeralgo.Card

	dealerIndex     int
	smallBlindIndex int
	bigBlindIndex   int

	actingPlayerIndex int

	CurrentBet           int
	AdditionalRaiseCount int

	onePlayerLeft bool
	runToShowdown bool
}

func New(options Options, actionSource ActionSource, onEvent func(Event)) *Game {
	return &Game{
		options:      options,
		actionSource: actionSource,
		onEvent:      onEvent,
		players:      []*Player{},
		deck:         pokeralgo.NewDeck(),
		board:        make([]pokeralgo.Card, 0, 5),
	}
}

func (g *Game) SeatPlayers(playerSpecs []PlayerSpec) error {
	if g.actionSource == nil {
		return fmt.Errorf("%w: action source has not been initialized", ErrInternal)
	}
	if len(playerSpecs) < 2 {
		return fmt.Errorf("%w: Not enough players", ErrGame)
	}

	g.players = g.players[:0]
	g.deck.Reset()
	g.dealerIndex = -1
	g.handNumber = 0

	for _, pi := range playerSpecs {
		player, err := NewPlayerFromSpec(pi, g.options.BuyIn, pokeralgo.Card{Rank: 0, Suit: pokeralgo.Hearts, IsHoleCard: true}, pokeralgo.Card{Rank: 0, Suit: pokeralgo.Spades, IsHoleCard: true})
		if err != nil {
			return err
		}
		g.players = append(g.players, player)
	}

	e := g.newEvent(GameStarted)
	g.emit(e)

	return nil
}

func (g *Game) PlayHand() error {
	if len(g.players) == 0 {
		return fmt.Errorf("%w: call SeatPlayers before PlayHand", ErrGame)
	}

	g.handNumber++
	g.deck.Reset()
	g.board = g.board[:0]

	e := g.newEvent(HandStarted)
	e.Board = slices.Clone(g.board)
	g.emit(e)

	g.advanceBlinds()
	e = g.newEvent(BlindsAdvanced)
	e.BlindIndices = &BlindIndices{
		Dealer:     g.dealerIndex,
		SmallBlind: g.smallBlindIndex,
		BigBlind:   g.bigBlindIndex,
	}
	g.emit(e)

	// Pre-flop
	if err := g.players[g.smallBlindIndex].postBlind(g.options.BigBlind / 2); err != nil {
		return err
	}
	if err := g.players[g.bigBlindIndex].postBlind(g.options.BigBlind); err != nil {
		return err
	}
	g.CurrentBet = g.options.BigBlind

	e = g.newEvent(BlindsPosted)
	e.BlindIndices = &BlindIndices{
		Dealer:     g.dealerIndex,
		SmallBlind: g.smallBlindIndex,
		BigBlind:   g.bigBlindIndex,
	}
	g.emit(e)

	for _, p := range g.players {
		if err := p.deal(g.deck.MustDraw(), g.deck.MustDraw()); err != nil {
			return err
		}
	}
	e = g.newEvent(HoleCardsDealt)
	g.emit(e)

	g.street = Preflop
	if err := g.runStreet(); err != nil {
		return err
	}
	g.actingPlayerIndex = g.nextPlayerIndex(g.dealerIndex)

	// Flop
	g.deck.MustDraw()
	if !g.onePlayerLeft {
		g.street = Flop
		g.board = append(g.board, g.deck.MustDrawN(3)...)
	}
	if err := g.runStreet(); err != nil {
		return err
	}
	g.actingPlayerIndex = g.nextPlayerIndex(g.dealerIndex)

	// Turn
	g.deck.MustDraw()
	if !g.onePlayerLeft {
		g.street = Turn
		g.board = append(g.board, g.deck.MustDraw())
	}
	if err := g.runStreet(); err != nil {
		return err
	}
	g.actingPlayerIndex = g.nextPlayerIndex(g.dealerIndex)

	// River
	g.deck.MustDraw()
	if !g.onePlayerLeft {
		g.street = River
		g.board = append(g.board, g.deck.MustDraw())
	}
	if err := g.runStreet(); err != nil {
		return err
	}

	if !g.onePlayerLeft {
		if err := g.resolveShowdown(); err != nil {
			return err
		}
	}

	e = g.newEvent(HandEnded)
	e.Board = slices.Clone(g.board)
	// pots?!!!
	g.emit(e)

	g.resetForNextHand()
	return nil
}

func (g *Game) runStreet() error {
	if g.onePlayerLeft || g.runToShowdown {
		if g.runToShowdown {
			fmt.Println("Skip to showdown")
			e := g.newEvent(StreetStarted)
			e.Board = slices.Clone(g.board)
			s := g.street
			e.Street = &s
			g.emit(e)
		}
		return nil
	}
	e := g.newEvent(StreetStarted)
	e.Board = slices.Clone(g.board)
	s := g.street
	e.Street = &s
	g.emit(e)

	isStreetOver := false
	for !isStreetOver {
		for {
			currentPlayer := g.players[g.actingPlayerIndex]

			// Before
			if g.remainingPlayerCount() == 1 {
				g.onePlayerLeft = true
				isStreetOver = true
				break
			} else if g.playersAbleToAct() < 2 && g.betsSettled() {
				g.runToShowdown = true
				isStreetOver = true
				e := g.newEvent(RunToShowdown)
				g.emit(e)
				break
			} else if currentPlayer.Folded || isAllIn(currentPlayer) {
				g.actingPlayerIndex = g.nextPlayerIndex(g.actingPlayerIndex)
				continue
			} else if g.allPlayersActed() && g.betsSettled() {
				isStreetOver = true
				break
			}

			// Action
			actionRequest := g.actionRequest()
			validMoves := actionRequest.LegalActions

			e := g.newEvent(ActionRequested)
			id := g.players[g.actingPlayerIndex].ID
			e.PlayerID = &id
			e.Board = slices.Clone(g.board)
			e.LegalActions = slices.Clone(validMoves)
			e.Amount = &actionRequest.ToCall
			g.emit(e)

			input, err := g.actionSource.NextAction(actionRequest)
			if err != nil {
				return err
			}
			if !containsAction(validMoves, input.Type) {

				e := g.newEvent(ActionInvalid)
				id := g.players[g.actingPlayerIndex].ID
				e.PlayerID = &id
				e.Board = slices.Clone(g.board)
				e.LegalActions = slices.Clone(validMoves)
				e.Amount = &actionRequest.ToCall
				a := input.Type
				e.ActionType = &a
				g.emit(e)

				return fmt.Errorf("%w: Input not valid", ErrGame)
			}

			// After
			if input.Type == Fold {
				if err := currentPlayer.Fold(); err != nil {
					return err
				}
				if g.remainingPlayerCount() == 1 {
					g.onePlayerLeft = true
					isStreetOver = true
					e := g.newEvent(ActionValid)
					id := currentPlayer.ID
					e.PlayerID = &id
					e.Board = slices.Clone(g.board)
					amount := input.Amount
					e.Amount = &amount
					actionType := input.Type
					e.ActionType = &actionType
					g.emit(e)
					break
				}
			} else if input.Type == Raise || input.Type == Call {
				if g.allPlayersActed() && input.Type == Raise {
					g.AdditionalRaiseCount++
				}

				if input.Type == Raise {
					toCall := g.CurrentBet - currentPlayer.Bet
					if input.Amount < toCall {
						if err := currentPlayer.commitChips(toCall + 10); err != nil { // magic number warning
							return err
						}
					} else {
						if err := currentPlayer.commitChips(input.Amount); err != nil {
							return err
						}
					}
				} else {
					if err := currentPlayer.commitChips(g.CurrentBet - currentPlayer.Bet); err != nil {
						return err
					}
				}

				if currentPlayer.Bet > g.CurrentBet {
					g.CurrentBet = currentPlayer.Bet
				}
			} else if input.Type == Check {
				currentPlayer.Check()
			}

			e = g.newEvent(ActionValid)
			id2 := g.players[g.actingPlayerIndex].ID
			e.PlayerID = &id2
			e.Board = slices.Clone(g.board)
			am := input.Amount
			e.Amount = &am
			a := input.Type
			e.ActionType = &a
			g.emit(e)

			if g.playersAbleToAct() < 2 && g.betsSettled() {
				isStreetOver = true
				g.runToShowdown = true
				e := g.newEvent(RunToShowdown)
				g.emit(e)
				break
			}

			// Move To Next Player
			g.actingPlayerIndex = g.nextPlayerIndex(g.actingPlayerIndex)
		}
	}

	if g.onePlayerLeft {
		// 1 non-folded player remains
		winner, err := g.remainingPlayer()
		if err != nil {
			return err
		}
		pot := 0
		for _, p := range g.players {
			pot += p.Bet
		}
		if err := winner.credit(pot); err != nil {
			return err
		}

		if g.remainingPlayerCount() != 1 {
			return fmt.Errorf("%w: 1 non-folded player remains but count is not 1", ErrInternal)
		}

		e = g.newEvent(OnePlayerLeft)
		e.Board = slices.Clone(g.board)
		id := winner.ID
		e.PlayerID = &id
		e.Pots = &[]PotState{
			{EligiblePlayerIDs: []string{winner.ID},
				Amount:    pot,
				WinnerIDs: []string{winner.ID}},
		}
		g.emit(e)
	}
	e = g.newEvent(StreetEnded)
	e.Board = slices.Clone(g.board)
	g.emit(e)

	g.resetStreet()
	return nil
}

func (g *Game) resolveShowdown() error {
	if g.onePlayerLeft {
		fmt.Println("No Showdown. Only one player left")
		return nil
	}

	pots, err := buildPots(g.players)
	if err != nil {
		return err
	}

	e := g.newEvent(PotsCreated)
	e.Board = slices.Clone(g.board)
	eventPots := make([]PotState, 0, len(pots))
	for _, p := range pots {
		elig := make([]string, 0, len(p.EligiblePlayers))
		for _, e := range p.EligiblePlayers {
			elig = append(elig, e.ID)
		}
		eventPots = append(eventPots, PotState{
			EligiblePlayerIDs: elig,
			Amount:            p.Amount,
		})
	}
	e.Pots = &eventPots
	g.emit(e)

	algoPlayers := toAlgoPlayers(g.players)
	fmt.Println("All Algo Players")
	for _, item := range algoPlayers {
		fmt.Println(item)
	}
	fmt.Println()
	for _, pot := range pots {
		if len(pot.EligiblePlayers) == 1 {
			pot.Winners = pot.EligiblePlayers
		} else if len(pot.EligiblePlayers) < 1 {
			return fmt.Errorf("%w: pot has no players", ErrInternal)
		} else {
			if len(g.board) != 5 {
				return fmt.Errorf("%w: community cards count is not 5", ErrInternal)
			}

			winners, err := pokeralgo.DetermineWinners(toAlgoPlayers(pot.EligiblePlayers), g.board)
			if err != nil {
				return err
			}

			pot.Winners = g.mapAlgoPlayers(winners)
		}
	}
	e = g.newEvent(WinnersDetermined)
	e.Board = slices.Clone(g.board)
	eventPots = make([]PotState, 0, len(pots))
	for _, p := range pots {
		eligIDs := make([]string, 0, len(p.EligiblePlayers))
		for _, e := range p.EligiblePlayers {
			eligIDs = append(eligIDs, e.ID)
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

	fmt.Print("\n--- PAY ---\n\n")
	for _, item := range pots {
		if err := item.Distribute(); err != nil {
			return err
		}
	}

	e = g.newEvent(ChipsAwarded)
	e.Board = slices.Clone(g.board)
	eventPots = make([]PotState, 0, len(pots))
	for _, p := range pots {
		eligIDs := make([]string, 0, len(p.EligiblePlayers))
		for _, e := range p.EligiblePlayers {
			eligIDs = append(eligIDs, e.ID)
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
	return nil
}

func (g *Game) resetStreet() {
	for _, p := range g.players {
		p.resetForStreet()
	}
	g.AdditionalRaiseCount = 0
}

func (g *Game) resetForNextHand() {
	for _, p := range g.players {
		p.resetForNextHand()
	}
	g.AdditionalRaiseCount = 0
	g.onePlayerLeft = false
	g.runToShowdown = false
	g.street = NoStreet
}
