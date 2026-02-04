using System.Diagnostics;

namespace PokerGame;

public class PokerGame
{
    // Engine
    public readonly PokerEngineOptions EngineOptions;
    private readonly IEngineIO _io;

    // Table
    private readonly Deck _deck;
    private readonly List<GamePlayer> _players;

    // Hand
    public readonly List<Card> CommunityCards = new(5);

    public int DealerIndex { get; private set; } = 0;
    public int SbIndex { get; private set; } = 0;
    public int BbIndex { get; private set; } = 0;

    public int CurrentPlayerIndex { get; private set; } = 0;

    public int CurrentBet { get; set; } = 0;
    public int AdditionalRaiseCount { get; set; } = 0;

    public bool IsOnePlayerLeft { get; set; } = false;
    public bool IsSkipToShowdown { get; set; } = false;

    public PokerGame(PokerEngineOptions options, IEngineIO io)
    {
        EngineOptions = options;
        _io = io;
        _players = [];
        _deck = new();
    }

    public void InitializeTable(List<PlayerInfo> playersInfo)
    {
        if (playersInfo.Count < 2) throw new Exception("Not enough players");

        _players.Clear();
        _deck.ResetDeck();
        DealerIndex = -1;

        foreach (var pi in playersInfo)
        {
            _players.Add(new GamePlayer(pi, EngineOptions.BuyIn, _deck.NextCard(), _deck.NextCard()));
        }
    }

    public void StartHand()
    {
        _deck.ResetDeck();
        CommunityCards.Clear();
        AdvanceBlinds();

        // Pre-flop
        Console.WriteLine("Pre-Flop");
        _players[SbIndex].MakeBlindBet(EngineOptions.BigBlind / 2);
        _players[BbIndex].MakeBlindBet(EngineOptions.BigBlind);
        CurrentBet = EngineOptions.BigBlind;

        foreach (var p in _players)
        {
            p.NewHand(_deck.NextCard(), _deck.NextCard());
        }

        ExecuteBetting();
        CurrentPlayerIndex = GetNextIndex(DealerIndex);

        // Flop
        Console.WriteLine("Flop");
        _deck.NextCard();
        if(!IsOnePlayerLeft) CommunityCards.AddRange(_deck.NextCards(3));
        ExecuteBetting();
        CurrentPlayerIndex = GetNextIndex(DealerIndex);

        // Turn
        Console.WriteLine("Turn");
        _deck.NextCard();
        if (!IsOnePlayerLeft) CommunityCards.Add(_deck.NextCard());
        ExecuteBetting();
        CurrentPlayerIndex = GetNextIndex(DealerIndex);

        // River
        Console.WriteLine("River");
        _deck.NextCard();
        if (!IsOnePlayerLeft) CommunityCards.Add(_deck.NextCard());
        ExecuteBetting();

        if (!IsOnePlayerLeft)
        {
            Showdown();
        }

        ResetHand();
    }

