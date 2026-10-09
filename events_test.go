package pokergame

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"testing"

	"pokeralgo"
)

//TODO: Check cases

func eventTestGame() *Game {
	a, b := playerTestValue(), playerTestValue()
	b.ID, b.Stack, b.Bet, b.Folded, b.Left = "B", 70, 55, true, true
	b.HoleCards = pokeralgo.HoleCards{
		First:  pokeralgo.MustCard(10, pokeralgo.Clubs, true),
		Second: pokeralgo.MustCard(9, pokeralgo.Diamonds, true),
	}
	return &Game{
		options: Options{EventVerbosity: EventsAll}, players: []*player{&a, &b},
		handNumber: 3, street: Flop, dealerIndex: 0, smallBlindIndex: 0, bigBlindIndex: 1,
		board: []pokeralgo.Card{
			pokeralgo.MustCard(2, pokeralgo.Clubs, false),
			pokeralgo.MustCard(5, pokeralgo.Diamonds, false),
			pokeralgo.MustCard(8, pokeralgo.Hearts, false),
		},
	}
}

func TestEventPayloadsFilteringAndSnapshots(t *testing.T) {
	board := eventTestGame().board
	blinds := BlindIndices{Dealer: 0, SmallBlind: 0, BigBlind: 1}
	id, amount, toCall, action := "A", 20, 30, Raise
	legal := []ActionType{Fold, Call, Raise, Leave}
	errEvent := fmt.Errorf("%w: test failure", ErrInternal)
	created := []PotState{{EligiblePlayerIDs: []string{"A", "B"}, Amount: 80}}
	awarded := []PotState{{EligiblePlayerIDs: []string{"A", "B"}, Amount: 80, WinnerIDs: []string{"A"}}}
	loneWinner := []PotState{{EligiblePlayerIDs: []string{"A"}, Amount: 80, WinnerIDs: []string{"A"}}}
	for _, tt := range []struct {
		kind    EventType
		level   EventVerbosity
		emit    func(*Game, ActionRequest, []*Pot)
		payload Event
	}{
		{GameStarted, EventsCore, func(g *Game, _ ActionRequest, _ []*Pot) { g.emitGameStarted() }, Event{}},
		{HandStarted, EventsAll, func(g *Game, _ ActionRequest, _ []*Pot) { g.emitHandStarted() }, Event{Board: board}},
		{BlindsAdvanced, EventsAll, func(g *Game, _ ActionRequest, _ []*Pot) { g.emitBlindsAdvanced() }, Event{BlindIndices: &blinds}},
		{BlindsPosted, EventsCore, func(g *Game, _ ActionRequest, _ []*Pot) { g.emitBlindsPosted() }, Event{BlindIndices: &blinds}},
		{HoleCardsDealt, EventsAll, func(g *Game, _ ActionRequest, _ []*Pot) { g.emitHoleCardsDealt() }, Event{}},
		{StreetStarted, EventsCore, func(g *Game, _ ActionRequest, _ []*Pot) { g.emitStreetStarted() }, Event{Board: board}},
		{RunToShowdown, EventsAll, func(g *Game, _ ActionRequest, _ []*Pot) { g.emitRunToShowdown() }, Event{}},
		{ActionRequested, EventsCore, func(g *Game, r ActionRequest, _ []*Pot) { g.emitActionRequested(r) },
			Event{Board: board, PlayerID: &id, Amount: &toCall, LegalActions: legal}},
		{ActionInvalid, EventsCore, func(g *Game, r ActionRequest, _ []*Pot) { g.emitActionInvalid(r, Action{Type: Raise, Amount: 20}) },
			Event{Board: board, PlayerID: &id, Amount: &toCall, LegalActions: legal, ActionType: &action}},
		{ActionValid, EventsCore, func(g *Game, _ ActionRequest, _ []*Pot) {
			g.emitActionValid(g.players[0].ID, Action{Type: Raise, Amount: 20})
		},
			Event{Board: board, PlayerID: &id, Amount: &amount, ActionType: &action}},
		{OnePlayerLeft, EventsAll, func(g *Game, _ ActionRequest, _ []*Pot) { g.emitOnePlayerLeft(g.players[0].ID, 80) },
			Event{Board: board, PlayerID: &id, Pots: &loneWinner}},
		{StreetEnded, EventsAll, func(g *Game, _ ActionRequest, _ []*Pot) { g.emitStreetEnded() }, Event{Board: board}},
		{PotsCreated, EventsAll, func(g *Game, _ ActionRequest, p []*Pot) { g.emitPotsCreated(p) }, Event{Board: board, Pots: &created}},
		{WinnersDetermined, EventsAll, func(g *Game, _ ActionRequest, p []*Pot) { g.emitWinnersDetermined(p) }, Event{Board: board, Pots: &awarded}},
		{ChipsAwarded, EventsCore, func(g *Game, _ ActionRequest, p []*Pot) { g.emitChipsAwarded(p) }, Event{Board: board, Pots: &awarded}},
		{ChipsAwarded, EventsCore, func(g *Game, _ ActionRequest, _ []*Pot) { g.emitChipsOnePlayerLeft(g.players[0].ID, 80) }, Event{Board: board, Pots: &loneWinner}},
		{HandEnded, EventsAll, func(g *Game, _ ActionRequest, _ []*Pot) { g.emitHandEnded() }, Event{Board: board}},
		{PlayerBusted, EventsCore, func(g *Game, _ ActionRequest, _ []*Pot) { g.emitPlayerBusted(g.players[0].ID) }, Event{PlayerID: &id}},
		{PlayerLeft, EventsCore, func(g *Game, _ ActionRequest, _ []*Pot) { g.emitPlayerLeft(g.players[0].ID) }, Event{PlayerID: &id}},
		{GameEnded, EventsCore, func(g *Game, _ ActionRequest, _ []*Pot) { g.emitGameEnded(g.players[0].ID) }, Event{PlayerID: &id}},
		{ErrorState, EventsOff, func(g *Game, _ ActionRequest, _ []*Pot) { g.emitErr(errEvent) }, Event{Err: errEvent}},
	} {
		t.Run(tt.kind.String(), func(t *testing.T) {
			want := tt.payload
			want.Type, want.HandNumber, want.Street = tt.kind, 3, Flop
			if tt.kind != ErrorState {
				for _, p := range eventTestGame().players {
					want.Players = append(want.Players, PlayerState{ID: p.ID, Stack: p.Stack, Bet: p.Bet,
						Folded: p.Folded, Left: p.Left, HoleCards: p.HoleCards})
				}
			}
			for _, level := range []EventVerbosity{EventsOff, EventsCore, EventsAll} {
				t.Run(fmt.Sprintf("level %d", level), func(t *testing.T) {
					g := eventTestGame()
					g.SetEventsVerbosity(level)
					request := ActionRequest{ToCall: 30, LegalActions: slices.Clone(legal)}
					pots := []*Pot{{EligiblePlayers: slices.Clone(g.players), Amount: 80, Winners: []*player{g.players[0]}}}
					var events []Event
					g.onEvent = func(e Event) { events = append(events, e) }
					tt.emit(g, request, pots)
					if level < tt.level {
						if len(events) != 0 {
							t.Fatalf("filtered event emitted: %v", events)
						}
						if g.newBaseEvent(tt.kind) != nil {
							t.Error("filtered event still constructed a snapshot")
						}
						return
					}
					if len(events) != 1 {
						t.Fatalf("expected one event, got %d", len(events))
					}
					if !reflect.DeepEqual(events[0], want) {
						t.Errorf("expected %#v, got %#v", want, events[0])
					}
				})
			}
			t.Run("source mutation preserves snapshot", func(t *testing.T) {
				g := eventTestGame()
				request := ActionRequest{ToCall: 30, LegalActions: slices.Clone(legal)}
				pots := []*Pot{{EligiblePlayers: slices.Clone(g.players), Amount: 80, Winners: []*player{g.players[0]}}}
				var got Event
				g.onEvent = func(e Event) { got = e }
				tt.emit(g, request, pots)
				for _, p := range g.players {
					*p = player{ID: "changed", Stack: -1}
				}
				for i := range g.board {
					g.board[i] = pokeralgo.Card{}
				}
				g.handNumber, g.street = 99, River
				g.dealerIndex, g.smallBlindIndex, g.bigBlindIndex = 9, 9, 9
				request.ToCall, request.LegalActions[0] = 999, Check
				pots[0].Amount = -1
				pots[0].EligiblePlayers[0], pots[0].Winners[0] = &player{ID: "replacement"}, &player{ID: "replacement"}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("source mutation changed snapshot: expected %#v, got %#v", want, got)
				}
			})
			t.Run("snapshot mutation preserves source and other snapshots", func(t *testing.T) {
				g := eventTestGame()
				request := ActionRequest{ToCall: 30, LegalActions: slices.Clone(legal)}
				pots := []*Pot{{EligiblePlayers: slices.Clone(g.players), Amount: 80, Winners: []*player{g.players[0]}}}
				var events []Event
				g.onEvent = func(e Event) { events = append(events, e) }
				tt.emit(g, request, pots)
				tt.emit(g, request, pots)
				mutateEventSnapshot(&events[0])
				tt.emit(g, request, pots)
				for _, i := range []int{1, 2} {
					if !reflect.DeepEqual(events[i], want) {
						t.Errorf("mutating event affected snapshot %d: %#v", i, events[i])
					}
				}
				original := eventTestGame()
				if !reflect.DeepEqual(g.players, original.players) || !slices.Equal(g.board, original.board) ||
					g.dealerIndex != original.dealerIndex || g.smallBlindIndex != original.smallBlindIndex || g.bigBlindIndex != original.bigBlindIndex ||
					g.handNumber != original.handNumber || g.street != original.street {
					t.Error("mutating event changed engine state")
				}
				if request.ToCall != 30 || !slices.Equal(request.LegalActions, legal) {
					t.Error("mutating event changed action request")
				}
				if pots[0].Amount != 80 || len(pots[0].EligiblePlayers) != 2 || pots[0].EligiblePlayers[0] != g.players[0] ||
					pots[0].EligiblePlayers[1] != g.players[1] || len(pots[0].Winners) != 1 || pots[0].Winners[0] != g.players[0] {
					t.Error("mutating event changed source pot")
				}
			})
			t.Run("nil callback", func(t *testing.T) {
				g := eventTestGame()
				tt.emit(g, ActionRequest{ToCall: 30, LegalActions: slices.Clone(legal)},
					[]*Pot{{EligiblePlayers: g.players, Amount: 80, Winners: []*player{g.players[0]}}})
			})
		})
	}
}

