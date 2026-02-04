namespace PokerGame;

[Serializable]
public class InternalPokerEngineException : PokerEngineException
{
    public InternalPokerEngineException() { }

    public InternalPokerEngineException(string message) : base(message) { }

    public InternalPokerEngineException(string message, Exception inner) : base(message, inner) { }
}