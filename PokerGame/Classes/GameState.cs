namespace PokerGame;

public record GameState
{
    public required List<PlayerState> PlayerStates;
    public required List<Card> CommunityCards;
    public required OutputType OutputType;
    
    public PlayerState? PlayerToAct;
    public List<PlayerMove> PossibleMoves;
    public required int ToCall;
}