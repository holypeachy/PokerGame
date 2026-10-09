package pokergame

import (
	"errors"
	"reflect"
	"slices"
	"testing"

	"pokeralgo"
)

//TODO: Check cases

func showdownTestPlayer(id string, rank, bet int) *player {
	return &player{
		ID: id, Stack: 1000 - bet, Bet: bet,
		HoleCards: pokeralgo.HoleCards{
			First:  pokeralgo.MustCard(rank, pokeralgo.Clubs, true),
			Second: pokeralgo.MustCard(rank, pokeralgo.Diamonds, true),
		},
	}
}

func TestResolveShowdown(t *testing.T) {
	board := []pokeralgo.Card{
		pokeralgo.MustCard(2, pokeralgo.Spades, false),
		pokeralgo.MustCard(5, pokeralgo.Hearts, false),
		pokeralgo.MustCard(8, pokeralgo.Spades, false),
		pokeralgo.MustCard(11, pokeralgo.Hearts, false),
		pokeralgo.MustCard(12, pokeralgo.Spades, false),
	}
	var royalBoard []pokeralgo.Card
	for rank := 10; rank <= 14; rank++ {
		royalBoard = append(royalBoard, pokeralgo.MustCard(rank, pokeralgo.Spades, false))
	}
	folded := showdownTestPlayer("F", 12, 100)
	folded.Folded = true // Three queens would beat the active players' pairs.
	for _, tt := range []struct {
		name    string
		players []*player
		board   []pokeralgo.Card
		pots    []PotState
		stacks  []int
	}{
		{"sole winner", []*player{showdownTestPlayer("A", 14, 100), showdownTestPlayer("B", 13, 100)}, board,
			[]PotState{{EligiblePlayerIDs: []string{"A", "B"}, Amount: 200, WinnerIDs: []string{"A"}}}, []int{1100, 900}},
		{"board plays for both", []*player{showdownTestPlayer("A", 14, 100), showdownTestPlayer("B", 13, 100)}, royalBoard,
			[]PotState{{EligiblePlayerIDs: []string{"A", "B"}, Amount: 200, WinnerIDs: []string{"A", "B"}}}, []int{1000, 1000}},
		{"different main and side pot winners", []*player{showdownTestPlayer("C", 10, 150), showdownTestPlayer("A", 14, 50), showdownTestPlayer("B", 13, 100)}, board,
			[]PotState{
				{EligiblePlayerIDs: []string{"C", "A", "B"}, Amount: 150, WinnerIDs: []string{"A"}},
				{EligiblePlayerIDs: []string{"C", "B"}, Amount: 100, WinnerIDs: []string{"B"}},
				{EligiblePlayerIDs: []string{"C"}, Amount: 50, WinnerIDs: []string{"C"}},
			}, []int{900, 1100, 1000}},
		{"folded best hand cannot win", []*player{showdownTestPlayer("A", 14, 100), folded, showdownTestPlayer("B", 13, 100)}, board,
			[]PotState{{EligiblePlayerIDs: []string{"A", "B"}, Amount: 300, WinnerIDs: []string{"A"}}}, []int{1200, 900, 900}},
		{"single eligible player needs no board", []*player{showdownTestPlayer("A", 14, 100)}, nil,
			[]PotState{{EligiblePlayerIDs: []string{"A"}, Amount: 100, WinnerIDs: []string{"A"}}}, []int{1000}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var events []Event
			g := &Game{players: tt.players, board: slices.Clone(tt.board), street: River, handNumber: 3,
				options: Options{EventVerbosity: EventsAll}, onEvent: func(e Event) { events = append(events, e) }}
			original := slices.Clone(g.players)
			before := make([]player, len(original))
			for i, p := range original {
				before[i] = *p
			}
			if err := g.resolveShowdown(); err != nil {
				t.Fatal(err)
			}
			assertStreetEvents(t, events, []EventType{PotsCreated, WinnersDetermined, ChipsAwarded})
			if len(events) != 3 {
				t.Fatalf("expected three payout events, got %d", len(events))
			}
			for stage, e := range events {
				wantPots := slices.Clone(tt.pots)
				if stage == 0 {
					for i := range wantPots {
						wantPots[i].WinnerIDs = nil
					}
				}
				if e.Pots == nil || !reflect.DeepEqual(*e.Pots, wantPots) {
					t.Errorf("%s: expected pots %+v, got %v", e.Type, wantPots, e.Pots)
				}
				if e.HandNumber != 3 || e.Street != River || !slices.Equal(e.Board, tt.board) {
					t.Errorf("%s: incorrect hand context", e.Type)
				}
				if len(e.Players) != len(before) {
					t.Fatalf("%s: incorrect roster size", e.Type)
				}
				for i, p := range e.Players {
					wantStack := before[i].Stack
					if stage == 2 {
						wantStack = tt.stacks[i]
					}
					if p.ID != before[i].ID || p.Stack != wantStack {
						t.Errorf("%s: player %d expected %s stack %d, got %+v", e.Type, i, before[i].ID, wantStack, p)
					}
				}
			}
			for i, p := range g.players {
				want := before[i]
				want.Stack = tt.stacks[i]
				if p != original[i] || *p != want {
					t.Errorf("player %d: payout must update the original player and only its stack", i)
				}
			}
			if !slices.Equal(g.board, tt.board) {
				t.Error("showdown modified the board")
			}
		})
	}
}

