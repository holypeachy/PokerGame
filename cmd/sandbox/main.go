package main

import (
	"fmt"
	"os"

	pokergame "pokergame"
)

func main() {
	engineOptions := pokergame.Options{BuyIn: 1000, BigBlind: 50, AdditionalRaises: 1, EventVerbosity: pokergame.EventsAll}
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

	// pokeralgo.SetDebugLevel(pokeralgo.DebugSummary)
	if err := engine.Play(playersInfo); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("--- Main Execution Ends ---")
}

/*
! ISSUES:
!

TODO
TODO: Create a standard logger (logs to stdout + files). Turns events into detailed log entries.
TODO: Unit testing
TODO: Create a standard replay sytem
TODO: Test std replay
TODO: Integration testing
TODO: Add limit hold'em rules

? Future Ideas
? Add increading blinds up to a blind limit
?

* Notes
* Logging format
	<emoji> [component] section or result
		details

* Changes
* When removing players, dealer index is adjusted
* If a call is all in, the raise option will be removed
* when 2 players are left, dealer is small blind, and acting is dealer
* removed debug.go
* stale hole card data between hands
*
*/
