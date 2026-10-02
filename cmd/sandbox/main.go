package main

import (
	"fmt"
	"os"

	pokergame "pokergame"
)

func main() {
	engineOptions := pokergame.Options{BuyIn: 1000, BigBlind: 50, AdditionalRaises: 1, EnableDebug: true}
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
TODO: Refactor event emission
TODO: 2. gate event construction based on needs to prevent unnecessary heap allocs. aka: levels of verbosity for logging (full), replay system and gui (core), ai training (off)
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
* Pulled out event construction to their own functions
* Errors now emit error events at public interfaces rather than all over the place
* Engine core logic loop done yay!!
*
*/
