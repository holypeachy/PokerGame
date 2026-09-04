package pokergame

import (
	"fmt"

	pokeralgo "pokeralgo"
)

type PokerGame struct {
	// Engine
	EngineOptions PokerEngineOptions
	actionSource  ActionSource

	// Table
	deck    *pokeralgo.Deck
	players []*GamePlayer

	// Hand
	CommunityCards []pokeralgo.Card

	DealerIndex int
	SbIndex     int
	BbIndex     int

	CurrentPlayerIndex int

	CurrentBet           int
	AdditionalRaiseCount int

	IsOnePlayerLeft  bool
	IsSkipToShowdown bool
}

func NewPokerGame(options PokerEngineOptions, actionSource ActionSource) *PokerGame {
	return &PokerGame{
		EngineOptions:  options,
		actionSource:   actionSource,
		players:        []*GamePlayer{},
		deck:           pokeralgo.NewDeck(),
		CommunityCards: make([]pokeralgo.Card, 0, 5),
	}
}

func (g *PokerGame) InitializeActionSource(actionSource ActionSource) error {
	if g.actionSource != nil {
		return newInternalPokerGameError("action source is already initialized, action source cannot be reassigned")
	}
	g.actionSource = actionSource
	return nil
}

func (g *PokerGame) InitializeTable(playersInfo []PlayerInfoDto) error {
	if g.actionSource == nil {
		return newInternalPokerGameError("action source has not been initialized")
	}
	if len(playersInfo) < 2 {
		return fmt.Errorf("Not enough players")
	}

	g.players = g.players[:0]
	g.deck.ResetDeck()
	g.DealerIndex = -1

	for _, pi := range playersInfo {
		player, err := NewGamePlayerFromInfo(pi, g.EngineOptions.BuyIn, g.deck.MustNextCard(), g.deck.MustNextCard())
		if err != nil {
			return err
		}
		g.players = append(g.players, player)
	}
	return nil
}

func (g *PokerGame) StartHand() error {
	if len(g.players) == 0 {
		return newPokerGameError("Call InitializeTable before StartHand.")
	}

	g.deck.ResetDeck()
	g.CommunityCards = g.CommunityCards[:0]
	g.AdvanceBlinds()

	// Pre-flop
	fmt.Println("Pre-Flop")
	if err := g.players[g.SbIndex].MakeBlindBet(g.EngineOptions.BigBlind / 2); err != nil {
		return err
	}
	if err := g.players[g.BbIndex].MakeBlindBet(g.EngineOptions.BigBlind); err != nil {
		return err
	}
	g.CurrentBet = g.EngineOptions.BigBlind

	for _, p := range g.players {
		if err := p.NewHand(g.deck.MustNextCard(), g.deck.MustNextCard()); err != nil {
			return err
		}
	}

	if err := g.ExecuteBetting(); err != nil {
		return err
	}
	g.CurrentPlayerIndex = g.GetNextIndex(g.DealerIndex)

	// Flop
	fmt.Println("Flop")
	g.deck.MustNextCard()
	if !g.IsOnePlayerLeft {
		g.CommunityCards = append(g.CommunityCards, g.deck.MustNextCards(3)...)
	}
	if err := g.ExecuteBetting(); err != nil {
		return err
	}
	g.CurrentPlayerIndex = g.GetNextIndex(g.DealerIndex)

	// Turn
	fmt.Println("Turn")
	g.deck.MustNextCard()
	if !g.IsOnePlayerLeft {
		g.CommunityCards = append(g.CommunityCards, g.deck.MustNextCard())
	}
	if err := g.ExecuteBetting(); err != nil {
		return err
	}
	g.CurrentPlayerIndex = g.GetNextIndex(g.DealerIndex)

	// River
	fmt.Println("River")
	g.deck.MustNextCard()
	if !g.IsOnePlayerLeft {
		g.CommunityCards = append(g.CommunityCards, g.deck.MustNextCard())
	}
	if err := g.ExecuteBetting(); err != nil {
		return err
	}

	if !g.IsOnePlayerLeft {
		if err := g.Showdown(); err != nil {
			return err
		}
	}

	g.ResetHand()
	return nil
}

