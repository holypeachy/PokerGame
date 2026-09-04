package pokergame

import "fmt"

type Pot struct {
	Players []*GamePlayer
	Value   int

	Winners []*GamePlayer
}

func NewPot(value int, players []*GamePlayer) *Pot {
	return &Pot{
		Players: players,
		Value:   value,
	}
}

func (p *Pot) PayWinners() error {
	if p.Winners == nil || len(p.Winners) == 0 {
		return fmt.Errorf("Winners should never be null. This means we never determined the winner(s) of this pot.")
	}

	split := p.Value / len(p.Winners)
	for _, w := range p.Winners {
		if err := w.Pay(split); err != nil {
			return err
		}
	}
	return nil
}

func (p *Pot) String() string {
	players := "| "
	for _, player := range p.Players {
		players += player.Name + " | "
	}

	wString := ""
	if p.Winners != nil {
		for _, winner := range p.Winners {
			wString += fmt.Sprintf("\t%s (%d) | %d => %d\n", winner.Name, p.Value/len(p.Winners), winner.Stack, winner.Stack+p.Value/len(p.Winners))
		}
	}

	return fmt.Sprintf("Players (%d): \n%s\nValue: %d\nWinner(s):\n%s", len(p.Players), players, p.Value, wString)
}
