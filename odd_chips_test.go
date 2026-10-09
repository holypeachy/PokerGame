package pokergame

import (
	"slices"
	"testing"

	"pokeralgo"
)

//TODO: Check cases

func TestPotDistributeOddChips(t *testing.T) {
	for _, tt := range []struct {
		name          string
		amount, count int
	}{
		{"one remainder two winners", 101, 2},
		{"one remainder three winners", 100, 3},
		{"two remainders three winners", 101, 3},
		{"one chip two winners", 1, 2},
		{"two chips three winners", 2, 3},
	} {
		t.Run(tt.name, func(t *testing.T) {
			players := []*player{{ID: "A", Stack: 100}, {ID: "B", Stack: 200}, {ID: "C", Stack: 300}, {ID: "F", Stack: 400, Folded: true}}
			winners := []*player{players[2], players[0], players[1]}
			winners = winners[:tt.count]
			p := newPot(tt.amount, players[:3])
			p.Winners = slices.Clone(winners)
			if err := p.distribute(); err != nil {
				t.Fatal(err)
			}
			total, extras := 0, 0
			for i, player := range players {
				received := player.Stack - (i+1)*100
				total += received
				if !slices.Contains(winners, player) {
					if received != 0 {
						t.Errorf("non-winner %s received %d chips", player.ID, received)
					}
					continue
				}
				share := tt.amount / tt.count
				if received != share && received != share+1 {
					t.Errorf("%s received %d; expected %d or %d", player.ID, received, share, share+1)
				}
				if received == share+1 {
					extras++
				}
			}
			if total != tt.amount {
				t.Errorf("lost chips: pot=%d distributed=%d", tt.amount, total)
			}
			if extras != tt.amount%tt.count {
				t.Errorf("expected %d extra-chip recipients, got %d", tt.amount%tt.count, extras)
			}
			if p.Amount != tt.amount || !slices.Equal(p.Winners, winners) {
				t.Error("distribution changed the pot amount or winners")
			}
		})
	}
}

func TestResolveShowdownOddChips(t *testing.T) {
	for _, tt := range []struct {
		name              string
		active, foldedBet int
	}{
		{"two tied winners", 2, 1},
		{"three tied winners", 3, 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := &Game{street: River, options: Options{EventVerbosity: EventsAll}}
			for rank := 10; rank <= 14; rank++ {
				g.board = append(g.board, pokeralgo.MustCard(rank, pokeralgo.Spades, false))
			}
			ids := []string{"A", "B", "C"}[:tt.active]
			for i, id := range ids {
				g.players = append(g.players, showdownTestPlayer(id, 14-i, 100))
			}
			folded := showdownTestPlayer("F", 8, tt.foldedBet)
			folded.Folded = true
			g.players = append(g.players, folded)
			var payout []Event
			g.onEvent = func(e Event) {
				if e.Type == ChipsAwarded {
					payout = append(payout, e)
				}
			}
			if err := g.resolveShowdown(); err != nil {
				t.Fatal(err)
			}
			amount := tt.active*100 + tt.foldedBet
			total := 0
			for _, p := range g.players {
				total += p.Stack
			}
			if total != len(g.players)*1000 {
				t.Errorf("lost chips across showdown: expected %d, got %d", len(g.players)*1000, total)
			}
			for _, p := range g.players[:tt.active] {
				received := p.Stack - 900
				if received != amount/tt.active && received != amount/tt.active+1 {
					t.Errorf("unfair share for %s: %d", p.ID, received)
				}
			}
			if folded.Stack != 1000-tt.foldedBet {
				t.Error("folded player received part of the payout")
			}
			if len(payout) != 1 || payout[0].Pots == nil || len(*payout[0].Pots) != 1 {
				t.Fatal("missing payout event")
			}
			pot := (*payout[0].Pots)[0]
			if pot.Amount != amount || !slices.Equal(pot.WinnerIDs, ids) {
				t.Errorf("incorrect payout pot: %+v", pot)
			}
			if len(payout[0].Players) != len(g.players) {
				t.Fatal("incorrect payout roster")
			}
			for i, p := range payout[0].Players {
				if p.Stack != g.players[i].Stack {
					t.Errorf("payout snapshot disagrees with actual stack for %s", p.ID)
				}
			}
		})
	}
}