func mutateEventSnapshot(e *Event) {
	e.Type, e.HandNumber, e.Street = ErrorState, 99, River
	for i := range e.Players {
		e.Players[i] = PlayerState{ID: "changed", Stack: -1}
	}
	for i := range e.Board {
		e.Board[i] = pokeralgo.Card{}
	}
	if e.BlindIndices != nil {
		*e.BlindIndices = BlindIndices{Dealer: 9, SmallBlind: 9, BigBlind: 9}
	}
	if e.PlayerID != nil {
		*e.PlayerID = "changed"
	}
	if e.Amount != nil {
		*e.Amount = -1
	}
	if e.ActionType != nil {
		*e.ActionType = Check
	}
	for i := range e.LegalActions {
		e.LegalActions[i] = Check
	}
	if e.Pots != nil {
		for i := range *e.Pots {
			p := &(*e.Pots)[i]
			p.Amount = -1
			for j := range p.EligiblePlayerIDs {
				p.EligiblePlayerIDs[j] = "changed"
			}
			for j := range p.WinnerIDs {
				p.WinnerIDs[j] = "changed"
			}
		}
	}
	e.Err = nil
}

func TestSetEventsVerbosity(t *testing.T) {
	g := eventTestGame()
	var events []Event
	g.onEvent = func(e Event) { events = append(events, e) }
	for _, level := range []EventVerbosity{EventsOff, EventsCore, EventsAll, EventsOff} {
		g.SetEventsVerbosity(level)
		g.emitGameStarted()
		g.emitHandStarted()
	}
	assertStreetEvents(t, events, []EventType{GameStarted, GameStarted, HandStarted})
}

