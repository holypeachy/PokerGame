package pokergame

import (
	"fmt"

	pokeralgo "pokeralgo"
)

type Game struct {
	// Engine
	options      Options
	actionSource ActionSource

	// Table
	deck    *pokeralgo.Deck
	players []*Player

	// Hand
	Board []pokeralgo.Card

	dealerIndex     int
	smallBlindIndex int
	bigBlindIndex   int

	actingPlayerIndex int

	CurrentBet           int
	AdditionalRaiseCount int

	onePlayerLeft bool
	runToShowdown bool
}

func New(options Options, actionSource ActionSource) *Game {
	return &Game{
		options:      options,
		actionSource: actionSource,
		players:      []*Player{},
		deck:         pokeralgo.NewDeck(),
		Board:        make([]pokeralgo.Card, 0, 5),
	}
}

func (g *Game) InitActionSource(actionSource ActionSource) error {
	if g.actionSource != nil {
		return fmt.Errorf("%w: action source is already initialized, action source cannot be reassigned", ErrInternal)
	}
	g.actionSource = actionSource
	return nil
}

func (g *Game) SeatPlayers(playersInfo []PlayerSpec) error {
	if g.actionSource == nil {
		return fmt.Errorf("%w: action source has not been initialized", ErrInternal)
	}
	if len(playersInfo) < 2 {
		return fmt.Errorf("%w: Not enough players", ErrGame)
	}

	g.players = g.players[:0]
	g.deck.Reset()
	g.dealerIndex = -1

	for _, pi := range playersInfo {
		player, err := NewPlayerFromInfo(pi, g.options.BuyIn, g.deck.MustDraw(), g.deck.MustDraw())
		if err != nil {
			return err
		}
		g.players = append(g.players, player)
	}
	return nil
}

func (g *Game) PlayHand() error {
	if len(g.players) == 0 {
		return fmt.Errorf("%w: call SeatPlayers before PlayHand", ErrGame)
	}

	g.deck.Reset()
	g.Board = g.Board[:0]
	g.advanceBlinds()

	// Pre-flop
	fmt.Println("Pre-Flop")
	if err := g.players[g.smallBlindIndex].postBlind(g.options.BigBlind / 2); err != nil {
		return err
	}
	if err := g.players[g.bigBlindIndex].postBlind(g.options.BigBlind); err != nil {
		return err
	}
	g.CurrentBet = g.options.BigBlind

	for _, p := range g.players {
		if err := p.deal(g.deck.MustDraw(), g.deck.MustDraw()); err != nil {
			return err
		}
	}

	if err := g.runBettingRound(); err != nil {
		return err
	}
	g.actingPlayerIndex = g.nextPlayerIndex(g.dealerIndex)

	// Flop
	fmt.Println("Flop")
	g.deck.MustDraw()
	if !g.onePlayerLeft {
		g.Board = append(g.Board, g.deck.MustDrawN(3)...)
	}
	if err := g.runBettingRound(); err != nil {
		return err
	}
	g.actingPlayerIndex = g.nextPlayerIndex(g.dealerIndex)

	// Turn
	fmt.Println("Turn")
	g.deck.MustDraw()
	if !g.onePlayerLeft {
		g.Board = append(g.Board, g.deck.MustDraw())
	}
	if err := g.runBettingRound(); err != nil {
		return err
	}
	g.actingPlayerIndex = g.nextPlayerIndex(g.dealerIndex)

	// River
	fmt.Println("River")
	g.deck.MustDraw()
	if !g.onePlayerLeft {
		g.Board = append(g.Board, g.deck.MustDraw())
	}
	if err := g.runBettingRound(); err != nil {
		return err
	}

	if !g.onePlayerLeft {
		if err := g.resolveShowdown(); err != nil {
			return err
		}
	}

	g.resetForNextHand()
	return nil
}

