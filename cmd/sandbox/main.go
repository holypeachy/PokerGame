package main

import (
	"fmt"
	"io"
	"os"
	"pokergame/internal/eventlog"

	pokergame "pokergame"
)

func main() {
	file, err := os.Create("game.log")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer file.Close()

	logger := eventlog.New(io.MultiWriter(os.Stdout, file))

	engineOptions := pokergame.Options{BuyIn: 1000, BigBlind: 50, AdditionalRaises: 1, EventVerbosity: pokergame.EventsCore}
	actionSource := NewBasicActionSource()
	engine := pokergame.New(engineOptions, actionSource, logger.OnEvent)

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
! Add odd chip handling

TODO
TODO: Verify test cases
TODO: Create a standard replay sytem
TODO: Test std replay
TODO: Integration testing with replay system

? Future Ideas
? Add limit hold'em rules
? Add increading blinds up to a blind limit
?

* Notes
* Logging format
	<emoji> [component] section or result
		details

* Changes
* Removed console_action_source
* NextAction no longer returns an error
* Added pot_algo, street, and legal actions unit tests
* Added between hand transition unit tests
* Added player unit tests
* Added showdown and events unit tests
*/
