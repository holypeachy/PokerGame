package pokergame

import (
	"errors"
	"reflect"
	"slices"
	"testing"

	"pokeralgo"
)

type streetStep struct {
	playerID string
	toCall   int
	legal    []ActionType
	action   Action
	raises   int
}

type streetActionSource struct {
	t     *testing.T
	game  *Game
	steps []streetStep
	next  int
}

//TODO: Check cases

func (s *streetActionSource) NextAction(request ActionRequest) Action {
	s.t.Helper()
	if s.next >= len(s.steps) {
		s.t.Fatalf("unexpected action request after %d scripted actions", s.next)
	}
	step := s.steps[s.next]
	s.next++
	if request.PlayerToAct == nil {
		s.t.Fatalf("step %d: missing player", s.next)
	}
	if request.PlayerToAct.ID != step.playerID {
		s.t.Fatalf("step %d: expected player %s, got %s", s.next, step.playerID, request.PlayerToAct.ID)
	}
	if request.ToCall != step.toCall {
		s.t.Fatalf("step %d: expected ToCall %d, got %d", s.next, step.toCall, request.ToCall)
	}
	if !slices.Equal(request.LegalActions, step.legal) {
		s.t.Fatalf("step %d: expected legal actions %v, got %v", s.next, step.legal, request.LegalActions)
	}
	if s.game.additionalRaiseCount != step.raises {
		s.t.Fatalf("step %d: expected additional raise count %d, got %d", s.next, step.raises, s.game.additionalRaiseCount)
	}
	return step.action
}

func newStreetTestGame(t *testing.T, players []player, currentBet int, steps []streetStep) (*Game, *[]Event) {
	t.Helper()
	g := &Game{
		options:    Options{AdditionalRaises: 1, EventVerbosity: EventsAll},
		handNumber: 1,
		street:     Flop,
		currentBet: currentBet,
		board: []pokeralgo.Card{
			pokeralgo.MustCard(2, pokeralgo.Clubs, false),
			pokeralgo.MustCard(7, pokeralgo.Hearts, false),
			pokeralgo.MustCard(13, pokeralgo.Spades, false),
		},
	}
	for _, p := range players {
		g.players = append(g.players, &p)
	}
	source := &streetActionSource{t: t, game: g, steps: steps}
	g.actionSource = source
	t.Cleanup(func() {
		if source.next != len(source.steps) {
			t.Errorf("consumed %d of %d scripted actions", source.next, len(source.steps))
		}
	})
	var events []Event
	g.onEvent = func(e Event) { events = append(events, e) }
	return g, &events
}

