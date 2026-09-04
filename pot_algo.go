package pokergame

import "fmt"

var EnablePotDebugLog = true

type chipTracker struct {
	Owner     *GamePlayer
	Value     int
	HasFolded bool
}

func (t *chipTracker) String() string {
	return fmt.Sprintf("Owner: %s | Value: %d | Folded: %t", t.Owner.Name, t.Value, t.HasFolded)
}

func GetPots(players []*GamePlayer) ([]*Pot, error) {
	atLeastOneNonFolded := false
	for _, item := range players {
		if !item.HasFolded {
			atLeastOneNonFolded = true
			break
		}
	}
	if atLeastOneNonFolded == false {
		return nil, newInternalPokerGameError("GetPots() called when all players have folded.")
	}

	trackers := []*chipTracker{}
	for _, p := range players {
		if p.Bet != 0 {
			if p.Bet < 0 {
				return nil, newInternalPokerGameError("Player %s has a negative bet value.", p.Name)
			}
			trackers = append(trackers, &chipTracker{Owner: p, Value: p.Bet, HasFolded: p.HasFolded})
		}
	}

	if len(trackers) == 0 {
		return nil, newInternalPokerGameError("GetPots() called when no players have bet anything.")
	}

	if EnablePotDebugLog {
		fmt.Println("Initial Chip Trackers:")
		for _, t := range trackers {
			fmt.Println(t)
		}
		fmt.Println()
	}

	return splitPot(trackers)
}

func splitPot(trackers []*chipTracker) ([]*Pot, error) {
	// end condition
	if len(trackers) == 0 {
		if EnablePotDebugLog {
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
	potPlayers := []*GamePlayer{}

	// loop through trackers and remove value
	for _, t := range trackers {
		if t.HasFolded {
			if t.Value <= min {
				foldedTotal += t.Value
				t.Value = 0
			} else {
				foldedTotal += min
				t.Value -= min
			}
		} else {
			potPlayers = append(potPlayers, t.Owner)
			potTotal += min
			t.Value -= min
		}
	}

	pot := NewPot(potTotal+foldedTotal, potPlayers)
	if EnablePotDebugLog {
		fmt.Println("Pot in Recursion:")
		fmt.Println(pot)
		fmt.Printf("Current Number of Trackers: %d\n", len(trackers))
		fmt.Println()
	}

	// prepare trackers for next recursion
	nextTrackers := trackers[:0]
	for _, t := range trackers {
		if t.Value != 0 {
			nextTrackers = append(nextTrackers, t)
		}
	}

	// we combine all the pots
	pots := []*Pot{pot}
	nextPots, err := splitPot(nextTrackers)
	if err != nil {
		return nil, err
	}
	pots = append(pots, nextPots...)
	return pots, nil
}

func getMinBet(trackers []*chipTracker) (int, error) {
	min := int(^uint(0) >> 1)
	found := false

	for _, t := range trackers {
		if t.HasFolded {
			continue
		}

		found = true
		if t.Value < min {
			min = t.Value
		}
	}

	if !found {
		return 0, newInternalPokerGameError("GetMin() called with no non-folded trackers remaining. There should be at least one non-folded player, with the highest bet.")
	}

	return min, nil
}