func (g *PokerGame) ExecuteBetting() error {
	if g.IsOnePlayerLeft || g.IsSkipToShowdown {
		if g.IsSkipToShowdown {
			fmt.Println("Skip to showdown")
		}
		return nil
	}

	isBettingRoundOver := false
	for !isBettingRoundOver {
		for {
			currentPlayer := g.players[g.CurrentPlayerIndex]

			// Before
			if currentPlayer.HasFolded || isPlayerAllIn(currentPlayer) {
				g.CurrentPlayerIndex = g.GetNextIndex(g.CurrentPlayerIndex)
				continue
			} else if g.CountNonFoldedPlayers() == 1 {
				g.IsOnePlayerLeft = true
				isBettingRoundOver = true
				break
			} else if g.CountPlayersThatCanAct() < 2 && g.IsEveryoneSettledAtCurrentBet() {
				g.IsSkipToShowdown = true
				isBettingRoundOver = true
				break
			} else if g.HasEveryoneActed() && g.IsEveryoneSettledAtCurrentBet() {
				isBettingRoundOver = true
				break
			}

			// Action
			gameState := g.GetGameState()
			validMoves := gameState.PossibleMoves
			input, err := g.actionSource.NextAction(gameState)
			if err != nil {
				return err
			}
			if !containsMove(validMoves, input.Move) {
				return fmt.Errorf("Input not valid")
			}

			if input.Move == Fold {
				if err := currentPlayer.Fold(); err != nil {
					return err
				}
				if g.CountNonFoldedPlayers() == 1 {
					g.IsOnePlayerLeft = true
					isBettingRoundOver = true
					break
				}
			} else if input.Move == Raise || input.Move == Call {
				if g.HasEveryoneActed() && input.Move == Raise {
					g.AdditionalRaiseCount++
				}

				if input.Move == Raise {
					toCall := g.CurrentBet - currentPlayer.Bet
					if input.Amount < toCall {
						if err := currentPlayer.MakeBet(toCall + 10); err != nil {
							return err
						}
					} else {
						if err := currentPlayer.MakeBet(input.Amount); err != nil {
							return err
						}
					}
				} else {
					if err := currentPlayer.MakeBet(g.CurrentBet - currentPlayer.Bet); err != nil {
						return err
					}
				}

				if currentPlayer.Bet > g.CurrentBet {
					g.CurrentBet = currentPlayer.Bet
				}
			} else if input.Move == Check {
				currentPlayer.Check()
			}

			// After
			if g.CountPlayersThatCanAct() < 2 && g.IsEveryoneSettledAtCurrentBet() {
				isBettingRoundOver = true
				g.IsSkipToShowdown = true
				break
			}

			// Move To Next Player
			g.CurrentPlayerIndex = g.GetNextIndex(g.CurrentPlayerIndex)
		}
	}

	if g.IsOnePlayerLeft {
		fmt.Println("1 non-folded player remains")
		// 1 non-folded player remains
		winner, err := g.GetTheNonFoldedPlayer()
		if err != nil {
			return err
		}
		pot := 0
		for _, p := range g.players {
			pot += p.Bet
		}
		if err := winner.Pay(pot); err != nil {
			return err
		}

		if g.CountNonFoldedPlayers() != 1 {
			return fmt.Errorf("1 non-folded player remains but count is not 1")
		}
	}
	g.ResetBettingRound()
	return nil
}

func (g *PokerGame) Showdown() error {
	if g.IsOnePlayerLeft {
		fmt.Println("No Showdown. Only one player left")
		return nil
	}

	pots, err := GetPots(g.players)
	if err != nil {
		return err
	}
	algoPlayers := gamePlayersToAlgoPlayers(g.players)
	fmt.Println("All Algo Players")
	for _, item := range algoPlayers {
		fmt.Println(item)
	}
	fmt.Println()
	for _, pot := range pots {
		if len(pot.Players) == 1 {
			pot.Winners = pot.Players
		} else if len(pot.Players) < 1 {
			return fmt.Errorf("pot has no players")
		} else {
			if len(g.CommunityCards) != 5 {
				return fmt.Errorf("community cards count is not 5")
			}

			winners, err := pokeralgo.GetWinners(gamePlayersToAlgoPlayers(pot.Players), g.CommunityCards)
			if err != nil {
				return err
			}
			fmt.Println("Algo Winners for THIS pot")
			for _, item := range winners {
				fmt.Println(item)
				if item.WinningHand != nil {
					fmt.Println(item.WinningHand)
				}
			}
			fmt.Println()
			pot.Winners = g.MapAlgoPlayersToGamePlayers(winners)
		}
	}
	fmt.Print("\n--- PAY ---\n\n")
	for _, item := range pots {
		fmt.Print("\nPot:\n")
		fmt.Println(item)
		if err := item.PayWinners(); err != nil {
			return err
		}
	}
	return nil
}

