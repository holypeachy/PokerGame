package main

import (
	"fmt"
	"os"
	"pokeralgo"

	pokergame "pokergame"
)

func main() {
	engineOptions := pokergame.Options{BuyIn: 1000, BigBlind: 50, AdditionalRaises: 1, EventVerbosity: pokergame.EventsCore}
	actionSource := NewBasicActionSource()
	engine := pokergame.New(engineOptions, actionSource, func(e pokergame.Event) {
		fmt.Println(e)
	})

	playersInfo := []pokergame.PlayerSpec{
		pokergame.NewPlayerSpec("Alpha"),
		pokergame.NewPlayerSpec("Tango"),
		pokergame.NewPlayerSpec("Sierra"),
		pokergame.NewPlayerSpec("Quebec"),
		pokergame.NewPlayerSpec("Zulu"),
	}

	pokeralgo.SetDebugLevel(pokeralgo.DebugSummary)
	if err := engine.Play(playersInfo); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("--- Main Execution Ends ---")
}

/*
! ISSUES:
! If a call is all in, then raise should be removed
! when 2 players are left, dealer is small blind
! stale hole card data between hands
!

TODO
TODO: Create a standard logger (logs to stdout + files). Turns events into detailed log entries.
TODO: Add end hand and end game logic and reporting
TODO: Add Marked for removal logic
TODO: Create a standard replay sytem
TODO: Unit and integration tests
TODO: Add limit hold'em rules

? Future Ideas
?

* Notes
* Logging format
	<emoji> [component] section or result
		details

* Changes
* Refactored event emissions
* Added EventVerbosity and implemented a guard to prevent event construction
* Added createPotEvent
* Added chipsAwardedOnePlayerLeft
*/
