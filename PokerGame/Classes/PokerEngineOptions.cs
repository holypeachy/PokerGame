namespace PokerGame;

public record PokerEngineOptions
{
    public int BuyIn;
    public int BigBlind;
    public int AdditionalRaises;
    public bool EnableDebug;
    public DebugVerbosity DebugVerbosity;
}