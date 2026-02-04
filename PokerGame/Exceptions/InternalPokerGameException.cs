namespace PokerGame;

[Serializable]
public class InternalPokerGameException : PokerGameException
{
    public InternalPokerGameException() { }

    public InternalPokerGameException(string message) : base(message) { }

    public InternalPokerGameException(string message, Exception inner) : base(message, inner) { }
}