func (g *Game) runBettingRound() error {
	if g.onePlayerLeft || g.runToShowdown {
		if g.runToShowdown {
			fmt.Println("Skip to showdown")
		}
		return nil
	}

	isBettingRoundOver := false
	for !isBettingRoundOver {
		for {
			currentPlayer := g.players[g.actingPlayerIndex]

			// Before
			if currentPlayer.Folded || isAllIn(currentPlayer) {
				g.actingPlayerIndex = g.nextPlayerIndex(g.actingPlayerIndex)
				continue
			} else if g.remainingPlayerCount() == 1 {
				g.onePlayerLeft = true
				isBettingRoundOver = true
				break
			} else if g.playersAbleToAct() < 2 && g.betsSettled() {
				g.runToShowdown = true
				isBettingRoundOver = true
				break
			} else if g.allPlayersActed() && g.betsSettled() {
				isBettingRoundOver = true
				break
			}

			// Action
			gameState := g.GameState()
			validMoves := gameState.LegalActions
			input, err := g.actionSource.NextAction(gameState)
			if err != nil {
				return err
			}
			if !containsAction(validMoves, input.Type) {
				return fmt.Errorf("%w: Input not valid", ErrGame)
			}

			if input.Type == Fold {
				if err := currentPlayer.Fold(); err != nil {
					return err
				}
				if g.remainingPlayerCount() == 1 {
					g.onePlayerLeft = true
					isBettingRoundOver = true
					break
				}
			} else if input.Type == Raise || input.Type == Call {
				if g.allPlayersActed() && input.Type == Raise {
					g.AdditionalRaiseCount++
				}

				if input.Type == Raise {
					toCall := g.CurrentBet - currentPlayer.Bet
					if input.Amount < toCall {
						if err := currentPlayer.commitChips(toCall + 10); err != nil {
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

			// After
			if g.playersAbleToAct() < 2 && g.betsSettled() {
				isBettingRoundOver = true
				g.runToShowdown = true
				break
			}

			// Move To Next Player
			g.actingPlayerIndex = g.nextPlayerIndex(g.actingPlayerIndex)
		}
	}

	if g.onePlayerLeft {
		fmt.Println("1 non-folded player remains")
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
	}
	g.resetBettingRound()
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
			if len(g.Board) != 5 {
				return fmt.Errorf("%w: community cards count is not 5", ErrInternal)
			}

			winners, err := pokeralgo.DetermineWinners(toAlgoPlayers(pot.EligiblePlayers), g.Board)
			if err != nil {
				return err
			}
			fmt.Println("Algo Winners for THIS pot")
			for _, item := range winners {
				fmt.Println(item)
				if item.BestHand != nil {
					fmt.Println(item.BestHand)
				}
			}
			fmt.Println()
			pot.Winners = g.MapAlgoPlayers(winners)
		}
	}
	fmt.Print("\n--- PAY ---\n\n")
	for _, item := range pots {
		fmt.Print("\nPot:\n")
		fmt.Println(item)
		if err := item.Distribute(); err != nil {
			return err
		}
	}
	return nil
}

func (g *Game) MapAlgoPlayers(players []pokeralgo.Player) []*Player {
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

	moves := []ActionType{}
	toCall := g.CurrentBet - player.Bet
	if toCall > 0 {
		moves = append(moves, Call)
		moves = append(moves, Fold)
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

func (g *Game) resetBettingRound() {
	for _, p := range g.players {
		p.resetForBettingRound()
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
}

func (g *Game) GameState() GameState {
	playerStates := []PlayerState{}
	for _, player := range g.players {
		holeCards := player.HoleCards
		playerStates = append(playerStates, PlayerState{ID: player.ID, Stack: player.Stack, HoleCards: &holeCards, Bet: player.Bet, HasFolded: player.Folded})
	}
	currentPlayer := &playerStates[g.actingPlayerIndex]
	if currentPlayer.HasFolded {
		panic("Current player cannot be folded")
	}
	toCall := g.CurrentBet - currentPlayer.Bet
	return GameState{
		PlayerStates:   playerStates,
		CommunityCards: append([]pokeralgo.Card(nil), g.Board...),
		OutputType:     InputRequest,
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
