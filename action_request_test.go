package pokergame

import (
	"reflect"
	"slices"
	"testing"

	"pokeralgo"
)

//TODO: Check cases

func TestActiobbestPayload(t *testing.T) {
	g := eventTestGame()
	g.currentBet = 55
	r := g.actionRequest()
	wantPlayers := []PlayerState{
		{ID: "A", Stack: 100, Bet: 25, HoleCards: g.players[0].HoleCards},
		{ID: "B", Stack: 70, Bet: 55, Folded: true, Left: true, HoleCards: g.players[1].HoleCards},
	}
	if !reflect.DeepEqual(r.PlayerStates, wantPlayers) {
		t.Errorf("expected player states %+v, got %+v", wantPlayers, r.PlayerStates)
	}
	if r.PlayerToAct == nil || *r.PlayerToAct != wantPlayers[0] {
		t.Errorf("expected acting player A, got %v", r.PlayerToAct)
	}
	if r.ToCall != 30 || !slices.Equal(r.LegalActions, []ActionType{Fold, Call, Leave}) {
		t.Errorf("unexpected ToCall=%d LegalActions=%v", r.ToCall, r.LegalActions)
	}
	if !slices.Equal(r.CommunityCards, g.board) {
		t.Error("request board differs from game board")
	}
}

func TestActionRequestSnapshotIsolation(t *testing.T) {
	t.Run("engine mutation preserves request", func(t *testing.T) {
		g := eventTestGame()
		g.currentBet = 55
		request := g.actionRequest()
		want := request
		want.PlayerStates = slices.Clone(request.PlayerStates)
		want.CommunityCards = slices.Clone(request.CommunityCards)
		want.LegalActions = slices.Clone(request.LegalActions)
		actingPlayer := *request.PlayerToAct
		want.PlayerToAct = &actingPlayer
		for _, p := range g.players {
			*p = player{ID: "changed", Stack: -1}
		}
		for i := range g.board {
			g.board[i] = pokeralgo.Card{}
		}
		g.actingPlayerIndex, g.currentBet = 1, 999
		if !reflect.DeepEqual(request, want) {
			t.Errorf("engine mutation changed request: expected %+v, got %+v", want, request)
		}
	})
	t.Run("request mutation preserves engine and other requests", func(t *testing.T) {
		g := eventTestGame()
		g.currentBet = 55
		wantGame := eventTestGame()
		wantGame.currentBet = 55
		want := wantGame.actionRequest()
		request, previous := g.actionRequest(), g.actionRequest()
		*request.PlayerToAct = PlayerState{ID: "changed", Stack: -1}
		for i := range request.PlayerStates {
			request.PlayerStates[i] = PlayerState{ID: "changed", Stack: -1}
		}
		for i := range request.CommunityCards {
			request.CommunityCards[i] = pokeralgo.Card{}
		}
		for i := range request.LegalActions {
			request.LegalActions[i] = Check
		}
		request.ToCall = -1
		if !reflect.DeepEqual(g, wantGame) {
			t.Error("request mutation changed engine state")
		}
		if !reflect.DeepEqual(previous, want) || !reflect.DeepEqual(g.actionRequest(), want) {
			t.Error("request mutation changed an earlier or later request")
		}
	})
}
