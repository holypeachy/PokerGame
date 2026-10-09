package pokergame

import (
	"slices"
	"testing"

	"pokeralgo"
)

//TODO: Check cases

func TestPlayHandSeededDealing(t *testing.T) {
	for _, tt := range []struct {
		name                   string
		ids, preflop, postflop []string
		calls                  []int
	}{
		{"heads up", []string{"A", "B"}, []string{"A", "B"}, []string{"B", "A"}, []int{25, 0}},
		{"three players", []string{"A", "B", "C"}, []string{"A", "B", "C"}, []string{"B", "C", "A"}, []int{50, 25, 0}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var events []Event
			g := New(Options{BuyIn: 1000, BigBlind: 50, AdditionalRaises: 1, EventVerbosity: EventsAll}, nil,
				func(e Event) { events = append(events, e) })
			g.deck = pokeralgo.NewDeckWithSeed(42)
			source := &streetActionSource{t: t, game: g}
			for i, id := range tt.preflop {
				action, legal := Call, []ActionType{Fold, Call, Raise, Leave}
				if tt.calls[i] == 0 {
					action, legal = Check, []ActionType{Fold, Check, Raise, Leave}
				}
				source.steps = append(source.steps, streetStep{id, tt.calls[i], legal, Action{Type: action}, 0})
			}
			for range 3 {
				for _, id := range tt.postflop {
					source.steps = append(source.steps, streetStep{id, 0, []ActionType{Fold, Check, Raise, Leave}, Action{Type: Check}, 0})
				}
			}
			g.actionSource = source
			var specs []PlayerSpec
			for _, id := range tt.ids {
				specs = append(specs, NewPlayerSpec(id))
			}
			if err := g.seatPlayers(specs); err != nil {
				t.Fatal(err)
			}

			// Seating and starting the hand each reset the seeded deck.
			reference := pokeralgo.NewDeckWithSeed(42)
			reference.Reset()
			reference.Reset()
			cards := reference.Cards()
			holeCount := len(tt.ids) * 2
			wantBoard := []pokeralgo.Card{cards[holeCount+1], cards[holeCount+2], cards[holeCount+3], cards[holeCount+5], cards[holeCount+7]}
			if err := g.playHand(); err != nil {
				t.Fatal(err)
			}
			if source.next != len(source.steps) {
				t.Errorf("consumed %d of %d actions", source.next, len(source.steps))
			}
			dealt := 0
			var streets []Street
			for _, e := range events {
				if e.Type == HoleCardsDealt {
					dealt++
					if len(e.Players) != len(tt.ids) {
						t.Fatal("incorrect dealt roster")
					}
					for i, p := range e.Players {
						want := pokeralgo.HoleCards{First: cards[2*i], Second: cards[2*i+1]}
						if p.ID != tt.ids[i] || p.HoleCards != want {
							t.Errorf("player %d: expected %s cards %v, got %+v", i, tt.ids[i], want, p)
						}
					}
				}
				if e.Type == StreetStarted {
					streets = append(streets, e.Street)
					lengths := map[Street]int{Preflop: 0, Flop: 3, Turn: 4, River: 5}
					n, ok := lengths[e.Street]
					if !ok {
						t.Fatalf("unexpected street %s", e.Street)
					}
					if !slices.Equal(e.Board, wantBoard[:n]) {
						t.Errorf("%s: expected board %v, got %v", e.Street, wantBoard[:n], e.Board)
					}
				}
			}
			if dealt != 1 || !slices.Equal(streets, []Street{Preflop, Flop, Turn, River}) {
				t.Errorf("unexpected dealing count %d or streets %v", dealt, streets)
			}
			if !slices.Equal(g.board, wantBoard) {
				t.Errorf("expected final board %v, got %v", wantBoard, g.board)
			}
			if got, want := g.deck.Remaining(), 52-holeCount-8; got != want {
				t.Errorf("expected %d cards remaining after three burns, got %d", want, got)
			}
		})
	}
}

