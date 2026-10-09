package pokergame

import (
	"errors"
	"slices"
	"testing"

	"pokeralgo"
)

//TODO: Check cases

func TestRemoveBustedLeftAdjustsDealer(t *testing.T) {
	tests := []struct {
		name       string
		dealer     int
		busted     []int
		left       []int
		wantIDs    []string
		wantDealer int
		wantNext   string
	}{
		{"no removals", 2, nil, nil, []string{"A", "B", "C", "D", "E"}, 2, "D"},
		{"before dealer", 2, []int{0}, nil, []string{"B", "C", "D", "E"}, 1, "D"},
		{"dealer removed", 2, nil, []int{2}, []string{"A", "B", "D", "E"}, 1, "D"},
		{"after dealer", 2, []int{3}, nil, []string{"A", "B", "C", "E"}, 2, "E"},
		{"multiple before dealer", 2, []int{0, 1}, nil, []string{"C", "D", "E"}, 0, "D"},
		{"before and including dealer", 2, []int{0}, []int{1, 2}, []string{"D", "E"}, -1, "D"},
		{"first dealer removed", 0, []int{0}, nil, []string{"B", "C", "D", "E"}, -1, "B"},
		{"last dealer wraps", 4, nil, nil, []string{"A", "B", "C", "D", "E"}, 4, "A"},
		{"last dealer removed wraps", 4, []int{4}, []int{0}, []string{"B", "C", "D"}, 2, "B"},
		{"multiple after dealer wraps", 2, []int{3}, []int{4}, []string{"A", "B", "C"}, 2, "A"},
		{"busted and left removed once", 2, []int{1}, []int{1}, []string{"A", "C", "D", "E"}, 1, "D"},
		{"one survivor", 2, []int{0, 2}, []int{1, 4}, []string{"D"}, -1, ""},
		{"no survivors", 2, []int{0, 1, 2, 3, 4}, nil, nil, -1, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Game{dealerIndex: tt.dealer, options: Options{EventVerbosity: EventsAll}}
			for i, id := range []string{"A", "B", "C", "D", "E"} {
				p := &player{ID: id, Stack: 1000, Bet: 50, Acted: true, Folded: i == 3}
				if slices.Contains(tt.busted, i) {
					p.Stack = 0
				}
				p.Left = slices.Contains(tt.left, i)
				g.players = append(g.players, p)
			}
			original := slices.Clone(g.players)
			before := make([]player, len(original))
			for i, p := range original {
				before[i] = *p
			}
			var removedIDs []string
			g.onEvent = func(e Event) {
				if e.Type != PlayerBusted && e.Type != PlayerLeft {
					t.Fatalf("unexpected removal event %s", e.Type)
				}
				removedIDs = append(removedIDs, *e.PlayerID)
			}
			g.removeBustedLeft()
			if len(g.players) != len(tt.wantIDs) {
				t.Fatalf("expected survivors %v, got %d players", tt.wantIDs, len(g.players))
			}
			for i, id := range tt.wantIDs {
				originalIndex := slices.IndexFunc(original, func(p *player) bool { return p.ID == id })
				if g.players[i] != original[originalIndex] {
					t.Errorf("survivor %d: expected original player %s", i, id)
				}
			}
			var wantRemoved []string
			for i, p := range original {
				if *p != before[i] {
					t.Errorf("removal modified player %s", p.ID)
				}
				if !slices.Contains(tt.wantIDs, p.ID) {
					wantRemoved = append(wantRemoved, p.ID)
				}
			}
			if !slices.Equal(removedIDs, wantRemoved) {
				t.Errorf("expected removal events for %v, got %v", wantRemoved, removedIDs)
			}
			if g.dealerIndex != tt.wantDealer {
				t.Errorf("expected adjusted dealer index %d, got %d", tt.wantDealer, g.dealerIndex)
			}
			if tt.wantNext != "" {
				g.advanceBlinds()
				if id := g.players[g.dealerIndex].ID; id != tt.wantNext {
					t.Errorf("expected next dealer %s, got %s", tt.wantNext, id)
				}
			}
		})
	}
}

