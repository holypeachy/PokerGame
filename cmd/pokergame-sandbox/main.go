package main

import (
	"fmt"
	"os"

	pokergame "pokergame"
)

func main() {
	engineOptions := pokergame.PokerEngineOptions{BuyIn: 1000, BigBlind: 50, AdditionalRaises: 1, EnableDebug: true}
	actionSource := NewConsoleActionSource()
	engine := pokergame.NewPokerGame(engineOptions, actionSource)
	actionSource.SetEngine(engine)

	playersInfo := []pokergame.PlayerInfoDto{
		pokergame.NewPlayerInfoDto("Alpha"),
		pokergame.NewPlayerInfoDto("Tango"),
		pokergame.NewPlayerInfoDto("Sierra"),
		pokergame.NewPlayerInfoDto("Quebec"),
		pokergame.NewPlayerInfoDto("Zulu"),
	}

	if err := engine.InitializeTable(playersInfo); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := engine.StartHand(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := engine.StartHand(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

/*
! ISSUES:
!

TODO
TODO: Add detailed and standardized logging, log to file as well. Deep engine logging.
TODO: Add end hand and end game logic and reporting
TODO: Implement replay functionality
TODO: Unit and integration tests
TODO: Account for flexible ruleset so I can make changes later

? Future Ideas
?

* Notes
*

* Changes
* switch to .net 10
* start working on logging system
*
*/
