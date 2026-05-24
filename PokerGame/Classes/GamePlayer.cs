namespace PokerGame;

public class GamePlayer
{
    public string Name { get; }
    public int Stack { get; private set; }

    public Pair HoleCards { get; private set; }
    public int Bet { get; private set; } = 0;
    public bool HasActed { get; private set; } = false;
    public bool HasFolded { get; private set; } = false;

    public GamePlayer(string name, int stack, Card first, Card second)
    {
        if (first.Equals(second)) throw new InternalPokerGameException("Hole cards should never be the same card.");

        Name = name;
        Stack = stack;
        HoleCards = new Pair(first, second);
    }

    public GamePlayer(PlayerInfoDto playerInfo, int stack, Card first, Card second)
    {
        if (first.Equals(second)) throw new InternalPokerGameException("Hole cards should never be the same card.");

        Name = playerInfo.Id;
        Stack = stack;
        HoleCards = new Pair(first, second);
    }

    public void ResetHand()
    {
        Bet = 0;
        HasActed = false;
        HasFolded = false;
    }

    public void ResetBettingRound()
    {
        HasActed = false;
    }

    public void Pay(int amount)
    {
        if (amount < 0) throw new InternalPokerGameException("Pay amount cannot be negative.");

        Stack += amount;
    }

    public void Fold()
    {
        if (HasFolded) throw new InternalPokerGameException("Player has already folded.");
        HasFolded = true;
    }

    public void Check()
    {
        HasActed = true;
    }

    public void MakeBet(int amount)
    {
        if (amount < 0) throw new InternalPokerGameException("Bet amount cannot be negative.");

        MakeBlindBet(amount);
        HasActed = true;
    }

    public void MakeBlindBet(int amount)
    {
        if (amount < 0) throw new InternalPokerGameException("Bet amount cannot be negative.");
        
        if (amount > Stack)
        {
            Bet += Stack;
            Stack = 0;
            return;
        }
        Bet += amount;
        Stack -= amount;
    }

    public void NewHand(Card first, Card second)
    {
        if (first.Equals(second)) throw new InternalPokerGameException("Hole cards should never be the same card.");
        HoleCards = new(first, second);
    }

    public override string ToString()
    {
        return $"{Name}\nCurrentBet: {Bet} | Hand: {HoleCards} | Stack: {Stack}";
    }
}