func TestAdvanceBlinds(t *testing.T) {
	tests := []struct {
		name       string
		count      int
		dealer     int
		wantBlinds BlindIndices
		wantActor  int
	}{
		{"five players first hand", 5, -1, BlindIndices{Dealer: 0, SmallBlind: 1, BigBlind: 2}, 3},
		{"five players wrap", 5, 4, BlindIndices{Dealer: 0, SmallBlind: 1, BigBlind: 2}, 3},
		{"blinds cross end of table", 5, 2, BlindIndices{Dealer: 3, SmallBlind: 4, BigBlind: 0}, 1},
		{"three players", 3, -1, BlindIndices{Dealer: 0, SmallBlind: 1, BigBlind: 2}, 0},
		{"heads up first hand", 2, -1, BlindIndices{Dealer: 0, SmallBlind: 0, BigBlind: 1}, 0},
		{"heads up alternate dealer", 2, 0, BlindIndices{Dealer: 1, SmallBlind: 1, BigBlind: 0}, 1},
		{"heads up wrap", 2, 1, BlindIndices{Dealer: 0, SmallBlind: 0, BigBlind: 1}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &Game{players: make([]*player, tt.count), dealerIndex: tt.dealer, actingPlayerIndex: 99}
			g.advanceBlinds()
			got := BlindIndices{Dealer: g.dealerIndex, SmallBlind: g.smallBlindIndex, BigBlind: g.bigBlindIndex}
			if got != tt.wantBlinds || g.actingPlayerIndex != tt.wantActor {
				t.Errorf("expected blinds %+v actor %d, got %+v actor %d", tt.wantBlinds, tt.wantActor, got, g.actingPlayerIndex)
			}
		})
	}
}

func TestResetForNextHand(t *testing.T) {
	for _, runout := range []bool{false, true} {
		name := "one player left"
		if runout {
			name = "showdown runout"
		}
		t.Run(name, func(t *testing.T) {
			g := &Game{handNumber: 7, street: River, additionalRaiseCount: 1, onePlayerLeft: !runout, runToShowdown: runout}
			g.players = []*player{
				{ID: "A", Stack: 1250, Bet: 100, Acted: true, HoleCards: pokeralgo.HoleCards{
					First: pokeralgo.MustCard(14, pokeralgo.Spades, true), Second: pokeralgo.MustCard(13, pokeralgo.Hearts, true),
				}},
				{ID: "B", Stack: 750, Bet: 50, Acted: true, Folded: true, HoleCards: pokeralgo.HoleCards{
					First: pokeralgo.MustCard(2, pokeralgo.Clubs, true), Second: pokeralgo.MustCard(3, pokeralgo.Diamonds, true),
				}},
			}
			original := slices.Clone(g.players)
			g.resetForNextHand()
			for i, p := range g.players {
				wantStack := []int{1250, 750}[i]
				wantID := []string{"A", "B"}[i]
				if p != original[i] || p.ID != wantID || p.Stack != wantStack {
					t.Errorf("player %d identity or stack changed", i)
				}
				if p.Bet != 0 || p.Acted || p.Folded || p.Left {
					t.Errorf("player %s retains hand state: %+v", p.ID, *p)
				}
				if p.HoleCards.First.Rank != 0 || p.HoleCards.Second.Rank != 0 {
					t.Errorf("player %s retains dealt hole cards", p.ID)
				}
			}
			if g.street != NoStreet || g.onePlayerLeft || g.runToShowdown || g.additionalRaiseCount != 0 {
				t.Error("hand control state was not reset")
			}
			if g.handNumber != 7 || g.gameOver {
				t.Error("reset changed the hand number or ended the game")
			}
		})
	}
}

