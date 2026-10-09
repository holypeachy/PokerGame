package pokergame

import (
	"errors"
	"testing"

	"pokeralgo"
)

//TODO: Check cases

func playerTestValue() player {
	return player{
		ID: "A", Stack: 100, Bet: 25,
		HoleCards: pokeralgo.HoleCards{
			First:  pokeralgo.MustCard(14, pokeralgo.Spades, true),
			Second: pokeralgo.MustCard(13, pokeralgo.Hearts, true),
		},
	}
}

func TestPlayerConstruction(t *testing.T) {
	for _, factory := range []struct {
		name string
		new  func(string, int, pokeralgo.Card, pokeralgo.Card) (*player, error)
	}{
		{"newPlayer", newPlayer},
		{"newPlayerFromSpec", func(id string, stack int, first, second pokeralgo.Card) (*player, error) {
			return newPlayerFromSpec(NewPlayerSpec(id), stack, first, second)
		}},
	} {
		t.Run(factory.name, func(t *testing.T) {
			first := pokeralgo.MustCard(14, pokeralgo.Spades, true)
			for _, tt := range []struct {
				name    string
				second  pokeralgo.Card
				wantErr bool
			}{
				{"distinct ranks", pokeralgo.MustCard(13, pokeralgo.Spades, true), false},
				{"same rank different suit", pokeralgo.MustCard(14, pokeralgo.Hearts, true), false},
				{"duplicate", first, true},
				{"duplicate with different marker", pokeralgo.MustCard(14, pokeralgo.Spades, false), true},
			} {
				t.Run(tt.name, func(t *testing.T) {
					got, err := factory.new("A", 1000, first, tt.second)
					if tt.wantErr {
						if !errors.Is(err, ErrInternal) || got != nil {
							t.Fatalf("expected nil player and ErrInternal, got %v, %v", got, err)
						}
						return
					}
					if err != nil || got == nil {
						t.Fatalf("construction: player %v, error %v", got, err)
					}
					want := player{ID: "A", Stack: 1000, HoleCards: pokeralgo.HoleCards{First: first, Second: tt.second}}
					if *got != want {
						t.Errorf("expected %#v, got %#v", want, *got)
					}
				})
			}
		})
	}
}

func TestPlayerChipContributions(t *testing.T) {
	for _, operation := range []struct {
		name      string
		apply     func(*player, int) error
		setsActed bool
	}{
		{"commitChips", (*player).commitChips, true},
		{"postBlind", (*player).postBlind, false},
	} {
		t.Run(operation.name, func(t *testing.T) {
			for _, tt := range []struct {
				name      string
				amounts   []int
				acted     bool
				wantStack int
				wantBet   int
			}{
				{"zero", []int{0}, false, 100, 25},
				{"partial", []int{30}, false, 70, 55},
				{"exact stack", []int{100}, false, 0, 125},
				{"over stack", []int{150}, false, 0, 125},
				{"repeated", []int{20, 30}, false, 50, 75},
				{"repeated reaches all in", []int{60, 80}, false, 0, 125},
				{"empty stack", []int{100, 50}, false, 0, 125},
				{"preserve acted", []int{30}, true, 70, 55},
				{"preserve acted over stack", []int{150}, true, 0, 125},
			} {
				t.Run(tt.name, func(t *testing.T) {
					p := playerTestValue()
					p.Acted = tt.acted
					want := p
					want.Stack, want.Bet = tt.wantStack, tt.wantBet
					want.Acted = tt.acted || operation.setsActed
					for _, amount := range tt.amounts {
						if err := operation.apply(&p, amount); err != nil {
							t.Fatal(err)
						}
					}
					if p != want {
						t.Errorf("expected %#v, got %#v", want, p)
					}
				})
			}
		})
	}
}

func TestPlayerCredit(t *testing.T) {
	for _, amount := range []int{0, 50} {
		p := playerTestValue()
		p.Acted, p.Folded, p.Left = true, true, true
		want := p
		want.Stack += amount
		if err := p.credit(amount); err != nil {
			t.Fatal(err)
		}
		if p != want {
			t.Errorf("credit %d: expected %#v, got %#v", amount, want, p)
		}
	}
}

func TestPlayerActions(t *testing.T) {
	t.Run("fold", func(t *testing.T) {
		p := playerTestValue()
		want := p
		want.Folded = true
		if err := p.fold(); err != nil {
			t.Fatal(err)
		}
		if p != want {
			t.Errorf("expected %#v, got %#v", want, p)
		}
	})
	t.Run("leave", func(t *testing.T) {
		p := playerTestValue()
		want := p
		want.Left = true
		if err := p.leave(); err != nil {
			t.Fatal(err)
		}
		if p != want {
			t.Errorf("expected %#v, got %#v", want, p)
		}
	})
	t.Run("check", func(t *testing.T) {
		p := playerTestValue()
		want := p
		want.Acted = true
		for range 2 {
			p.check()
			if p != want {
				t.Errorf("expected %#v, got %#v", want, p)
			}
		}
	})
}

func TestPlayerDeal(t *testing.T) {
	p := playerTestValue()
	p.Acted, p.Folded, p.Left = true, true, true
	first := pokeralgo.MustCard(2, pokeralgo.Clubs, true)
	second := pokeralgo.MustCard(2, pokeralgo.Diamonds, true)
	want := p
	want.HoleCards = pokeralgo.HoleCards{First: first, Second: second}
	if err := p.deal(first, second); err != nil {
		t.Fatal(err)
	}
	if p != want {
		t.Errorf("expected %#v, got %#v", want, p)
	}
}

func TestPlayerRejectedOperationsPreserveState(t *testing.T) {
	for _, tt := range []struct {
		name  string
		apply func(*player) error
	}{
		{"negative commitment", func(p *player) error { return p.commitChips(-1) }},
		{"negative blind", func(p *player) error { return p.postBlind(-1) }},
		{"negative credit", func(p *player) error { return p.credit(-1) }},
		{"already folded", (*player).fold},
		{"already left", (*player).leave},
		{"duplicate cards", func(p *player) error {
			card := pokeralgo.MustCard(2, pokeralgo.Clubs, true)
			return p.deal(card, card)
		}},
		{"duplicate cards with different markers", func(p *player) error {
			return p.deal(pokeralgo.MustCard(2, pokeralgo.Clubs, true), pokeralgo.MustCard(2, pokeralgo.Clubs, false))
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := playerTestValue()
			p.Folded, p.Left = true, true
			before := p
			if err := tt.apply(&p); !errors.Is(err, ErrInternal) {
				t.Errorf("expected ErrInternal, got %v", err)
			}
			if p != before {
				t.Errorf("rejected operation changed player: before %#v, after %#v", before, p)
			}
		})
	}
}

func TestPlayerResets(t *testing.T) {
	for _, tt := range []struct {
		name  string
		reset func(*player)
	}{
		{"street", (*player).resetForNextStreet},
		{"hand", (*player).resetForNextHand},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := playerTestValue()
			p.Acted, p.Folded, p.Left = true, true, true
			want := p
			want.Acted = false
			if tt.name == "hand" {
				want.Bet, want.Folded = 0, false
			}
			for range 2 {
				tt.reset(&p)
				if p != want {
					t.Errorf("expected %#v, got %#v", want, p)
				}
			}
		})
	}
}