func TestRunStreetBetting(t *testing.T) {
	checkRaise := []ActionType{Fold, Check, Raise, Leave}
	callRaise := []ActionType{Fold, Call, Raise, Leave}
	callOnly := []ActionType{Fold, Call, Leave}
	tests := []struct {
		name        string
		players     []player
		currentBet  int
		steps       []streetStep
		wantPlayers []player
		wantBet     int
	}{
		{
			name:       "check around",
			players:    []player{{ID: "A", Stack: 950, Bet: 50}, {ID: "B", Stack: 950, Bet: 50}, {ID: "C", Stack: 950, Bet: 50}},
			currentBet: 50,
			steps: []streetStep{
				{"A", 0, checkRaise, Action{Type: Check}, 0},
				{"B", 0, checkRaise, Action{Type: Check}, 0},
				{"C", 0, checkRaise, Action{Type: Check}, 0},
			},
			wantPlayers: []player{{ID: "A", Stack: 950, Bet: 50}, {ID: "B", Stack: 950, Bet: 50}, {ID: "C", Stack: 950, Bet: 50}},
			wantBet:     50,
		},
		{
			name:       "call around preserves big blind option",
			players:    []player{{ID: "A", Stack: 1000}, {ID: "B", Stack: 975, Bet: 25}, {ID: "C", Stack: 950, Bet: 50}},
			currentBet: 50,
			steps: []streetStep{
				{"A", 50, callRaise, Action{Type: Call}, 0},
				{"B", 25, callRaise, Action{Type: Call}, 0},
				{"C", 0, checkRaise, Action{Type: Check}, 0},
			},
			wantPlayers: []player{{ID: "A", Stack: 950, Bet: 50}, {ID: "B", Stack: 950, Bet: 50}, {ID: "C", Stack: 950, Bet: 50}},
			wantBet:     50,
		},
		{
			name:       "earlier check responds to raise",
			players:    []player{{ID: "A", Stack: 950, Bet: 50}, {ID: "B", Stack: 950, Bet: 50}, {ID: "C", Stack: 950, Bet: 50}},
			currentBet: 50,
			steps: []streetStep{
				{"A", 0, checkRaise, Action{Type: Check}, 0},
				{"B", 0, checkRaise, Action{Type: Raise, Amount: 50}, 0},
				{"C", 50, callRaise, Action{Type: Call}, 0},
				{"A", 50, callRaise, Action{Type: Call}, 0},
			},
			wantPlayers: []player{{ID: "A", Stack: 900, Bet: 100}, {ID: "B", Stack: 900, Bet: 100}, {ID: "C", Stack: 900, Bet: 100}},
			wantBet:     100,
		},
		{
			name:       "additional raise reaches cap and requires responses",
			players:    []player{{ID: "A", Stack: 950, Bet: 50}, {ID: "B", Stack: 950, Bet: 50}, {ID: "C", Stack: 950, Bet: 50}},
			currentBet: 50,
			steps: []streetStep{
				{"A", 0, checkRaise, Action{Type: Check}, 0},
				{"B", 0, checkRaise, Action{Type: Raise, Amount: 50}, 0},
				{"C", 50, callRaise, Action{Type: Call}, 0},
				{"A", 50, callRaise, Action{Type: Raise, Amount: 100}, 0},
				{"B", 50, callOnly, Action{Type: Call}, 1},
				{"C", 50, callOnly, Action{Type: Call}, 1},
			},
			wantPlayers: []player{{ID: "A", Stack: 850, Bet: 150}, {ID: "B", Stack: 850, Bet: 150}, {ID: "C", Stack: 850, Bet: 150}},
			wantBet:     150,
		},
		{
			name:       "skip folded and all-in players",
			players:    []player{{ID: "F", Stack: 975, Bet: 25, Folded: true}, {ID: "I", Bet: 100}, {ID: "A", Stack: 950, Bet: 50}, {ID: "B", Stack: 950, Bet: 50}},
			currentBet: 100,
			steps: []streetStep{
				{"A", 50, callRaise, Action{Type: Call}, 0},
				{"B", 50, callRaise, Action{Type: Call}, 0},
			},
			wantPlayers: []player{{ID: "F", Stack: 975, Bet: 25, Folded: true}, {ID: "I", Bet: 100}, {ID: "A", Stack: 900, Bet: 100}, {ID: "B", Stack: 900, Bet: 100}},
			wantBet:     100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, events := newStreetTestGame(t, tt.players, tt.currentBet, tt.steps)
			if err := g.runStreet(); err != nil {
				t.Fatalf("runStreet: %v", err)
			}
			assertStreetPlayers(t, g, tt.wantPlayers)
			if g.currentBet != tt.wantBet || g.onePlayerLeft || g.runToShowdown {
				t.Errorf("unexpected street result: bet=%d onePlayerLeft=%t runToShowdown=%t", g.currentBet, g.onePlayerLeft, g.runToShowdown)
			}
			if g.additionalRaiseCount != 0 {
				t.Errorf("additional raise count not reset: %d", g.additionalRaiseCount)
			}
			wantEvents := []EventType{StreetStarted}
			for range tt.steps {
				wantEvents = append(wantEvents, ActionRequested, ActionValid)
			}
			wantEvents = append(wantEvents, StreetEnded)
			assertStreetEvents(t, *events, wantEvents)
		})
	}
}