func (g *PokerGame) MapAlgoPlayersToGamePlayers(players []pokeralgo.Player) []*GamePlayer {
	enginePlayers := append([]*GamePlayer(nil), g.players...)
	filtered := []*GamePlayer{}
	for _, p := range enginePlayers {
		for _, p2 := range players {
			if p2.Name == p.Name {
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

func gamePlayersToAlgoPlayers(enginePlayers []*GamePlayer) []pokeralgo.Player {
	players := []pokeralgo.Player{}
	for _, ep := range enginePlayers {
		players = append(players, pokeralgo.NewPlayer(ep.Name, ep.HoleCards.First, ep.HoleCards.Second))
	}
	return players
}

func (g *PokerGame) GetPossibleMoves(player *GamePlayer) []PlayerMove {
	if player.HasFolded {
		panic("Cannot get possible moves for folded player")
	}

	moves := []PlayerMove{}
	toCall := g.CurrentBet - player.Bet
	if toCall > 0 {
		moves = append(moves, Call)
		moves = append(moves, Fold)
	} else if toCall == 0 {
		moves = append(moves, Check)
	}
	// we count players that can act because we don't want to raise if another player is all-in
	if g.AdditionalRaiseCount != g.EngineOptions.AdditionalRaises && g.CountPlayersThatCanAct() > 1 {
		moves = append(moves, Raise)
	}
	return moves
}

func (g *PokerGame) GetTheNonFoldedPlayer() (*GamePlayer, error) {
	for _, p := range g.players {
		if !p.HasFolded {
			return p, nil
		}
	}
	return nil, fmt.Errorf("No non-folded player")
}

func (g *PokerGame) IsEveryoneSettledAtCurrentBet() bool {
	for _, p := range g.players {
		if !p.HasFolded && !isPlayerAllIn(p) && g.CurrentBet != p.Bet {
			return false
		}
	}
	return true
}

func (g *PokerGame) HasEveryoneActed() bool {
	for _, p := range g.players {
		if !p.HasFolded && !p.HasActed && !isPlayerAllIn(p) {
			return false
		}
	}
	return true
}

func (g *PokerGame) CountPlayersThatCanAct() int {
	count := 0
	for _, p := range g.players {
		if !p.HasFolded && !isPlayerAllIn(p) {
			count++
		}
	}
	return count
}

func (g *PokerGame) CountNonFoldedPlayers() int {
	count := 0
	for _, p := range g.players {
		if !p.HasFolded {
			count++
		}
	}
	return count
}

func isPlayerAllIn(player *GamePlayer) bool {
	if player.Stack == 0 && player.Bet == 0 {
		panic(&InternalPokerGameError{Message: "Player has 0 stack and 0 bet. Busted players should never be in play."})
	}

	if player.Stack == 0 && player.Bet > 0 {
		return true
	}
	return false
}

func (g *PokerGame) ResetBettingRound() {
	for _, p := range g.players {
		p.ResetBettingRound()
	}
	g.AdditionalRaiseCount = 0
}

func (g *PokerGame) ResetHand() {
	for _, p := range g.players {
		p.ResetHand()
	}
	g.AdditionalRaiseCount = 0
	g.IsOnePlayerLeft = false
	g.IsSkipToShowdown = false
}

func (g *PokerGame) GetGameState() GameState {
	playerStates := []PlayerState{}
	for _, player := range g.players {
		holeCards := player.HoleCards
		playerStates = append(playerStates, PlayerState{Id: player.Name, Stack: player.Stack, HoleCards: &holeCards, Bet: player.Bet, HasFolded: player.HasFolded})
	}
	currentPlayer := &playerStates[g.CurrentPlayerIndex]
	if currentPlayer.HasFolded {
		panic("Current player cannot be folded")
	}
	toCall := g.CurrentBet - currentPlayer.Bet
	return GameState{
		PlayerStates:   playerStates,
		CommunityCards: append([]pokeralgo.Card(nil), g.CommunityCards...),
		OutputType:     InputRequest,
		PlayerToAct:    currentPlayer,
		PossibleMoves:  g.GetPossibleMoves(g.players[g.CurrentPlayerIndex]),
		ToCall:         toCall,
	}
}

func (g *PokerGame) AdvanceBlinds() {
	g.DealerIndex = g.GetNextIndex(g.DealerIndex)
	g.SbIndex = g.GetNextIndex(g.DealerIndex)
	g.BbIndex = g.GetNextIndex(g.SbIndex)
	g.CurrentPlayerIndex = g.GetNextIndex(g.BbIndex)
}

func (g *PokerGame) GetNextIndex(index int) int {
	temp := index + 1
	if temp > len(g.players)-1 {
		temp = 0
	}
	return temp
}

func containsMove(moves []PlayerMove, target PlayerMove) bool {
	for _, move := range moves {
		if move == target {
			return true
		}
	}
	return false
}
