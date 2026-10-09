package pokergame

import (
	"fmt"

	pokeralgo "pokeralgo"
)

type player struct {
	ID        string
	Stack     int
	HoleCards pokeralgo.HoleCards
	Bet       int
	Acted     bool
	Folded    bool
	Left      bool
}

func newPlayer(id string, stack int, first pokeralgo.Card, second pokeralgo.Card) (*player, error) {
	if first.Equal(second) {
		return nil, fmt.Errorf("%w: Hole cards should never be the same card.", ErrInternal)
	}

	return &player{
		ID:        id,
		Stack:     stack,
		HoleCards: pokeralgo.HoleCards{First: first, Second: second},
	}, nil
}

func newPlayerFromSpec(playerInfo PlayerSpec, stack int, first pokeralgo.Card, second pokeralgo.Card) (*player, error) {
	if first.Equal(second) {
		return nil, fmt.Errorf("%w: Hole cards should never be the same card.", ErrInternal)
	}

	return &player{
		ID:        playerInfo.ID,
		Stack:     stack,
		HoleCards: pokeralgo.HoleCards{First: first, Second: second},
	}, nil
}

func (p *player) resetForNextHand() {
	p.Bet = 0
	p.Acted = false
	p.Folded = false
}

func (p *player) resetForNextStreet() {
	p.Acted = false
}

func (p *player) credit(amount int) error {
	if amount < 0 {
		return fmt.Errorf("%w: Pay amount cannot be negative.", ErrInternal)
	}

	p.Stack += amount
	return nil
}

func (p *player) fold() error {
	if p.Folded {
		return fmt.Errorf("%w: Player has already folded.", ErrInternal)
	}
	p.Folded = true
	return nil
}

func (p *player) leave() error {
	if p.Left {
		return fmt.Errorf("%w: Player already left.", ErrInternal)
	}
	p.Left = true
	return nil
}

func (p *player) check() {
	p.Acted = true
}

func (p *player) commitChips(amount int) error {
	if amount < 0 {
		return fmt.Errorf("%w: Bet amount cannot be negative.", ErrInternal)
	}

	if amount > p.Stack {
		p.Bet += p.Stack
		p.Stack = 0
		p.Acted = true
		return nil
	}
	p.Bet += amount
	p.Stack -= amount

	p.Acted = true
	return nil
}

func (p *player) postBlind(amount int) error {
	if amount < 0 {
		return fmt.Errorf("%w: Bet amount cannot be negative.", ErrInternal)
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

func (p *player) deal(first pokeralgo.Card, second pokeralgo.Card) error {
	if first.Equal(second) {
		return fmt.Errorf("%w: Hole cards should never be the same card.", ErrInternal)
	}
	p.HoleCards = pokeralgo.HoleCards{First: first, Second: second}
	return nil
}

func (p *player) String() string {
	return fmt.Sprintf("%s | CurrentBet: %d | Hand: %s | Stack: %d", p.ID, p.Bet, p.HoleCards, p.Stack)
}