func TestRunStreetRunout(t *testing.T) {
	callOnly := []ActionType{Fold, Call, Leave}
	tests := []struct {
		name    string
		players []player
		steps   []streetStep
		want    []player
	}{
		{"all players all-in", []player{{ID: "A", Bet: 100}, {ID: "B", Bet: 100}}, nil,
			[]player{{ID: "A", Bet: 100}, {ID: "B", Bet: 100}}},
		{"one actionable player already settled", []player{{ID: "A", Bet: 100}, {ID: "B", Stack: 900, Bet: 100}}, nil,
			[]player{{ID: "A", Bet: 100}, {ID: "B", Stack: 900, Bet: 100}}},
		{"one actionable player still owes", []player{{ID: "A", Bet: 100}, {ID: "B", Stack: 950, Bet: 50}},
			[]streetStep{{"B", 50, callOnly, Action{Type: Call}, 0}},
			[]player{{ID: "A", Bet: 100}, {ID: "B", Stack: 900, Bet: 100}}},
		{"short call is clamped to stack", []player{{ID: "A", Bet: 100}, {ID: "B", Stack: 25, Bet: 50}},
			[]streetStep{{"B", 50, callOnly, Action{Type: Call}, 0}},
			[]player{{ID: "A", Bet: 100}, {ID: "B", Bet: 75}}},
		{"exact all-in call", []player{{ID: "A", Bet: 100}, {ID: "B", Stack: 50, Bet: 50}},
			[]streetStep{{"B", 50, callOnly, Action{Type: Call}, 0}},
			[]player{{ID: "A", Bet: 100}, {ID: "B", Bet: 100}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, events := newStreetTestGame(t, tt.players, 100, tt.steps)
			beforeBoard := slices.Clone(g.board)
			if err := g.runStreet(); err != nil {
				t.Fatalf("runStreet: %v", err)
			}
			assertStreetPlayers(t, g, tt.want)
			if !g.runToShowdown || g.onePlayerLeft {
				t.Errorf("expected runout without a lone winner")
			}
			if !slices.Equal(g.board, beforeBoard) || g.street != Flop {
				t.Error("runStreet should leave board dealing and street advancement to its caller")
			}
			wantEvents := []EventType{StreetStarted}
			for range tt.steps {
				wantEvents = append(wantEvents, ActionRequested, ActionValid)
			}
			wantEvents = append(wantEvents, RunToShowdown, StreetEnded)
			assertStreetEvents(t, *events, wantEvents)
			// A later street in runout mode announces the street without asking for input.
			if err := g.runStreet(); err != nil {
				t.Fatalf("subsequent runout street: %v", err)
			}
			assertStreetPlayers(t, g, tt.want)
			assertStreetEvents(t, *events, append(wantEvents, StreetStarted))
		})
	}
}

func TestRunStreetRaiseFallsBackToCall(t *testing.T) {
	for _, tt := range []struct {
		name   string
		amount int
	}{
		{"below amount to call", 25},
		{"equal to amount to call", 50},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g, events := newStreetTestGame(t,
				[]player{{ID: "A", Stack: 950, Bet: 50}, {ID: "B", Stack: 900, Bet: 100}}, 100,
				[]streetStep{
					{"A", 50, []ActionType{Fold, Call, Raise, Leave}, Action{Type: Raise, Amount: tt.amount}, 0},
					{"B", 0, []ActionType{Fold, Check, Raise, Leave}, Action{Type: Check}, 0},
				})
			if err := g.runStreet(); err != nil {
				t.Fatalf("runStreet: %v", err)
			}
			assertStreetPlayers(t, g, []player{{ID: "A", Stack: 900, Bet: 100}, {ID: "B", Stack: 900, Bet: 100}})
			if g.currentBet != 100 {
				t.Errorf("fallback call raised current bet to %d", g.currentBet)
			}
			assertStreetEvents(t, *events, []EventType{StreetStarted, ActionRequested, ActionValid, ActionRequested, ActionValid, StreetEnded})
		})
	}
}

func TestRunStreetAllInRaise(t *testing.T) {
	g, events := newStreetTestGame(t,
		[]player{{ID: "A", Stack: 100, Bet: 50}, {ID: "B", Stack: 950, Bet: 50}}, 50,
		[]streetStep{
			{"A", 0, []ActionType{Fold, Check, Raise, Leave}, Action{Type: Raise, Amount: 100}, 0},
			{"B", 100, []ActionType{Fold, Call, Leave}, Action{Type: Call}, 0},
		})
	if err := g.runStreet(); err != nil {
		t.Fatalf("runStreet: %v", err)
	}
	assertStreetPlayers(t, g, []player{{ID: "A", Bet: 150}, {ID: "B", Stack: 850, Bet: 150}})
	if g.currentBet != 150 || !g.runToShowdown || g.onePlayerLeft {
		t.Error("expected matched all-in raise followed by runout")
	}
	assertStreetEvents(t, *events, []EventType{StreetStarted, ActionRequested, ActionValid, ActionRequested, ActionValid, RunToShowdown, StreetEnded})
}

func TestRunStreetAlreadySettled(t *testing.T) {
	g, events := newStreetTestGame(t,
		[]player{{ID: "A", Stack: 950, Bet: 50, Acted: true}, {ID: "B", Stack: 950, Bet: 50, Acted: true}}, 50, nil)
	g.additionalRaiseCount = 1
	if err := g.runStreet(); err != nil {
		t.Fatalf("runStreet: %v", err)
	}
	assertStreetPlayers(t, g, []player{{ID: "A", Stack: 950, Bet: 50}, {ID: "B", Stack: 950, Bet: 50}})
	if g.additionalRaiseCount != 0 || g.onePlayerLeft || g.runToShowdown {
		t.Error("settled street did not reset normally")
	}
	assertStreetEvents(t, *events, []EventType{StreetStarted, StreetEnded})
}

