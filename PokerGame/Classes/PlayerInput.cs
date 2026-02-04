namespace PokerGame;

public record PlayerInput
{
    public required PlayerMove Move;
    public int Amount;
}