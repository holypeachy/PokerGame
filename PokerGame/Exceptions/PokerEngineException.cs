namespace PokerGame;

[Serializable]
public class PokerEngineException : Exception
{
    public PokerEngineException() { }
    public PokerEngineException(string message) : base(message) { }
    public PokerEngineException(string message, Exception inner) : base(message, inner) { }
}