func TestCheckGameOver(t *testing.T) {
	for _, tt := range []struct {
		name  string
		count int
	}{
		{"no survivor", 0}, {"one winner", 1}, {"two survivors", 2}, {"five survivors", 5},
	} {
		t.Run(tt.name, func(t *testing.T) {
			g := &Game{options: Options{EventVerbosity: EventsAll}}
			for _, id := range []string{"A", "B", "C", "D", "E"}[:tt.count] {
				g.players = append(g.players, &player{ID: id, Stack: 1000})
			}
			var events []Event
			g.onEvent = func(e Event) { events = append(events, e) }
			err := g.checkGameOver()
			if tt.count == 0 {
				if !errors.Is(err, ErrInternal) {
					t.Fatalf("expected ErrInternal, got %v", err)
				}
			} else if err != nil {
				t.Fatalf("checkGameOver: %v", err)
			}
			if g.gameOver != (tt.count == 1) {
				t.Errorf("unexpected gameOver=%t for %d survivors", g.gameOver, tt.count)
			}
			if tt.count == 1 {
				assertStreetEvents(t, events, []EventType{GameEnded})
				if len(events) != 1 || events[0].PlayerID == nil || *events[0].PlayerID != "A" {
					t.Fatal("missing winner A in GameEnded")
				}
			} else if len(events) != 0 {
				t.Errorf("unexpected events: %v", events)
			}
		})
	}
}

func TestPlayHandActionOrder(t *testing.T) {
	for _, tt := range []struct {
		name     string
		ids      []string
		preflop  []string
		postflop []string
		calls    []int
	}{
		{"heads up", []string{"A", "B"}, []string{"A", "B"}, []string{"B", "A"}, []int{25, 0}},
		{"three players", []string{"A", "B", "C"}, []string{"A", "B", "C"}, []string{"B", "C", "A"}, []int{50, 25, 0}},
		{"five players", []string{"A", "B", "C", "D", "E"}, []string{"D", "E", "A", "B", "C"}, []string{"B", "C", "D", "E", "A"}, []int{50, 50, 50, 25, 0}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			checkLegal := []ActionType{Fold, Check, Raise, Leave}
			callLegal := []ActionType{Fold, Call, Raise, Leave}
			var steps []streetStep
			var wantStreets []Street
			for i, id := range tt.preflop {
				action, legal := Call, callLegal
				if tt.calls[i] == 0 {
					action, legal = Check, checkLegal
				}
				steps = append(steps, streetStep{id, tt.calls[i], legal, Action{Type: action}, 0})
				wantStreets = append(wantStreets, Preflop)
			}
			for _, street := range []Street{Flop, Turn, River} {
				for i, id := range tt.postflop {
					if street == River && i == len(tt.postflop)-1 {
						break
					}
					action := Check
					if street == River {
						action = Fold
					}
					steps = append(steps, streetStep{id, 0, checkLegal, Action{Type: action}, 0})
					wantStreets = append(wantStreets, street)
				}
			}
			var events []Event
			g := New(Options{BuyIn: 1000, BigBlind: 50, AdditionalRaises: 1, EventVerbosity: EventsAll}, nil,
				func(e Event) { events = append(events, e) })
			g.deck = pokeralgo.NewDeckWithSeed(1)
			source := &streetActionSource{t: t, game: g, steps: steps}
			g.actionSource = source
			var specs []PlayerSpec
			for _, id := range tt.ids {
				specs = append(specs, NewPlayerSpec(id))
			}
			if err := g.seatPlayers(specs); err != nil {
				t.Fatal(err)
			}
			if err := g.playHand(); err != nil {
				t.Fatal(err)
			}
			if source.next != len(steps) {
				t.Errorf("consumed %d of %d actions", source.next, len(steps))
			}
			var gotStreets []Street
			for _, e := range events {
				if e.Type == ActionRequested {
					gotStreets = append(gotStreets, e.Street)
				}
			}
			if !slices.Equal(gotStreets, wantStreets) {
				t.Errorf("expected action streets %v, got %v", wantStreets, gotStreets)
			}
		})
	}
}

