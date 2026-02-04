namespace PokerGame;

[Serializable]
public class PokerGameException : Exception
{
    public PokerGameException() { }
    public PokerGameException(string message) : base(message) { }
    public PokerGameException(string message, Exception inner) : base(message, inner) { }
}