    private void ExecuteBetting()
    {
        if (IsOnePlayerLeft || IsSkipToShowdown) {
            if(IsSkipToShowdown) Console.WriteLine("Skip to showdown");
            return;
        }

        bool isBettingRoundOver = false;
        while (!isBettingRoundOver)
        {
            while (true)
            {
                GamePlayer currentPlayer = _players[CurrentPlayerIndex];

                // Before
                if (currentPlayer.HasFolded || IsPlayerAllIn(currentPlayer))
                {
                    CurrentPlayerIndex = GetNextIndex(CurrentPlayerIndex);
                    continue;
                }
                else if (CountNonFoldedPlayers() == 1)
                {
                    IsOnePlayerLeft = true;
                    isBettingRoundOver = true;
                    break;
                }
                else if (CountPlayersThatCanAct() < 2 && IsEveryoneSettledAtCurrentBet())
                {
                    IsSkipToShowdown = true;
                    isBettingRoundOver = true;
                    break;
                }
                else if (HasEveryoneActed() && IsEveryoneSettledAtCurrentBet())
                {
                    isBettingRoundOver = true;
                    break;
                }

                // Action
                var gameState = GetGameState();
                var validMoves = gameState.PossibleMoves ?? throw new Exception();
                PlayerInput input = _io.GetInput(gameState);
                if (!validMoves.Contains(input.Move)) throw new Exception("Input not valid");

                if (input.Move is PlayerMove.Fold)
                {
                    currentPlayer.Fold();
                    if (CountNonFoldedPlayers() == 1)
                    {
                        IsOnePlayerLeft = true;
                        isBettingRoundOver = true;
                        break;
                    }
                }
                else if (input.Move is PlayerMove.Raise || input.Move is PlayerMove.Call)
                {
                    if (HasEveryoneActed() && input.Move is PlayerMove.Raise) AdditionalRaiseCount++;

                    if (input.Move is PlayerMove.Raise)
                    {
                        int toCall = CurrentBet - currentPlayer.Bet;
                        if (input.Amount < toCall) currentPlayer.MakeBet(toCall + 10);
                        else currentPlayer.MakeBet(input.Amount);
                    }
                    else currentPlayer.MakeBet(CurrentBet - currentPlayer.Bet);

                    if (currentPlayer.Bet > CurrentBet) CurrentBet = currentPlayer.Bet;
                }
                else if (input.Move is PlayerMove.Check) currentPlayer.Check();

                // After
                if (CountPlayersThatCanAct() < 2 && IsEveryoneSettledAtCurrentBet())
                {
                    isBettingRoundOver = true;
                    IsSkipToShowdown = true;
                    break;
                }

                // Move To Next Player
                CurrentPlayerIndex = GetNextIndex(CurrentPlayerIndex);
            }
        }

        if (IsOnePlayerLeft)
        {
            Console.WriteLine("1 non-folded player remains");
            // 1 non-folded player remains
            GamePlayer winner = GetNonFoldedPlayer();
            int pot = 0;
            foreach (var p in _players)
            {
                pot += p.Bet;
            }
            winner.Pay(pot);

            if (CountNonFoldedPlayers() != 1)
            {
                throw new Exception("1 non-folded player remains but count is not 1");
            }
        }
        ResetBettingRound();
    }

    private void Showdown()
    {
        if (IsOnePlayerLeft)
        {
            Console.WriteLine("No Showdown. Only one player left");
            return;
        }

        List<Pot> pots = PotAlgo.GetPots(_players);
        List<Player> algoPlayers = EnginePlayersToAlgoPlayers(_players);
        Console.WriteLine("All Algo Players");
        foreach (var item in algoPlayers)
        {
            Console.WriteLine(item);
        }
        Console.WriteLine();
        foreach (var pot in pots)
        {
            if (pot.Players.Count == 1)
            {
                pot.Winners = pot.Players;
            }
            else if (pot.Players.Count < 1) throw new Exception();
            else
            {
                if (CommunityCards.Count != 5) throw new Exception();

                List<Player> winners = Algo.GetWinners(EnginePlayersToAlgoPlayers(pot.Players), CommunityCards);
                Console.WriteLine("Algo Winners for THIS pot");
                foreach (var item in winners)
                {
                    Console.WriteLine(item);
                    Console.WriteLine(item.WinningHand);
                }
                Console.WriteLine();
                pot.Winners = MapAlgoPlayersToEnginePlayers(winners);
            }
        }
        Console.WriteLine("\n--- PAY ---\n");
        foreach (var item in pots)
        {
            Console.WriteLine("\nPot:");
            Console.WriteLine(item);
            item.PayWinners();
        }
    }


    private List<GamePlayer> MapAlgoPlayersToEnginePlayers(List<Player> players)
    {
        List<GamePlayer> enginePlayers = new(_players);
        enginePlayers = enginePlayers.Where(p => players.Find(p2 => p2.Name == p.Name) is not null).ToList();
        Debug.Assert(enginePlayers.Count > 0, "No PokerAlgo players match any engine players");
        return enginePlayers;
    }

    private static List<Player> EnginePlayersToAlgoPlayers(List<GamePlayer> enginePlayers)
    {
        List<Player> players = new();
        foreach (var ep in enginePlayers)
        {
            players.Add(new Player(ep.Name, ep.HoleCards.First, ep.HoleCards.Second));
        }
        return players;
    }


    private List<PlayerMove> GetPossibleMoves(GamePlayer player)
    {
        Debug.Assert(!player.HasFolded, "Cannot get possible moves for folded player");

        List<PlayerMove> moves = new();
        int toCall = CurrentBet - player.Bet;
        if (toCall > 0)
        {
            moves.Add(PlayerMove.Call);
            moves.Add(PlayerMove.Fold);
        }
        else if (toCall == 0)
        {
            moves.Add(PlayerMove.Check);
        }
        // we count players that can act because we don't want to raise if another player is all-in
        if (AdditionalRaiseCount != EngineOptions.AdditionalRaises && CountPlayersThatCanAct() > 1) moves.Add(PlayerMove.Raise);
        return moves;
    }

