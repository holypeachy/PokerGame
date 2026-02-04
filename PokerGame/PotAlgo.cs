namespace PokerGame;

public static class PotAlgo
{
    public static bool EnableDebugLog { get; set; } = true;

    class ChipTracker(GamePlayer owner, int value, bool HasFolded)
    {
        public GamePlayer Owner { get; } = owner;
        public int Value { get; set; } = value;
        public bool HasFolded { get; } = HasFolded;

        public override string ToString()
        {
            return $"Owner: {Owner.Name} | Value: {Value} | Folded: {HasFolded}";
        }
    }

    public static List<Pot> GetPots(List<GamePlayer> players)
    {
        bool atLeastOneNonFolded = false;
        foreach (var item in players)
        {
            if (!item.HasFolded)
            {
                atLeastOneNonFolded = true;
                break;
            }
        }
        if (atLeastOneNonFolded == false) throw new InternalPokerEngineException("GetPots() called when all players have folded.");

        List<ChipTracker> trackers = [];
        foreach (GamePlayer p in players)
        {
            if (p.Bet != 0)
            {
                if(p.Bet < 0) throw new InternalPokerEngineException($"Player {p.Name} has a negative bet value.");
                trackers.Add(new ChipTracker(p, p.Bet, p.HasFolded));
            }
        }

        if (trackers.Count == 0) throw new InternalPokerEngineException("GetPots() called when no players have bet anything.");

        if (EnableDebugLog)
        {
            Console.WriteLine("Trackers:");
            foreach (ChipTracker t in trackers)
            {
                Console.WriteLine(t);
            }
            Console.WriteLine();
        }

        return SplitPot(trackers);
    }

    private static List<Pot> SplitPot(List<ChipTracker> trackers)
    {
        // end condition
        if (trackers.Count == 0) {
            if (EnableDebugLog) Console.WriteLine("End of Recursion.");
            return [];
        }

        // pot splitting logic
        Pot pot;
        int min = GetMin(trackers);
        int potTotal = 0;
        int foldedTotal = 0;
        List<GamePlayer> potPlayers = [];

        // loop through trackers and decrease value
        foreach (ChipTracker t in trackers)
        {
            if (t.HasFolded)
            {
                if (t.Value <= min)
                {
                    foldedTotal += t.Value;
                    t.Value = 0;
                }
                else
                {
                    foldedTotal += min;
                    t.Value -= min;
                }
            }
            else
            {
                potPlayers.Add(t.Owner);
                potTotal += min;
                t.Value -= min;
            }
        }

        pot = new Pot(potTotal + foldedTotal, potPlayers);
        if (EnableDebugLog)
        {
            Console.WriteLine($"Current Number of Trackers: {trackers.Count}");
            Console.WriteLine("Pot in Recursion:");
            Console.WriteLine(pot);
            Console.WriteLine();
        }

        // prepare trackers for next recursion
        trackers.RemoveAll(t => t.Value == 0);

        // we combine all the pots
        List<Pot> pots = [pot];
        pots.AddRange(SplitPot(trackers));
        return pots;
    }

    private static int GetMin(List<ChipTracker> trackers)
    {
        int min = int.MaxValue;
        bool found = false;

        foreach (var t in trackers)
        {
            if (t.HasFolded) continue;

            found = true;
            if (t.Value < min) min = t.Value;
        }

        if (!found) throw new InternalPokerEngineException("GetMin() called with no non-folded trackers remaining.");

        return min;
    }

}