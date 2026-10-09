package pokergame

import (
	"fmt"
	"strings"
)

type Pot struct {
	EligiblePlayers []*player
	Amount          int

	Winners []*player
}

func newPot(value int, players []*player) *Pot {
	return &Pot{
		EligiblePlayers: players,
		Amount:          value,
	}
}

func (p *Pot) distribute() error {
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
	players := make([]string, len(p.EligiblePlayers))
	for i, player := range p.EligiblePlayers {
		players[i] = player.ID
	}

	winners := make([]string, len(p.Winners))
	for i, winner := range p.Winners {
		winners[i] = winner.ID
	}

	return fmt.Sprintf("Pot Eligible:[%s] Amount:%d Winners:[%s]", strings.Join(players, ", "), p.Amount, strings.Join(winners, ", "))
}
