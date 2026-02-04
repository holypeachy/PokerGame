namespace PokerGame.Sandbox;

public class EngineIO : IEngineIO
{
    private PokerGame _engine;

    public PlayerInput GetInput(GameState gameState)
    {
        Dictionary<int, PlayerMove> moves = new();
        _engine.PrintGameState();
        Console.WriteLine("IO Request");
        Console.WriteLine($"Output Type: {gameState.OutputType}");
        Console.WriteLine("Current Player: " + (gameState.PlayerToAct is not null ? gameState.PlayerToAct.Id : "null"));
        Console.WriteLine($"{nameof(_engine.AdditionalRaiseCount)}: {_engine.AdditionalRaiseCount}");
        Console.Write("Possible Moves: ");
        int count = 1;
        if (gameState.PossibleMoves is not null)
        {
            foreach (var item in gameState.PossibleMoves)
            {
                Console.Write($"{count}-" + item + " ");
                moves.Add(count++, item);
            }
            Console.WriteLine();
        }
        else
        {
            Console.WriteLine("null");
        }
        Console.WriteLine($"ToCall: {gameState.ToCall}");
        Console.WriteLine("Select your move:");
        string? moveIn = Console.ReadLine();
        if (string.IsNullOrEmpty(moveIn)) throw new Exception();
        PlayerMove selectedMove = moves[int.Parse(moveIn)];

        int amount = gameState.ToCall;
        if(selectedMove is PlayerMove.Raise)
        {
            Console.WriteLine("ToCall + What Amount:");
            if (!int.TryParse(Console.ReadLine(), out amount)) throw new Exception();
        }
        
        return new PlayerInput { Move = selectedMove, Amount = gameState.ToCall + amount };
    }

    public void SetEngine(PokerGame engine)
    {
        _engine = engine;
    }
}