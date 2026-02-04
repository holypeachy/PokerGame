using System.Diagnostics;

namespace PokerGame;

public class Pot(int value, List<GamePlayer> players)
{
    public List<GamePlayer> Players { get; private set; } = players;
    public int Value { get; private set; } = value;
    
    public List<GamePlayer>? Winners { get; set; } = null;

    public void PayWinners()
    {
        if (Winners is null || Winners.Count == 0) throw new InvalidOperationException("Winners should never be null. This means we never determined the winners of this pot.");

        int split = Value / Winners.Count;
        foreach (var w in Winners)
        {
            w.Pay(split);
        }
    }

    public override string ToString()
    {
        string players = "| ";
        foreach (var p in Players)
        {
            players += p.Name + " | ";
        }

        string wString = string.Empty;
        if (Winners is not null)
        {
            foreach (GamePlayer w in Winners)
            {
                wString += $"\t{w.Name} ({Value / Winners.Count()}) | {w.Stack} => {w.Stack + Value / Winners.Count()}\n";
            }
        }

        return $"Players ({Players.Count}): \n{players}\nValue: {Value}\nWinner(s):\n{wString}";
    }
}