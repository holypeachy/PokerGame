namespace PokerGame;

public record PlayerState
{
    public required string Id;
    public required int Stack;
    public required bool HasFolded;
    public Pair? HoleCards;
    public required int Bet;
}