func TestPlayHandShortBlindStacks(t *testing.T) {
	callRaise, checkRaise := []ActionType{Fold, Call, Raise, Leave}, []ActionType{Fold, Check, Raise, Leave}
	for _, tt := range []struct {
		name                 string
		stacks, posted, pots []int
		eligible             [][]string
		preflop              []streetStep
		postflop             []string
	}{
		{"short small blind", []int{1000, 10, 1000}, []int{0, 10, 50}, []int{30, 80}, [][]string{{"A", "B", "C"}, {"A", "C"}},
			[]streetStep{{"A", 50, callRaise, Action{Type: Call}, 0}, {"C", 0, checkRaise, Action{Type: Check}, 0}}, []string{"C", "A"}},
		{"short big blind", []int{1000, 1000, 10}, []int{0, 25, 10}, []int{30, 80}, [][]string{{"A", "B", "C"}, {"A", "B"}},
			[]streetStep{{"A", 50, callRaise, Action{Type: Call}, 0}, {"B", 25, callRaise, Action{Type: Call}, 0}}, []string{"B", "A"}},
		{"both blinds short", []int{1000, 10, 20}, []int{0, 10, 20}, []int{30, 20, 30}, [][]string{{"A", "B", "C"}, {"A", "C"}, {"A"}},
			[]streetStep{{"A", 50, []ActionType{Fold, Call, Leave}, Action{Type: Call}, 0}}, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var events []Event
			g := New(Options{BuyIn: 1000, BigBlind: 50, AdditionalRaises: 1, EventVerbosity: EventsAll}, nil,
				func(e Event) { events = append(events, e) })
			g.deck = pokeralgo.NewDeckWithSeed(42)
			source := &streetActionSource{t: t, game: g, steps: slices.Clone(tt.preflop)}
			for range 3 {
				for _, id := range tt.postflop {
					source.steps = append(source.steps, streetStep{id, 0, checkRaise, Action{Type: Check}, 0})
				}
			}
			g.actionSource = source
			if err := g.seatPlayers([]PlayerSpec{NewPlayerSpec("A"), NewPlayerSpec("B"), NewPlayerSpec("C")}); err != nil {
				t.Fatal(err)
			}
			initialTotal := 0
			for i, stack := range tt.stacks {
				g.players[i].Stack = stack
				initialTotal += stack
			}
			original := slices.Clone(g.players)
			if err := g.playHand(); err != nil {
				t.Fatal(err)
			}
			if source.next != len(source.steps) {
				t.Errorf("consumed %d of %d actions", source.next, len(source.steps))
			}
			posted, paid := 0, 0
			for _, e := range events {
				if e.Type == BlindsPosted {
					posted++
					if len(e.Players) != 3 {
						t.Fatal("incorrect blinds roster")
					}
					for i, p := range e.Players {
						if p.Bet != tt.posted[i] || p.Stack != tt.stacks[i]-tt.posted[i] {
							t.Errorf("incorrect blind deduction for %s: %+v", p.ID, p)
						}
					}
				}
				if e.Type == ChipsAwarded {
					paid++
					if len(e.Board) != 5 || e.Pots == nil || len(*e.Pots) != len(tt.pots) {
						t.Fatal("missing full board or expected side pots")
					}
					awards := map[string]int{}
					for i, p := range *e.Pots {
						if p.Amount != tt.pots[i] || !slices.Equal(p.EligiblePlayerIDs, tt.eligible[i]) {
							t.Errorf("pot %d: unexpected %+v", i, p)
						}
						if len(p.WinnerIDs) == 0 {
							t.Fatal("pot has no winners")
						}
						for _, id := range p.WinnerIDs {
							if !slices.Contains(p.EligiblePlayerIDs, id) {
								t.Errorf("ineligible winner %s", id)
							}
							awards[id] += p.Amount / len(p.WinnerIDs)
						}
					}
					if len(e.Players) != 3 {
						t.Fatal("players removed before payout")
					}
					total := 0
					for i, p := range e.Players {
						contribution := min(tt.stacks[i], 50)
						if p.Bet != contribution || p.Stack != tt.stacks[i]-contribution+awards[p.ID] {
							t.Errorf("unexpected payout for %s: %+v", p.ID, p)
						}
						total += p.Stack
					}
					if total != initialTotal {
						t.Errorf("expected total %d chips, got %d", initialTotal, total)
					}
				}
			}
			if posted != 1 || paid != 1 {
				t.Errorf("expected one blind and payout event, got %d and %d", posted, paid)
			}
			for _, p := range original {
				if slices.Contains(g.players, p) != (p.Stack > 0) {
					t.Errorf("incorrect survivor status for %s", p.ID)
				}
			}
		})
	}
}