func TestRunStreetOneRemainingPlayerBeforeInput(t *testing.T) {
	g, events := newStreetTestGame(t,
		[]player{{ID: "A", Bet: 100}, {ID: "F", Stack: 950, Bet: 50, Folded: true}}, 100, nil)
	if err := g.runStreet(); err != nil {
		t.Fatalf("runStreet: %v", err)
	}
	assertStreetPlayers(t, g, []player{{ID: "A", Stack: 150, Bet: 100}, {ID: "F", Stack: 950, Bet: 50, Folded: true}})
	if !g.onePlayerLeft || g.runToShowdown {
		t.Error("lone all-in player should receive the pot, not trigger runout")
	}
	assertStreetEvents(t, *events, []EventType{StreetStarted, OnePlayerLeft, ChipsAwarded, StreetEnded})
}

func TestRunStreetLastPlayerPayout(t *testing.T) {
	for _, action := range []ActionType{Fold, Leave} {
		t.Run(action.String(), func(t *testing.T) {
			legal := []ActionType{Fold, Check, Raise, Leave}
			g, events := newStreetTestGame(t,
				[]player{{ID: "A", Stack: 950, Bet: 50}, {ID: "B", Stack: 950, Bet: 50}, {ID: "C", Stack: 950, Bet: 50}}, 50,
				[]streetStep{{"A", 0, legal, Action{Type: action}, 0}, {"B", 0, legal, Action{Type: action}, 0}})
			if err := g.runStreet(); err != nil {
				t.Fatalf("runStreet: %v", err)
			}
			assertStreetPlayers(t, g, []player{
				{ID: "A", Stack: 950, Bet: 50, Folded: true, Left: action == Leave},
				{ID: "B", Stack: 950, Bet: 50, Folded: true, Left: action == Leave},
				{ID: "C", Stack: 1100, Bet: 50},
			})
			if !g.onePlayerLeft || g.runToShowdown {
				t.Error("expected one player left, not runout")
			}
			assertStreetEvents(t, *events, []EventType{StreetStarted, ActionRequested, ActionValid, ActionRequested, ActionValid, OnePlayerLeft, ChipsAwarded, StreetEnded})
			for _, e := range *events {
				if e.Type == ChipsAwarded {
					want := []PotState{{EligiblePlayerIDs: []string{"C"}, Amount: 150, WinnerIDs: []string{"C"}}}
					if e.Pots == nil || !reflect.DeepEqual(*e.Pots, want) {
						t.Fatalf("unexpected payout pots: %v", e.Pots)
					}
					if e.Players[2].Stack != 1100 {
						t.Error("payout event must contain the credited stack")
					}
				}
			}
			// Subsequent streets must not request input or pay the winner again.
			count := len(*events)
			if err := g.runStreet(); err != nil {
				t.Fatalf("subsequent street: %v", err)
			}
			if len(*events) != count || g.players[2].Stack != 1100 {
				t.Error("subsequent street emitted events or changed the winner's stack")
			}
		})
	}
}

func TestRunStreetRejectsInvalidAction(t *testing.T) {
	for _, action := range []ActionType{Check, Raise, ActionType(255)} {
		t.Run(action.String(), func(t *testing.T) {
			players := []player{{ID: "A", Stack: 950, Bet: 50}, {ID: "B", Bet: 100}}
			g, events := newStreetTestGame(t, players, 100,
				[]streetStep{{"A", 50, []ActionType{Fold, Call, Leave}, Action{Type: action, Amount: 100}, 0}})
			if err := g.runStreet(); !errors.Is(err, ErrGame) {
				t.Fatalf("expected ErrGame, got %v", err)
			}
			assertStreetPlayers(t, g, players)
			assertStreetEvents(t, *events, []EventType{StreetStarted, ActionRequested, ActionInvalid})
			if g.currentBet != 100 || g.onePlayerLeft || g.runToShowdown {
				t.Error("invalid action changed street state")
			}
		})
	}
}

func assertStreetPlayers(t *testing.T, g *Game, want []player) {
	t.Helper()
	if len(g.players) != len(want) {
		t.Fatalf("expected %d players, got %d", len(want), len(g.players))
	}
	for i, p := range g.players {
		if *p != want[i] {
			t.Errorf("player %d: expected %+v, got %+v", i, want[i], *p)
		}
	}
}

func assertStreetEvents(t *testing.T, events []Event, want []EventType) {
	t.Helper()
	got := make([]EventType, len(events))
	for i, e := range events {
		got[i] = e.Type
	}
	if !slices.Equal(got, want) {
		t.Errorf("expected events %v, got %v", want, got)
	}
}
