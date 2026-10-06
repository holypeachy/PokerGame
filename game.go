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

	// Street
	street Street
	board  []pokeralgo.Card

	dealerIndex     int
	smallBlindIndex int
	bigBlindIndex   int

	actingPlayerIndex int

	CurrentBet           int
	AdditionalRaiseCount int

	// Control Flow
	onePlayerLeft bool
	runToShowdown bool
	gameOver      bool
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

func (g *Game) Play(playerSpecs []PlayerSpec) error {
	if err := g.seatPlayers(playerSpecs); err != nil {
		return err
	}

	for !g.gameOver {
		if err := g.playHand(); err != nil {
			return err
		}
	}

	return nil
}

func (g *Game) seatPlayers(playerSpecs []PlayerSpec) error {
	err := g.seatPlayersInternal(playerSpecs)
	if err != nil {
		g.emitErr(err)
	}
	return err
}

func (g *Game) seatPlayersInternal(playerSpecs []PlayerSpec) error {
	if g.actionSource == nil {
		return fmt.Errorf("%w: action source has not been initialized", ErrInternal)
	} else if len(playerSpecs) < 2 {
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

	g.emitGameStarted()

	return nil
}

func (g *Game) playHand() error {
	err := g.playHandInternal()
	if err != nil {
		g.emitErr(err)
	}
	return err
}

func (g *Game) playHandInternal() error {
	if len(g.players) == 0 {
		return fmt.Errorf("%w: call SeatPlayers before PlayHand", ErrGame)
	}

	g.handNumber++
	g.deck.Reset()
	g.board = g.board[:0]

	g.emitHandStarted()

	g.advanceBlinds()
	g.emitBlindsAdvanced()

	// Pre-flop
	if err := g.players[g.smallBlindIndex].postBlind(g.options.BigBlind / 2); err != nil {
		return err
	}
	if err := g.players[g.bigBlindIndex].postBlind(g.options.BigBlind); err != nil {
		return err
	}
	g.CurrentBet = g.options.BigBlind

	g.emitBlindsPosted()

	for _, p := range g.players {
		if err := p.deal(g.deck.MustDraw(), g.deck.MustDraw()); err != nil {
			return err
		}
	}

	g.emitHoleCardsDealt()

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

	g.emitHandEnded()

	g.removeBustedLeft()
	g.resetForNextHand()
	if err := g.checkGameOver(); err != nil {
		return err
	}
	return nil
}

func (g *Game) runStreet() error {
	if g.onePlayerLeft || g.runToShowdown {
		if g.runToShowdown {
			g.emitStreetStarted()
		}
		return nil
	}
	g.emitStreetStarted()

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
				g.emitRunToShowdown()
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

			g.emitActionRequested(actionRequest)

			input, err := g.actionSource.NextAction(actionRequest)
			if err != nil {
				return err
			}
			if !containsAction(validMoves, input.Type) {
				g.emitActionInvalid(actionRequest, input)

				return fmt.Errorf("%w: Input not valid", ErrGame)
			}

			// After
			if input.Type == Leave {
				if err := currentPlayer.Leave(); err != nil {
					return err
				}
				// emit in fold
			}

			if input.Type == Fold || input.Type == Leave {
				if err := currentPlayer.Fold(); err != nil {
					return err
				}
				if g.remainingPlayerCount() == 1 {
					g.onePlayerLeft = true
					isStreetOver = true
					g.emitActionValid(currentPlayer.ID, input) // we emit cus break
					// no need emit one player left
					break
				}
			} else if input.Type == Raise || input.Type == Call {
				if g.allPlayersActed() && input.Type == Raise {
					g.AdditionalRaiseCount++
				}

				if input.Type == Raise {
					toCall := g.CurrentBet - currentPlayer.Bet
					if input.Amount <= toCall {
						if err := currentPlayer.commitChips(toCall); err != nil { // I had toCall + 10, I'll just call
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

			g.emitActionValid(g.players[g.actingPlayerIndex].ID, input)

			if g.playersAbleToAct() < 2 && g.betsSettled() {
				isStreetOver = true
				g.runToShowdown = true
				g.emitRunToShowdown()
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

		g.emitOnePlayerLeft(winner.ID, pot)
		g.emitChipsOnePlayerLeft(winner.ID, pot)
	}
	g.emitStreetEnded()

	g.resetStreet()
	return nil
}

func (g *Game) resolveShowdown() error {
	if g.onePlayerLeft {
		return nil
	}

	pots, err := buildPots(g.players)
	if err != nil {
		return err
	}

	g.emitPotsCreated(pots)

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
	g.emitWinnersDetermined(pots)

	for _, item := range pots {
		if err := item.Distribute(); err != nil {
			return err
		}
	}

	g.emitChipsAwarded(pots)
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
		p.HoleCards = pokeralgo.HoleCards{
			First: pokeralgo.Card{
				Rank: 0,
				Suit: pokeralgo.Spades,
			},
			Second: pokeralgo.Card{
				Rank: 0,
				Suit: pokeralgo.Hearts,
			},
		}
	}
	g.AdditionalRaiseCount = 0
	g.onePlayerLeft = false
	g.runToShowdown = false
	g.street = NoStreet
}

func (g *Game) removeBustedLeft() {
	newPlayers := make([]*Player, 0)
	removedBeforeDealer := 0
	indexToRemove := 0
	for _, p := range g.players {
		if p.Stack == 0 {
			indexToRemove = slices.IndexFunc(g.players, func(current *Player) bool {
				if current.ID == p.ID {
					return true
				}
				return false
			})
			if indexToRemove <= g.dealerIndex {
				removedBeforeDealer++
			}
			g.emitPlayerBusted(p.ID)
		} else if p.Left {
			indexToRemove = slices.IndexFunc(g.players, func(current *Player) bool {
				if current.ID == p.ID {
					return true
				}
				return false
			})
			if indexToRemove <= g.dealerIndex {
				removedBeforeDealer++
			}
			g.emitPlayerLeft(p.ID)
		} else {
			newPlayers = append(newPlayers, p)
		}
	}
	g.adjustDealer(removedBeforeDealer)
	g.players = newPlayers
}
