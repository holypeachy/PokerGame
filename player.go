package pokergame

import (
	"fmt"

	pokeralgo "pokeralgo"
)

type Player struct {
	ID        string
	Stack     int
	HoleCards pokeralgo.HoleCards
	Bet       int
	Acted     bool
	Folded    bool
	Left      bool
}

func NewPlayer(id string, stack int, first pokeralgo.Card, second pokeralgo.Card) (*Player, error) {
	if first.Equal(second) {
		return nil, fmt.Errorf("%w: Hole cards should never be the same card.", ErrInternal)
	}

	return &Player{
		ID:        id,
		Stack:     stack,
		HoleCards: pokeralgo.HoleCards{First: first, Second: second},
	}, nil
}

func NewPlayerFromSpec(playerInfo PlayerSpec, stack int, first pokeralgo.Card, second pokeralgo.Card) (*Player, error) {
	if first.Equal(second) {
		return nil, fmt.Errorf("%w: Hole cards should never be the same card.", ErrInternal)
	}

	return &Player{
		ID:        playerInfo.ID,
		Stack:     stack,
		HoleCards: pokeralgo.HoleCards{First: first, Second: second},
	}, nil
}

func (p *Player) resetForNextHand() {
	p.Bet = 0
	p.Acted = false
	p.Folded = false
}

func (p *Player) resetForStreet() {
	p.Acted = false
}

func (p *Player) credit(amount int) error {
	if amount < 0 {
		return fmt.Errorf("%w: Pay amount cannot be negative.", ErrInternal)
	}

	p.Stack += amount
	return nil
}

func (p *Player) Fold() error {
	if p.Folded {
		return fmt.Errorf("%w: Player has already folded.", ErrInternal)
	}
	p.Folded = true
	return nil
}

func (p *Player) Leave() error {
	if p.Left {
		return fmt.Errorf("%w: Player already left.", ErrInternal)
	}
	p.Left = true
	return nil
}

func (p *Player) Check() {
	p.Acted = true
}

func (p *Player) commitChips(amount int) error {
	if amount < 0 {
		return fmt.Errorf("%w: Bet amount cannot be negative.", ErrInternal)
	}

	if err := p.postBlind(amount); err != nil {
		return err
	}
	p.Acted = true
	return nil
}

func (p *Player) postBlind(amount int) error {
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

func (p *Player) deal(first pokeralgo.Card, second pokeralgo.Card) error {
	if first.Equal(second) {
		return fmt.Errorf("%w: Hole cards should never be the same card.", ErrInternal)
	}
	p.HoleCards = pokeralgo.HoleCards{First: first, Second: second}
	return nil
}

func (p *Player) String() string {
	return fmt.Sprintf("%s\nCurrentBet: %d | Hand: %s | Stack: %d", p.ID, p.Bet, p.HoleCards, p.Stack)
}
