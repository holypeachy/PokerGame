package pokergame

import (
	"errors"
	"math"
	"slices"
	"testing"
)

//* Test cases checked

func TestBuildPots(t *testing.T) {
	type expectedPot struct {
		amount   int
		eligible []int
	}
	tests := []struct {
		name    string
		players []player
		want    []expectedPot
	}{
		{
			name:    "equal contributions",
			players: []player{{ID: "A", Bet: 100}, {ID: "B", Bet: 100}, {ID: "C", Bet: 100}},
			want:    []expectedPot{{300, []int{0, 1, 2}}},
		},
		{
			name:    "multiple contribution levels and single player excess",
			players: []player{{ID: "A", Bet: 50}, {ID: "B", Bet: 100}, {ID: "C", Bet: 200}},
			want:    []expectedPot{{150, []int{0, 1, 2}}, {100, []int{1, 2}}, {100, []int{2}}},
		},
		{
			name:    "shared highest contribution",
			players: []player{{ID: "A", Bet: 50}, {ID: "B", Bet: 100}, {ID: "C", Bet: 100}},
			want:    []expectedPot{{150, []int{0, 1, 2}}, {100, []int{1, 2}}},
		},
		{
			name:    "folded contribution below minimum",
			players: []player{{ID: "A", Bet: 100}, {ID: "B", Bet: 100}, {ID: "F", Bet: 25, Folded: true}},
			want:    []expectedPot{{225, []int{0, 1}}},
		},
		{
			name:    "folded contribution between levels",
			players: []player{{ID: "A", Bet: 50}, {ID: "B", Bet: 100}, {ID: "F", Bet: 75, Folded: true}},
			want:    []expectedPot{{150, []int{0, 1}}, {75, []int{1}}},
		},
		{
			name:    "folded contribution equals highest",
			players: []player{{ID: "A", Bet: 50}, {ID: "B", Bet: 100}, {ID: "F", Bet: 100, Folded: true}},
			want:    []expectedPot{{150, []int{0, 1}}, {100, []int{1}}},
		},
		{
			name:    "one non-folded contributor",
			players: []player{{ID: "A", Bet: 100}, {ID: "F", Bet: 50, Folded: true}},
			want:    []expectedPot{{150, []int{0}}},
		},
		{
			name:    "zero contribution excluded",
			players: []player{{ID: "A", Bet: 100}, {ID: "B", Bet: 100}, {ID: "C", Bet: 0}},
			want:    []expectedPot{{200, []int{0, 1}}},
		},
		{
			name:    "folded zero contribution excluded",
			players: []player{{ID: "A", Bet: 100}, {ID: "F", Bet: 0, Folded: true}, {ID: "B", Bet: 100}},
			want:    []expectedPot{{200, []int{0, 2}}},
		},
		{
			name: "multiple folded contributors",
			players: []player{
				{ID: "A", Bet: 50}, {ID: "F1", Bet: 25, Folded: true},
				{ID: "B", Bet: 100}, {ID: "F2", Bet: 75, Folded: true},
				{ID: "C", Bet: 200}, {ID: "F3", Bet: 200, Folded: true},
			},
			want: []expectedPot{{275, []int{0, 2, 4}}, {175, []int{2, 4}}, {200, []int{4}}},
		},
		{
			name:    "reordered contribution levels",
			players: []player{{ID: "C", Bet: 200}, {ID: "A", Bet: 50}, {ID: "B", Bet: 100}},
			want:    []expectedPot{{150, []int{0, 1, 2}}, {100, []int{0, 2}}, {100, []int{0}}},
		},
		{
			name:    "folded contributor first",
			players: []player{{ID: "F", Bet: 75, Folded: true}, {ID: "B", Bet: 100}, {ID: "A", Bet: 50}},
			want:    []expectedPot{{150, []int{1, 2}}, {75, []int{1}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			players := make([]*player, len(tt.players))
			contributed := 0
			for i := range tt.players {
				tt.players[i].Stack = 1000 - tt.players[i].Bet
				tt.players[i].Acted = true
				players[i] = &tt.players[i]
				contributed += players[i].Bet
			}
			before := slices.Clone(tt.players)

			pots, err := buildPots(players)
			if err != nil {
				t.Fatalf("buildPots: %v", err)
			}
			if len(pots) != len(tt.want) {
				t.Fatalf("expected %d pots, got %d", len(tt.want), len(pots))
			}

			total := 0
			for i, pot := range pots {
				if pot == nil {
					t.Fatalf("pot %d is nil", i)
				}
				want := tt.want[i]
				if pot.Amount != want.amount {
					t.Errorf("pot %d: expected amount %d, got %d", i, want.amount, pot.Amount)
				}
				total += pot.Amount
				if pot.Winners != nil {
					t.Errorf("pot %d: winners should be unset", i)
				}
				if len(pot.EligiblePlayers) != len(want.eligible) {
					t.Fatalf("pot %d: expected %d eligible players, got %d", i, len(want.eligible), len(pot.EligiblePlayers))
				}
				for j, p := range pot.EligiblePlayers {
					if p != &tt.players[want.eligible[j]] {
						t.Errorf("pot %d: eligible player %d is not the expected original player", i, j)
					}
					if p == nil || p.Folded {
						t.Errorf("pot %d: eligible player %d is nil or folded", i, j)
					}
				}
			}
			if total != contributed {
				t.Errorf("expected total contributions %d, got total pot value %d", contributed, total)
			}
			for i, p := range players {
				if p != &tt.players[i] {
					t.Errorf("input player reference at index %d changed", i)
				}
				if tt.players[i] != before[i] {
					t.Errorf("player %s changed: before %+v, after %+v", before[i].ID, before[i], tt.players[i])
				}
			}
		})
	}
}

func TestBuildPotsRejectsInvalidContributions(t *testing.T) {
	tests := []struct {
		name    string
		players []*player
	}{
		{"empty input", nil},
		{"all folded", []*player{{ID: "A", Bet: 100, Folded: true}, {ID: "B", Bet: 100, Folded: true}}},
		{"all bets zero", []*player{{ID: "A"}, {ID: "B"}}},
		{"negative non-folded bet", []*player{{ID: "A", Bet: -1}, {ID: "B", Bet: 100}}},
		{"negative folded bet", []*player{{ID: "F", Bet: -1, Folded: true}, {ID: "B", Bet: 100}}},
		{"only folded players contributed", []*player{{ID: "A"}, {ID: "F", Bet: 100, Folded: true}}},
		{"folded contribution exceeds all active contributions", []*player{{ID: "A", Bet: 50}, {ID: "B", Bet: 100}, {ID: "F", Bet: 150, Folded: true}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := buildPots(tt.players)
			if !errors.Is(err, ErrInternal) {
				t.Fatalf("expected ErrInternal, got %v", err)
			}
		})
	}
}

func TestGetMinBet(t *testing.T) {
	tests := []struct {
		name     string
		contribs []*contribution
		want     int
	}{
		{"minimum first", []*contribution{{Remaining: 25}, {Remaining: 50}, {Remaining: 100}}, 25},
		{"minimum middle", []*contribution{{Remaining: 100}, {Remaining: 25}, {Remaining: 50}}, 25},
		{"minimum last", []*contribution{{Remaining: 100}, {Remaining: 50}, {Remaining: 25}}, 25},
		{"equal contributions", []*contribution{{Remaining: 50}, {Remaining: 50}}, 50},
		{"ignores smaller folded contribution", []*contribution{{Remaining: 10, Folded: true}, {Remaining: 100}, {Remaining: 50}}, 50},
		{"single active contribution", []*contribution{{Remaining: 100, Folded: true}, {Remaining: 75}}, 75},
		{"maximum int contribution", []*contribution{{Remaining: math.MaxInt}}, math.MaxInt},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getMinBet(tt.contribs)
			if err != nil {
				t.Fatalf("getMinBet: %v", err)
			}
			if got != tt.want {
				t.Errorf("expected minimum %d, got %d", tt.want, got)
			}
		})
	}
}

func TestGetMinBetRequiresNonFoldedContributor(t *testing.T) {
	tests := []struct {
		name     string
		contribs []*contribution
	}{
		{"empty input", nil},
		{"all folded", []*contribution{{Remaining: 50, Folded: true}, {Remaining: 100, Folded: true}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := getMinBet(tt.contribs)
			if !errors.Is(err, ErrInternal) {
				t.Fatalf("expected ErrInternal, got %v", err)
			}
		})
	}
}

func TestSplitPotsEmpty(t *testing.T) {
	pots, err := splitPots(nil)
	if err != nil {
		t.Fatalf("splitPots: %v", err)
	}
	if len(pots) != 0 {
		t.Fatalf("expected no pots, got %d", len(pots))
	}
}