func TestPlayTwoHandTransition(t *testing.T) {
	callLegal := []ActionType{Fold, Call, Raise, Leave}
	checkLegal := []ActionType{Fold, Check, Raise, Leave}
	steps := []streetStep{
		// Hand one: A leaves, B folds, C receives both blinds.
		{"A", 50, callLegal, Action{Type: Leave}, 0},
		{"B", 25, callLegal, Action{Type: Fold}, 0},
		// Hand two: B is dealer/small blind; C acts first after the flop.
		{"B", 25, callLegal, Action{Type: Call}, 0},
		{"C", 0, checkLegal, Action{Type: Check}, 0},
		{"C", 0, checkLegal, Action{Type: Check}, 0},
		{"B", 0, checkLegal, Action{Type: Check}, 0},
		{"C", 0, checkLegal, Action{Type: Check}, 0},
		{"B", 0, checkLegal, Action{Type: Check}, 0},
		{"C", 0, checkLegal, Action{Type: Leave}, 0},
	}
	var events []Event
	g := New(Options{BuyIn: 1000, BigBlind: 50, AdditionalRaises: 1, EventVerbosity: EventsAll}, nil,
		func(e Event) { events = append(events, e) })
	g.deck = pokeralgo.NewDeckWithSeed(1)
	source := &streetActionSource{t: t, game: g, steps: steps}
	g.actionSource = source
	if err := g.Play([]PlayerSpec{NewPlayerSpec("A"), NewPlayerSpec("B"), NewPlayerSpec("C")}); err != nil {
		t.Fatalf("Play: %v", err)
	}
	if source.next != len(steps) || g.handNumber != 2 || !g.gameOver {
		t.Fatalf("expected completion after two hands and %d actions; got hand %d, actions %d, gameOver %t", len(steps), g.handNumber, source.next, g.gameOver)
	}
	if len(g.players) != 1 || g.players[0].ID != "B" || g.players[0].Stack != 1025 {
		t.Fatalf("expected B as sole survivor with 1025 chips, got %v", g.players)
	}
	var milestones []EventType
	var dealtHands []int
	for _, e := range events {
		switch e.Type {
		case ChipsAwarded, HandEnded, PlayerLeft, HandStarted, GameEnded:
			milestones = append(milestones, e.Type)
		}
		if e.Type == ChipsAwarded {
			wantWinner, wantAmount, winnerIndex := "C", 75, 2
			if e.HandNumber == 2 {
				wantWinner, wantAmount, winnerIndex = "B", 100, 0
			}
			if e.Pots == nil || len(*e.Pots) != 1 || (*e.Pots)[0].Amount != wantAmount || !slices.Equal((*e.Pots)[0].WinnerIDs, []string{wantWinner}) {
				t.Errorf("hand %d: unexpected payout %v", e.HandNumber, e.Pots)
			}
			if e.Players[winnerIndex].Stack != 1025 {
				t.Errorf("hand %d: payout snapshot lacks credited stack", e.HandNumber)
			}
		}
		if e.Type == HandStarted && e.HandNumber == 2 {
			if len(e.Players) != 2 || e.Players[0].ID != "B" || e.Players[1].ID != "C" {
				t.Fatalf("unexpected second-hand roster: %v", e.Players)
			}
			if e.Street != NoStreet || len(e.Board) != 0 || e.Players[0].Stack != 975 || e.Players[1].Stack != 1025 {
				t.Errorf("incorrect second-hand starting state: %v", e)
			}
			for _, p := range e.Players {
				if p.Bet != 0 || p.Folded || p.Left || p.HoleCards.First.Rank != 0 || p.HoleCards.Second.Rank != 0 {
					t.Errorf("second hand retains player state: %+v", p)
				}
			}
		}
		if e.Type == BlindsPosted && e.HandNumber == 2 {
			want := BlindIndices{Dealer: 0, SmallBlind: 0, BigBlind: 1}
			if e.BlindIndices == nil || *e.BlindIndices != want {
				t.Errorf("unexpected heads-up blinds: %v", e.BlindIndices)
			}
			if e.Players[0].Bet != 25 || e.Players[0].Stack != 950 || e.Players[1].Bet != 50 || e.Players[1].Stack != 975 {
				t.Error("incorrect second-hand blind deductions")
			}
		}
		if e.Type == HoleCardsDealt {
			dealtHands = append(dealtHands, e.HandNumber)
			seen := make(map[[2]int]bool)
			for _, p := range e.Players {
				for _, c := range []pokeralgo.Card{p.HoleCards.First, p.HoleCards.Second} {
					key := [2]int{c.Rank, int(c.Suit)}
					if c.Rank < 2 || c.Rank > 14 || seen[key] {
						t.Errorf("hand %d: invalid or duplicate dealt card %v", e.HandNumber, c)
					}
					seen[key] = true
				}
			}
		}
	}
	wantMilestones := []EventType{HandStarted, ChipsAwarded, HandEnded, PlayerLeft, HandStarted, ChipsAwarded, HandEnded, PlayerLeft, GameEnded}
	if !slices.Equal(milestones, wantMilestones) {
		t.Errorf("expected lifecycle %v, got %v", wantMilestones, milestones)
	}
	if !slices.Equal(dealtHands, []int{1, 2}) {
		t.Errorf("expected cards dealt for hands 1 and 2, got %v", dealtHands)
	}
}

