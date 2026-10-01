package pokergame

import "fmt"

type Pot struct {
	EligiblePlayers []*Player
	Amount          int

	Winners []*Player
}

func NewPot(value int, players []*Player) *Pot {
	return &Pot{
		EligiblePlayers: players,
		Amount:          value,
	}
}

func (p *Pot) Distribute() error {
	if len(p.Winners) == 0 {
		return fmt.Errorf("%w: Winners should never be null. This means we never determined the winner(s) of this pot.", ErrInternal)
	}

	split := p.Amount / len(p.Winners)
	for _, w := range p.Winners {
		if err := w.credit(split); err != nil {
			return err
		}
	}
	return nil
}

func (p *Pot) String() string {
	players := "| "
	for _, player := range p.EligiblePlayers {
		players += player.ID + " | "
	}

	wString := ""
	if p.Winners != nil {
		for _, winner := range p.Winners {
			wString += fmt.Sprintf("\t%s (%d) | %d => %d\n", winner.ID, p.Amount/len(p.Winners), winner.Stack, winner.Stack+p.Amount/len(p.Winners))
		}
	}

	return fmt.Sprintf("Players (%d): \n%s\nValue: %d\nWinner(s):\n%s", len(p.EligiblePlayers), players, p.Amount, wString)
}