func TestResolveShowdownErrorsDoNotAwardChips(t *testing.T) {
	for _, name := range []string{"short board", "long board", "all folded", "no contributions", "negative contribution"} {
		t.Run(name, func(t *testing.T) {
			g := &Game{players: []*player{showdownTestPlayer("A", 14, 100), showdownTestPlayer("B", 13, 100)},
				options: Options{EventVerbosity: EventsAll}}
			switch name {
			case "short board":
				g.board = make([]pokeralgo.Card, 4)
			case "long board":
				g.board = make([]pokeralgo.Card, 6)
			case "all folded":
				g.players[0].Folded, g.players[1].Folded = true, true
			case "no contributions":
				g.players[0].Bet, g.players[1].Bet = 0, 0
			case "negative contribution":
				g.players[0].Bet = -1
			}
			before := []player{*g.players[0], *g.players[1]}
			var events []Event
			g.onEvent = func(e Event) { events = append(events, e) }
			if err := g.resolveShowdown(); !errors.Is(err, ErrInternal) {
				t.Errorf("expected ErrInternal, got %v", err)
			}
			for i, p := range g.players {
				if *p != before[i] {
					t.Errorf("error changed player %d", i)
				}
			}
			for _, e := range events {
				if e.Type == WinnersDetermined || e.Type == ChipsAwarded {
					t.Errorf("unexpected %s after failed showdown", e.Type)
				}
			}
		})
	}
}

func TestResolveShowdownSkipsOnePlayerLeft(t *testing.T) {
	p := showdownTestPlayer("A", 14, 100)
	p.Stack = 1100 // Already credited by runStreet.
	g := &Game{players: []*player{p}, onePlayerLeft: true, options: Options{EventVerbosity: EventsAll},
		onEvent: func(e Event) { t.Errorf("unexpected event %s", e.Type) }}
	before := *p
	if err := g.resolveShowdown(); err != nil {
		t.Fatal(err)
	}
	if *p != before {
		t.Error("showdown changed an already-paid player")
	}
}

func TestPotDistribute(t *testing.T) {
	for _, tt := range []struct {
		name                string
		amount, winnerCount int
		wantErr             bool
	}{
		{"sole winner", 90, 1, false},
		{"two winners", 90, 2, false},
		{"three winners", 90, 3, false},
		{"zero amount", 0, 2, false},
		{"no winners", 90, 0, true},
		{"negative award", -90, 2, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			players := []*player{{ID: "A", Stack: 100}, {ID: "B", Stack: 200}, {ID: "C", Stack: 300}}
			p := newPot(tt.amount, players)
			p.Winners = players[:tt.winnerCount]
			err := p.distribute()
			if tt.wantErr {
				if !errors.Is(err, ErrInternal) {
					t.Errorf("expected ErrInternal, got %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			for i, player := range players {
				want := (i + 1) * 100
				if !tt.wantErr && i < tt.winnerCount {
					want += tt.amount / tt.winnerCount
				}
				if player.Stack != want {
					t.Errorf("%s: expected stack %d, got %d", player.ID, want, player.Stack)
				}
			}
		})
	}
}
