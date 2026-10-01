package pokergame

import "fmt"

var potDebugEnabled = true

type contribution struct {
	Player    *Player
	Remaining int
	HasFolded bool
}

func (t *contribution) String() string {
	return fmt.Sprintf("Owner: %s | Value: %d | Folded: %t", t.Player.ID, t.Remaining, t.HasFolded)
}

func buildPots(players []*Player) ([]*Pot, error) {
	atLeastOneNonFolded := false
	for _, item := range players {
		if !item.Folded {
			atLeastOneNonFolded = true
			break
		}
	}
	if atLeastOneNonFolded == false {
		return nil, fmt.Errorf("%w: buildPots() called when all players have folded.", ErrInternal)
	}

	trackers := []*contribution{}
	for _, p := range players {
		if p.Bet != 0 {
			if p.Bet < 0 {
				return nil, fmt.Errorf("%w: Player %s has a negative bet value.", ErrInternal, p.ID)
			}
			trackers = append(trackers, &contribution{Player: p, Remaining: p.Bet, HasFolded: p.Folded})
		}
	}

	if len(trackers) == 0 {
		return nil, fmt.Errorf("%w: buildPots() called when no players have bet anything.", ErrInternal)
	}

	if potDebugEnabled {
		fmt.Println("Initial Chip Trackers:")
		for _, t := range trackers {
			fmt.Println(t)
		}
		fmt.Println()
	}

	return splitPots(trackers)
}

func splitPots(trackers []*contribution) ([]*Pot, error) {
	// end condition
	if len(trackers) == 0 {
		if potDebugEnabled {
			fmt.Println("End of Recursion.")
		}
		return []*Pot{}, nil
	}

	// pot splitting logic
	min, err := getMinBet(trackers)
	if err != nil {
		return nil, err
	}
	potTotal := 0
	foldedTotal := 0
	potPlayers := []*Player{}

	// loop through trackers and remove value
	for _, t := range trackers {
		if t.HasFolded {
			if t.Remaining <= min {
				foldedTotal += t.Remaining
				t.Remaining = 0
			} else {
				foldedTotal += min
				t.Remaining -= min
			}
		} else {
			potPlayers = append(potPlayers, t.Player)
			potTotal += min
			t.Remaining -= min
		}
	}

	pot := NewPot(potTotal+foldedTotal, potPlayers)
	if potDebugEnabled {
		fmt.Println("Pot in Recursion:")
		fmt.Println(pot)
		fmt.Printf("Current Number of Trackers: %d\n", len(trackers))
		fmt.Println()
	}

	// prepare trackers for next recursion
	nextTrackers := trackers[:0]
	for _, t := range trackers {
		if t.Remaining != 0 {
			nextTrackers = append(nextTrackers, t)
		}
	}

	// we combine all the pots
	pots := []*Pot{pot}
	nextPots, err := splitPots(nextTrackers)
	if err != nil {
		return nil, err
	}
	pots = append(pots, nextPots...)
	return pots, nil
}

func getMinBet(trackers []*contribution) (int, error) {
	min := int(^uint(0) >> 1)
	found := false

	for _, t := range trackers {
		if t.HasFolded {
			continue
		}

		found = true
		if t.Remaining < min {
			min = t.Remaining
		}
	}

	if !found {
		return 0, fmt.Errorf("%w: getMinBet() called with no non-folded trackers remaining. There should be at least one non-folded player, with the highest bet.", ErrInternal)
	}

	return min, nil
}
