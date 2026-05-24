namespace PokerGame.Sandbox;

class Program
{
    static void Main()
    {
        var engineOptions = new PokerEngineOptions { BuyIn = 1000, BigBlind = 50, AdditionalRaises = 1, EnableDebug = true};
        EngineIO io = new();
        PokerGame engine = new(engineOptions, io);
        io.SetEngine(engine);

        List<PlayerInfoDto> playersInfo =
        [
            new PlayerInfoDto("Alpha"),
            new PlayerInfoDto("Tango"),
            new PlayerInfoDto("Sierra"),
            new PlayerInfoDto("Quebec"),
            new PlayerInfoDto("Zulu"),
        ];

        engine.InitializeTable(playersInfo);
        engine.StartHand();
        engine.StartHand();
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