func TestPlayHandPaysAllInWinnerBeforeRemovingLoser(t *testing.T) {
	var events []Event
	g := New(Options{BuyIn: 100, BigBlind: 50, AdditionalRaises: 1, EventVerbosity: EventsAll}, nil,
		func(e Event) { events = append(events, e) })
	g.deck = pokeralgo.NewDeckWithSeed(1)
	source := &streetActionSource{t: t, game: g, steps: []streetStep{
		{"A", 25, []ActionType{Fold, Call, Raise, Leave}, Action{Type: Raise, Amount: 75}, 0},
		{"B", 50, []ActionType{Fold, Call, Leave}, Action{Type: Call}, 0},
	}}
	g.actionSource = source
	if err := g.seatPlayers([]PlayerSpec{NewPlayerSpec("A"), NewPlayerSpec("B")}); err != nil {
		t.Fatal(err)
	}
	if err := g.playHand(); err != nil {
		t.Fatal(err)
	}
	if source.next != len(source.steps) {
		t.Fatalf("consumed %d of %d actions", source.next, len(source.steps))
	}
	var lifecycle []EventType
	var winnerID, bustedID string
	for _, e := range events {
		switch e.Type {
		case ChipsAwarded:
			lifecycle = append(lifecycle, e.Type)
			if len(e.Players) != 2 || e.Pots == nil || len(*e.Pots) != 1 || len((*e.Pots)[0].WinnerIDs) != 1 {
				t.Fatalf("expected a sole winner with both players still present: %v", e)
			}
			winnerID = (*e.Pots)[0].WinnerIDs[0]
			for _, p := range e.Players {
				wantStack := 0
				if p.ID == winnerID {
					wantStack = 200
				}
				if p.Stack != wantStack {
					t.Errorf("payout: expected %s stack %d, got %d", p.ID, wantStack, p.Stack)
				}
			}
		case HandEnded, GameEnded:
			lifecycle = append(lifecycle, e.Type)
		case PlayerBusted:
			lifecycle = append(lifecycle, e.Type)
			bustedID = *e.PlayerID
		}
	}
	assertOrder := []EventType{ChipsAwarded, HandEnded, PlayerBusted, GameEnded}
	if !slices.Equal(lifecycle, assertOrder) {
		t.Errorf("expected payout before removal: %v, got %v", assertOrder, lifecycle)
	}
	if !g.gameOver || len(g.players) != 1 || g.players[0].ID != winnerID || g.players[0].Stack != 200 || bustedID == winnerID {
		t.Errorf("incorrect survivors after all-in payout: %v", g.players)
	}
}