func TestShowdownStartedEventLevel(t *testing.T) {
	g := eventTestGame()
	for _, level := range []EventVerbosity{EventsOff, EventsCore, EventsAll} {
		g.SetEventsVerbosity(level)
		e := g.newBaseEvent(ShowdownStarted)
		if (e != nil) != (level == EventsAll) {
			t.Errorf("unexpected ShowdownStarted visibility at level %d", level)
		}
	}
}

func TestPlayEmitsReturnedErrorOnce(t *testing.T) {
	for _, duringHand := range []bool{false, true} {
		name := "setup failure"
		if duringHand {
			name = "hand failure"
		}
		t.Run(name, func(t *testing.T) {
			var events []Event
			g := New(Options{BuyIn: 1000, BigBlind: 50, AdditionalRaises: 1, EventVerbosity: EventsOff}, nil,
				func(e Event) { events = append(events, e) })
			wantErr := ErrInternal
			if duringHand {
				g.actionSource = &streetActionSource{t: t, game: g, steps: []streetStep{
					{"A", 25, []ActionType{Fold, Call, Raise, Leave}, Action{Type: Check}, 0},
				}}
				wantErr = ErrGame
			}
			err := g.Play([]PlayerSpec{NewPlayerSpec("A"), NewPlayerSpec("B")})
			if !errors.Is(err, wantErr) {
				t.Fatalf("expected %v, got %v", wantErr, err)
			}
			if len(events) != 1 || events[0].Type != ErrorState || events[0].Err != err {
				t.Fatalf("expected exactly one event containing the returned error, got %v", events)
			}
			wantHand, wantStreet := 0, NoStreet
			if duringHand {
				wantHand, wantStreet = 1, Preflop
			}
			if events[0].HandNumber != wantHand || events[0].Street != wantStreet {
				t.Error("error event has incorrect hand context")
			}
		})
	}
}
