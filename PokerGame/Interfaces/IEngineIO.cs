namespace PokerGame;

public interface IEngineIO
{
    PlayerInput GetInput(GameState gameState);
}