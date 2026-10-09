package pokergame

import (
	"slices"
	"testing"
)

//TODO: Check cases

func TestLegalActions(t *testing.T) {
	tests := []struct {
		name       string
		stack      int
		bet        int
		currentBet int
		raises     int
		limit      int
		opponent   player
		want       []ActionType
	}{
		{"check when matched", 950, 50, 50, 0, 1, player{Stack: 950, Bet: 50}, []ActionType{Fold, Check, Raise, Leave}},
		{"call when behind", 975, 25, 50, 0, 1, player{Stack: 950, Bet: 50}, []ActionType{Fold, Call, Raise, Leave}},
		{"exact all-in call", 25, 25, 50, 0, 1, player{Stack: 950, Bet: 50}, []ActionType{Fold, Call, Leave}},
		{"short all-in call", 10, 25, 50, 0, 1, player{Stack: 950, Bet: 50}, []ActionType{Fold, Call, Leave}},
		{"one chip beyond call permits raise", 26, 25, 50, 0, 1, player{Stack: 950, Bet: 50}, []ActionType{Fold, Call, Raise, Leave}},
		{"raise limit reached while owing", 975, 25, 50, 1, 1, player{Stack: 950, Bet: 50}, []ActionType{Fold, Call, Leave}},
		{"raise limit reached while matched", 950, 50, 50, 1, 1, player{Stack: 950, Bet: 50}, []ActionType{Fold, Check, Leave}},
		{"zero raise limit", 950, 50, 50, 0, 0, player{Stack: 950, Bet: 50}, []ActionType{Fold, Check, Leave}},
		{"below raise limit", 950, 50, 50, 1, 2, player{Stack: 950, Bet: 50}, []ActionType{Fold, Check, Raise, Leave}},
		{"only opponent all-in", 950, 50, 100, 0, 1, player{Stack: 0, Bet: 100}, []ActionType{Fold, Call, Leave}},
		{"only opponent folded", 950, 50, 50, 0, 1, player{Stack: 950, Bet: 50, Folded: true}, []ActionType{Fold, Check, Leave}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &player{ID: "A", Stack: tt.stack, Bet: tt.bet}
			g := &Game{
				options:              Options{AdditionalRaises: tt.limit},
				players:              []*player{p, &tt.opponent},
				currentBet:           tt.currentBet,
				additionalRaiseCount: tt.raises,
			}
			if got := g.legalActions(p); !slices.Equal(got, tt.want) {
				t.Errorf("expected legal actions %v, got %v", tt.want, got)
			}
		})
	}
}

func TestLegalActionsRejectsFoldedPlayer(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected a panic for a folded player")
		}
	}()
	g := &Game{}
	g.legalActions(&player{Folded: true})
}
