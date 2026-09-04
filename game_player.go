package pokergame

import (
	"fmt"

	pokeralgo "pokeralgo"
)

type GamePlayer struct {
	Name      string
	Stack     int
	HoleCards pokeralgo.Pair
	Bet       int
	HasActed  bool
	HasFolded bool
}

func NewGamePlayer(name string, stack int, first pokeralgo.Card, second pokeralgo.Card) (*GamePlayer, error) {
	if first.Equal(second) {
		return nil, newInternalPokerGameError("Hole cards should never be the same card.")
	}

	return &GamePlayer{
		Name:      name,
		Stack:     stack,
		HoleCards: pokeralgo.Pair{First: first, Second: second},
	}, nil
}

func NewGamePlayerFromInfo(playerInfo PlayerInfoDto, stack int, first pokeralgo.Card, second pokeralgo.Card) (*GamePlayer, error) {
	if first.Equal(second) {
		return nil, newInternalPokerGameError("Hole cards should never be the same card.")
	}

	return &GamePlayer{
		Name:      playerInfo.Id,
		Stack:     stack,
		HoleCards: pokeralgo.Pair{First: first, Second: second},
	}, nil
}

func (p *GamePlayer) ResetHand() {
	p.Bet = 0
	p.HasActed = false
	p.HasFolded = false
}

func (p *GamePlayer) ResetBettingRound() {
	p.HasActed = false
}

func (p *GamePlayer) Pay(amount int) error {
	if amount < 0 {
		return newInternalPokerGameError("Pay amount cannot be negative.")
	}

	p.Stack += amount
	return nil
}

func (p *GamePlayer) Fold() error {
	if p.HasFolded {
		return newInternalPokerGameError("Player has already folded.")
	}
	p.HasFolded = true
	return nil
}

func (p *GamePlayer) Check() {
	p.HasActed = true
}

func (p *GamePlayer) MakeBet(amount int) error {
	if amount < 0 {
		return newInternalPokerGameError("Bet amount cannot be negative.")
	}

	if err := p.MakeBlindBet(amount); err != nil {
		return err
	}
	p.HasActed = true
	return nil
}

func (p *GamePlayer) MakeBlindBet(amount int) error {
	if amount < 0 {
		return newInternalPokerGameError("Bet amount cannot be negative.")
	}

	if amount > p.Stack {
		p.Bet += p.Stack
		p.Stack = 0
		return nil
	}
	p.Bet += amount
	p.Stack -= amount
	return nil
}

func (p *GamePlayer) NewHand(first pokeralgo.Card, second pokeralgo.Card) error {
	if first.Equal(second) {
		return newInternalPokerGameError("Hole cards should never be the same card.")
	}
	p.HoleCards = pokeralgo.Pair{First: first, Second: second}
	return nil
}

func (p *GamePlayer) String() string {
	return fmt.Sprintf("%s\nCurrentBet: %d | Hand: %s | Stack: %d", p.Name, p.Bet, p.HoleCards, p.Stack)
}