    private GamePlayer GetNonFoldedPlayer()
    {
        foreach (var p in _players)
        {
            if (!p.HasFolded) return p;
        }
        throw new Exception("No non-folded player");
    }


    private bool IsEveryoneSettledAtCurrentBet()
    {
        foreach (var p in _players)
        {
            if (!p.HasFolded && !IsPlayerAllIn(p) && CurrentBet != p.Bet) return false;
        }
        return true;
    }

    private bool HasEveryoneActed()
    {
        foreach (var p in _players)
        {
            if (!p.HasFolded && !p.HasActed && !IsPlayerAllIn(p)) return false;
        }
        return true;
    }


    private int CountPlayersThatCanAct()
    {
        int count = 0;
        foreach (var p in _players)
        {
            if (!p.HasFolded && !IsPlayerAllIn(p)) count++;
        }
        return count;
    }

    private int CountNonFoldedPlayers()
    {
        int count = 0;
        foreach (var p in _players)
        {
            if (!p.HasFolded) count++;
        }
        return count;
    }

    private static bool IsPlayerAllIn(GamePlayer player)
    {
        if (player.Stack == 0 && player.Bet == 0) throw new InternalPokerEngineException("Player has 0 stack and 0 bet");
        
        if (player.Stack == 0 && player.Bet > 0) return true;
        return false;
    }


    private void ResetBettingRound()
    {
        foreach (var p in _players)
        {
            p.ResetBettingRound();
        }
        AdditionalRaiseCount = 0;
    }

    private void ResetHand()
    {
        foreach (var p in _players)
        {
            p.ResetHand();
        }
        AdditionalRaiseCount = 0;
        IsOnePlayerLeft = false;
        IsSkipToShowdown = false;
    }


    private GameState GetGameState()
    {
        List<PlayerState> playerStates = new();
        foreach (var player in _players)
        {
            playerStates.Add(new PlayerState { Id = player.Name, Stack = player.Stack, HoleCards = player.HoleCards, Bet = player.Bet, HasFolded = player.HasFolded});
        }
        PlayerState currentPlayer = playerStates[CurrentPlayerIndex];
        Debug.Assert(!currentPlayer.HasFolded, "Current player cannot be folded");
        int toCall = CurrentBet - currentPlayer.Bet;
        return new GameState
        {
            PlayerStates = playerStates,
            CommunityCards = new List<Card>(CommunityCards),
            OutputType = OutputType.InputRequest,
            PlayerToAct = currentPlayer,
            PossibleMoves = GetPossibleMoves(_players[CurrentPlayerIndex]),
            ToCall = toCall
        };
    }


    private void AdvanceBlinds()
    {
        DealerIndex = GetNextIndex(DealerIndex);
        SbIndex = GetNextIndex(DealerIndex);
        BbIndex = GetNextIndex(SbIndex);
        CurrentPlayerIndex = GetNextIndex(BbIndex);
    }

    private int GetNextIndex(int index)
    {
        int temp = index + 1;
        if (temp > _players.Count - 1) temp = 0;
        return temp;
    }


    public void PrintGameState()
    {
        Console.WriteLine("-- Game State --");
        Console.WriteLine("Players: ");
        foreach (var p in _players)
        {
            Console.Write(_players.IndexOf(p) == CurrentPlayerIndex ? " ➡️  " : "");
            Console.Write(p);
            Console.Write(_players.IndexOf(p) == DealerIndex ? "   🎃 DEALER" : "");
            Console.Write(_players.IndexOf(p) == SbIndex ? "   🥈 SMALL BLIND" :  "");
            Console.Write(_players.IndexOf(p) == BbIndex ? "   🪙  BIG BLIND" : "");
            Console.Write(p.HasFolded ? "   ⛔" : "");
            Console.WriteLine(!p.HasFolded && p.HasActed ? "  ✅" : "");
        }
        Console.WriteLine("\nCommunity Cards: ");
        foreach (var item in CommunityCards)
        {
            Console.Write(item + " ");
        }
        Console.WriteLine();
    }
}
