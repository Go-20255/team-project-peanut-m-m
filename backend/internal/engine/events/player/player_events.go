package player

import (
    "context"
    "fmt"
    "math/rand/v2"
    "monopoly-backend/internal"
    internaldb_event_cards "monopoly-backend/internal/db/event_cards"
    internaldb_game_state "monopoly-backend/internal/db/game_state"
    internaldb_players "monopoly-backend/internal/db/player"
    internaldb_tiles "monopoly-backend/internal/db/tile"
    "monopoly-backend/internal/engine/events"
    turn_events "monopoly-backend/internal/engine/events/turn"
    "net/http"
    "sort"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/rs/zerolog"
)

func getDeferredLandingKey(playerId int) string {
    return fmt.Sprintf("deferred_landing_%d", playerId)
}

func setDeferredLanding(
    e *internal.MonopolyEngine,
    playerId int,
    sessionId string,
    diceTotal int,
    playerMovement internal.PlayerMovement,
) {
    e.TempStore[getDeferredLandingKey(playerId)] = internal.DeferredLanding{
        PlayerId:       playerId,
        SessionId:      sessionId,
        DiceTotal:      diceTotal,
        PlayerMovement: playerMovement,
    }
}

func clearDeferredLanding(e *internal.MonopolyEngine, playerId int) {
    delete(e.TempStore, getDeferredLandingKey(playerId))
}

func getSkipMoveEventKey(playerId int) string {
    return fmt.Sprintf("skip_move_event_%d", playerId)
}

func setSkipMoveEvent(e *internal.MonopolyEngine, playerId int) {
    e.TempStore[getSkipMoveEventKey(playerId)] = true
}

func consumeSkipMoveEvent(e *internal.MonopolyEngine, playerId int) bool {
    key := getSkipMoveEventKey(playerId)
    skip, ok := e.TempStore[key].(bool)
    if ok && skip {
        delete(e.TempStore, key)
        return true
    }

    delete(e.TempStore, key)
    return false
}

func getDeferredLanding(e *internal.MonopolyEngine, playerId int) (*internal.DeferredLanding, bool) {
    data, ok := e.TempStore[getDeferredLandingKey(playerId)].(internal.DeferredLanding)
    if !ok {
        return nil, false
    }

    return &data, true
}

func shouldDeferLandingUntilAfterGoPayout(
    tileData internaldb_tiles.RentTileData,
    playerId int,
    playerMovement internal.PlayerMovement,
) bool {
    if playerMovement.FromCard || (!playerMovement.PassedGo && playerMovement.NewPosition != 0) {
        return false
    }

    if playerMovement.NewPosition == 4 || playerMovement.NewPosition == 20 || playerMovement.NewPosition == 38 {
        return true
    }

    if tileData.Name == "Community Chest" || tileData.Name == "Chance" {
        return true
    }

    if tileData.HasProperty && !tileData.Owned && tileData.PropertyId > 0 {
        return true
    }

    if tileData.HasProperty && tileData.Owned && tileData.OwnerId != playerId && !tileData.IsMortgaged {
        return true
    }

    return false
}

func SetPendingBankPayment(
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    playerId int,
    sessionId string,
    amount int,
    reason string,
) internal.UserActionStatus {
    if e.PendingBankPayout != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there is already a pending bank payout",
        }
    }

    if e.PendingBankPayment != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there is already a pending bank payment",
        }
    }

    e.PendingBankPayment = &internal.PendingBankPayment{
        PlayerId:  playerId,
        SessionId: sessionId,
        Amount:    amount,
        Reason:    reason,
    }

    e.Broker.Broadcast(log, "BankPaymentDueEvent", e.PendingBankPayment)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   *e.PendingBankPayment,
    }
}

func SetPendingBankPayout(
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    playerId int,
    sessionId string,
    amount int,
    reason string,
) internal.UserActionStatus {
    if e.PendingBankPayment != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there is already a pending bank payment",
        }
    }

    if e.PendingBankPayout != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there is already a pending bank payout",
        }
    }

    e.PendingBankPayout = &internal.PendingBankPayout{
        PlayerId:  playerId,
        SessionId: sessionId,
        Amount:    amount,
        Reason:    reason,
    }

    e.Broker.Broadcast(log, "BankPayoutDueEvent", e.PendingBankPayout)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   *e.PendingBankPayout,
    }
}

func SetPendingPlayerExchange(
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    actingPlayerId int,
    sessionId string,
    amount int,
    reason string,
    isPayingAll bool,
) internal.UserActionStatus {
    if e.PendingExchange != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there is already a pending player exchange",
        }
    }

    e.PendingExchange = &internal.PendingPlayerExchange{
        ActingPlayerId: actingPlayerId,
        SessionId:      sessionId,
        Amount:         amount,
        Reason:         reason,
        IsPayingAll:    isPayingAll,
    }

    e.Broker.Broadcast(log, "PlayerExchangeDueEvent", e.PendingExchange)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   *e.PendingExchange,
    }
}

func SetPendingTrade(
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    pendingTrade internal.PendingTrade,
) internal.UserActionStatus {
    if e.PendingTrade != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there is already a pending trade",
        }
    }

    e.PendingTrade = &pendingTrade
    e.Broker.Broadcast(log, "TradeProposedEvent", e.PendingTrade)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   *e.PendingTrade,
    }
}

func SetPendingTradeDraft(
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    pendingTradeDraft internal.PendingTradeDraft,
) internal.UserActionStatus {
    if e.PendingTrade != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there is already a pending trade",
        }
    }

    if e.PendingTradeDraft != nil {
        ClearPendingTradeDraft(log, e)
    }

    e.PendingTradeDraft = &pendingTradeDraft
    e.Broker.Broadcast(log, "TradeDraftOpenedEvent", e.PendingTradeDraft)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   *e.PendingTradeDraft,
    }
}

func ClearPendingTradeDraft(
    log zerolog.Logger,
    e *internal.MonopolyEngine,
) *internal.PendingTradeDraft {
    if e.PendingTradeDraft == nil {
        return nil
    }

    tradeDraft := e.PendingTradeDraft
    e.PendingTradeDraft = nil
    e.Broker.Broadcast(log, "TradeDraftClosedEvent", tradeDraft)
    return tradeDraft
}

func buildTradeEvent(pendingTrade *internal.PendingTrade, accepted bool) internal.Trade {
    return internal.Trade{
        FromPlayerId:        pendingTrade.FromPlayerId,
        ToPlayerId:          pendingTrade.ToPlayerId,
        SessionId:           pendingTrade.SessionId,
        OfferedMoney:        pendingTrade.OfferedMoney,
        RequestedMoney:      pendingTrade.RequestedMoney,
        OfferedProperties:   pendingTrade.OfferedProperties,
        RequestedProperties: pendingTrade.RequestedProperties,
        Accepted:            accepted,
    }
}

func validateTradeSelection(
    log zerolog.Logger,
    ctx context.Context,
    tx *pgxpool.Tx,
    sessionId string,
    ownerId int,
    propertyIds []int,
) ([]internal.TradeProperty, error) {
    selectedProperties := make([]internal.TradeProperty, 0, len(propertyIds))
    seenPropertyIds := map[int]bool{}

    ownedProperties, err := internaldb_players.GetPlayerOwnedProperties(log, ctx, tx, ownerId, sessionId)
    if err != nil {
        return nil, err
    }

    ownedPropertyMap := map[int]internal.OwnedProperty{}
    groupedOwnedProperties := map[string][]internal.OwnedProperty{}
    groupedSelectedPropertyIds := map[string]map[int]bool{}

    for _, ownedProperty := range ownedProperties {
        ownedPropertyMap[ownedProperty.PropertyInfo.Id] = ownedProperty
        groupedOwnedProperties[ownedProperty.PropertyInfo.PropertyType] = append(groupedOwnedProperties[ownedProperty.PropertyInfo.PropertyType], ownedProperty)
    }

    for _, propertyId := range propertyIds {
        if propertyId <= 0 {
            return nil, fmt.Errorf("invalid property in trade")
        }

        if seenPropertyIds[propertyId] {
            return nil, fmt.Errorf("duplicate property in trade")
        }
        seenPropertyIds[propertyId] = true

        ownedProperty, ok := ownedPropertyMap[propertyId]
        if !ok {
            return nil, fmt.Errorf("player does not own one or more traded properties")
        }

        propertyType := ownedProperty.PropertyInfo.PropertyType
        if groupedSelectedPropertyIds[propertyType] == nil {
            groupedSelectedPropertyIds[propertyType] = map[int]bool{}
        }
        groupedSelectedPropertyIds[propertyType][propertyId] = true

        selectedProperties = append(selectedProperties, internal.TradeProperty{
            PropertyId: propertyId,
            Name:       ownedProperty.PropertyInfo.Name,
        })
    }

    for propertyType, groupProperties := range groupedOwnedProperties {
        if propertyType == "RAILROAD" || propertyType == "UTILITY" {
            continue
        }

        hasBuildings := false
        for _, groupProperty := range groupProperties {
            if groupProperty.Houses > 0 || groupProperty.HasHotel {
                hasBuildings = true
                break
            }
        }

        if !hasBuildings {
            continue
        }

        selectedIds := groupedSelectedPropertyIds[propertyType]
        if len(selectedIds) == 0 {
            continue
        }

        if len(selectedIds) != len(groupProperties) {
            return nil, fmt.Errorf("properties with buildings must be traded as a full color set")
        }

        for _, groupProperty := range groupProperties {
            if !selectedIds[groupProperty.PropertyInfo.Id] {
                return nil, fmt.Errorf("properties with buildings must be traded as a full color set")
            }
        }
    }

    return selectedProperties, nil
}

func SetPendingCardDraw(
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    playerId int,
    sessionId string,
    cardType string,
    tileName string,
    position int,
    diceTotal int,
) internal.UserActionStatus {
    if e.PendingCardDraw != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there is already a pending card draw",
        }
    }

    if e.PendingDrawnCard != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there is already a drawn card awaiting resolution",
        }
    }

    e.PendingCardDraw = &internal.PendingCardDraw{
        PlayerId:  playerId,
        SessionId: sessionId,
        CardType:  cardType,
        TileName:  tileName,
        Position:  position,
        DiceTotal: diceTotal,
    }

    e.Broker.Broadcast(log, "CardDrawAvailableEvent", e.PendingCardDraw)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   *e.PendingCardDraw,
    }
}

func clearPendingCardState(e *internal.MonopolyEngine) {
    e.PendingCardDraw = nil
    e.PendingDrawnCard = nil
}

func finalizeLanding(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    tx *pgxpool.Tx,
    playerId int,
    sessionId string,
    diceTotal int,
    playerMovement internal.PlayerMovement,
) internal.UserActionStatus {
    skipMoveEvent := consumeSkipMoveEvent(e, playerId)
    setGoPayout := func() internal.UserActionStatus {
        if !playerMovement.FromCard && (playerMovement.PassedGo || playerMovement.NewPosition == 0) {
            payoutStatus := SetPendingBankPayout(log, e, playerId, sessionId, 200, "Collect as you pass Go")
            if payoutStatus.Status != http.StatusOK {
                return payoutStatus
            }
        }

        return internal.UserActionStatus{Status: http.StatusOK}
    }

    tileData, err := internaldb_tiles.GetRentTileData(log, ctx, tx, sessionId, playerMovement.NewPosition)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    playerMovement.PropertyId = tileData.PropertyId

    if playerMovement.FromCard {
        if playerMovement.NewPosition == 10 {
            player, err := internaldb_players.GetPlayer(log, ctx, tx, playerId, sessionId)
            if err != nil {
                return internal.UserActionStatus{
                    Status: http.StatusInternalServerError,
                    Msg:    err.Error(),
                }
            }

            if player.Jailed > 0 {
                e.ExtraRollAllowed[playerId] = false
                e.DoubleRollCounts[playerId] = 0
                playerMovement.RollAgain = false
            } else {
                playerMovement.RollAgain = e.ExtraRollAllowed[playerId]
            }
        } else {
            playerMovement.RollAgain = e.ExtraRollAllowed[playerId]
        }
    }

    e.PendingPropertyPurchase = nil

    if shouldDeferLandingUntilAfterGoPayout(tileData, playerId, playerMovement) {
        if !skipMoveEvent {
            e.Broker.Broadcast(log, "MovePlayerEvent", playerMovement)
        }
        payoutStatus := setGoPayout()
        if payoutStatus.Status != http.StatusOK {
            return payoutStatus
        }

        setDeferredLanding(e, playerId, sessionId, diceTotal, playerMovement)
        events.EmitGameBoardUpdate(log, ctx, e, tx)

        return internal.UserActionStatus{
            Status: http.StatusOK,
            Data:   playerMovement,
        }
    }

    if playerMovement.NewPosition == 30 {
        err = internaldb_players.UpdatePlayerPositionAndJailed(log, ctx, tx, playerId, sessionId, 10, 1)
        if err != nil {
            return internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg:    err.Error(),
            }
        }

        e.ExtraRollAllowed[playerId] = false
        e.DoubleRollCounts[playerId] = 0
        playerMovement.NewPosition = 10
        playerMovement.PropertyId = 0
        playerMovement.RollAgain = false

        if !skipMoveEvent {
            e.Broker.Broadcast(log, "MovePlayerEvent", playerMovement)
        }
        payoutStatus := setGoPayout()
        if payoutStatus.Status != http.StatusOK {
            return payoutStatus
        }
        events.EmitGameBoardUpdate(log, ctx, e, tx)

        return internal.UserActionStatus{
            Status: http.StatusOK,
            Data:   playerMovement,
        }
    }

    if playerMovement.NewPosition == 4 {
        paymentStatus := SetPendingBankPayment(log, e, playerId, sessionId, 200, "Tuition")
        if paymentStatus.Status != http.StatusOK {
            return paymentStatus
        }
    }

    if playerMovement.NewPosition == 20 {
        paymentStatus := SetPendingBankPayment(log, e, playerId, sessionId, 100, "Parking Ticket")
        if paymentStatus.Status != http.StatusOK {
            return paymentStatus
        }
    }

    if playerMovement.NewPosition == 38 {
        paymentStatus := SetPendingBankPayment(log, e, playerId, sessionId, 100, "Textbooks")
        if paymentStatus.Status != http.StatusOK {
            return paymentStatus
        }
    }

    if tileData.Name == "Community Chest" || tileData.Name == "Chance" {
        cardType := "COMMUNITY"
        if tileData.Name == "Chance" {
            cardType = "CHANCE"
        }

        if !skipMoveEvent {
            e.Broker.Broadcast(log, "MovePlayerEvent", playerMovement)
        }
        drawStatus := SetPendingCardDraw(log, e, playerId, sessionId, cardType, tileData.Name, playerMovement.NewPosition, diceTotal)
        if drawStatus.Status != http.StatusOK {
            return drawStatus
        }
        payoutStatus := setGoPayout()
        if payoutStatus.Status != http.StatusOK {
            return payoutStatus
        }

        events.EmitGameBoardUpdate(log, ctx, e, tx)

        return internal.UserActionStatus{
            Status: http.StatusOK,
            Data:   playerMovement,
        }
    }

    if tileData.HasProperty && tileData.Owned && tileData.OwnerId != playerId && !tileData.IsMortgaged {
        isUtilityCard, _ := e.TempStore["special_utility_rent"].(bool)
        isRailroadCard, _ := e.TempStore["special_railroad_rent"].(bool)

        rentAmount, err := getRentAmount(ctx, log, tx, sessionId, tileData, diceTotal, isUtilityCard, isRailroadCard)
        if err != nil {
            return internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg:    err.Error(),
            }
        }

        delete(e.TempStore, "special_utility_rent")
        delete(e.TempStore, "special_railroad_rent")

        if rentAmount > 0 {
            e.PendingRent = &internal.PendingRent{
                FromPlayerId:   playerId,
                ToPlayerId:     tileData.OwnerId,
                SessionId:      sessionId,
                PropertyId:     tileData.PropertyId,
                Position:       playerMovement.NewPosition,
                Amount:         rentAmount,
                DiceTotal:      diceTotal,
                IsUtilityCard:  isUtilityCard,
                IsRailroadCard: isRailroadCard,
            }

            playerMovement.RentDue = true
            playerMovement.RentAmount = rentAmount
            playerMovement.RentToId = tileData.OwnerId
            playerMovement.PropertyId = tileData.PropertyId

            if !skipMoveEvent {
                e.Broker.Broadcast(log, "MovePlayerEvent", playerMovement)
            }
            e.Broker.Broadcast(log, "RentDueEvent", e.PendingRent)
            payoutStatus := setGoPayout()
            if payoutStatus.Status != http.StatusOK {
                return payoutStatus
            }
            events.EmitGameBoardUpdate(log, ctx, e, tx)

            return internal.UserActionStatus{
                Status: http.StatusOK,
                Data:   playerMovement,
            }
        }
    }

    if !skipMoveEvent {
        e.Broker.Broadcast(log, "MovePlayerEvent", playerMovement)
    }
    payoutStatus := setGoPayout()
    if payoutStatus.Status != http.StatusOK {
        return payoutStatus
    }

    if tileData.HasProperty && !tileData.Owned && tileData.PropertyId > 0 {
        currentPlayer, err := internaldb_players.GetPlayer(log, ctx, tx, playerId, sessionId)
        if err != nil {
            return internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg:    err.Error(),
            }
        }

        propertyData, err := internaldb_tiles.GetPropertyData(log, ctx, tx, sessionId, tileData.PropertyId)
        if err != nil {
            return internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg:    err.Error(),
            }
        }

        e.PendingPropertyPurchase = &internal.PendingPropertyPurchase{
            PlayerId:     playerId,
            SessionId:    sessionId,
            PropertyId:   tileData.PropertyId,
            PurchaseCost: propertyData.PurchaseCost,
            PlayerMoney:  currentPlayer.Money,
            CanAfford:    currentPlayer.Money >= propertyData.PurchaseCost,
        }

        e.Broker.Broadcast(log, "PropertyPurchaseAvailableEvent", e.PendingPropertyPurchase)
    }

    events.EmitGameBoardUpdate(log, ctx, e, tx)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   playerMovement,
    }
}

func clearPlayerTurnState(e *internal.MonopolyEngine, playerId int) {
    delete(e.PendingRolls, playerId)
    delete(e.TurnHasRolled, playerId)
    delete(e.ExtraRollAllowed, playerId)
    delete(e.DoubleRollCounts, playerId)
    clearDeferredLanding(e, playerId)
    consumeSkipMoveEvent(e, playerId)
}

func getActionPlayers(
    ctx context.Context,
    log zerolog.Logger,
    tx *pgxpool.Tx,
    sessionId string,
    playerId int,
) (*internal.Player, *internal.Player, []internal.Player, int, *internal.UserActionStatus) {
    currentPlayer, actingPlayer, players, turnNumber, err := turn_events.GetActionPlayers(ctx, log, tx, sessionId, playerId)
    if err != nil {
        return nil, nil, nil, 0, &internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    if actingPlayer.Bankrupt {
        return nil, nil, nil, 0, &internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player is bankrupt",
        }
    }

    if currentPlayer == nil {
        return nil, nil, nil, 0, &internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there are no players in this game session",
        }
    }

    if currentPlayer.Id != playerId {
        return nil, nil, nil, 0, &internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "it is not this player's turn",
        }
    }

    return currentPlayer, actingPlayer, players, turnNumber, nil
}

// getNewPosition returns the player's new board position after moving by the
// given total and reports whether the move passed Go.
// The board is treated as a 40-tile loop.
func getNewPosition(position int, total int) (int, bool) {
    newPosition := (position + total) % 40
    return newPosition, position+total >= 40
}

// Connected handles a player joining the running game session.
// It verifies that the player exists in the session, marks them as in-game,
// notifies connected clients, and sends the full initial board state.
func Connected(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    log.Trace().Msg("player attempting to join game")

    data := action.Data.(struct {
        Id         int
        PlayerName string
        SessionId  string
    })

    // ensure player exists in session
    player_exists, err := internaldb_players.CheckPlayerExists(
        log,
        ctx,
        tx,
        data.Id,
        data.PlayerName,
        data.SessionId,
    )
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    if !player_exists {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player does not exist",
        }
    }

    internaldb_players.UpdatePlayerInGameStatus(log, ctx, tx, data.Id, data.SessionId, true)

    // announce to all connected users that another user has joined the game
    e.Broker.BroadcastComment(log, fmt.Sprintf("Player %v has joined", data.PlayerName))

    err = events.EmitInitialGameBoardData(log, ctx, e, tx, data)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    log.Trace().Msgf("player %v has joined server", data.PlayerName)
    return internal.UserActionStatus{
        Status: http.StatusOK,
    }
}

// Disconnected handles a player leaving the running game session.
// It verifies that the player exists, marks them as out of the game, updates
// ready status when needed, notifies connected clients, and broadcasts a board
// state update.
func Disconnected(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {

    log.Trace().Msg("player attempting to leave game")

    data := action.Data.(struct {
        Id         int
        PlayerName string
        SessionId  string
    })

    // ensure player exists in session
    player_exists, err := internaldb_players.CheckPlayerExists(
        log,
        ctx,
        tx,
        data.Id,
        data.PlayerName,
        data.SessionId,
    )
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    if !player_exists {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player does not exist",
        }
    }

    internaldb_players.UpdatePlayerInGameStatus(log, ctx, tx, data.Id, data.SessionId, false)

    // check if all players are ready
    ready_stats, err := internaldb_players.GetAllPlayersReadyUpStatus(
        log,
        ctx,
        tx,
        data.SessionId,
    )
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Data:   err.Error(),
        }
    }

    if ready_stats.Ready != ready_stats.Total {
        // everyone is ready, don't reset ready up
        internaldb_players.SetPlayerReadyUpStatus(log, ctx, tx, data.Id, data.SessionId, false)
    }

    // announce to all connected users that another user has left the game
    e.Broker.BroadcastComment(log, fmt.Sprintf("Player %v has left", data.PlayerName))

    // only emit simple board update since you don't need to re-emit all the tile
    // info on disconnect
    err = events.EmitGameBoardUpdate(log, ctx, e, tx)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    log.Trace().Msgf("player %v has left server", data.PlayerName)
    return internal.UserActionStatus{
        Status: http.StatusOK,
    }
}

// ReadyUp updates a player's ready-up status for the session.
// After updating the player's state, it broadcasts a board update and starts
// the game when all players in the session are ready.
func ReadyUp(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(struct {
        SessionId string
        PlayerId  int
        Status    bool
    })

    err := internaldb_players.SetPlayerReadyUpStatus(
        log,
        ctx,
        tx,
        data.PlayerId,
        data.SessionId,
        data.Status,
    )
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    events.EmitGameBoardUpdate(log, ctx, e, tx)

    // check if all players are ready
    ready_stats, err := internaldb_players.GetAllPlayersReadyUpStatus(
        log,
        ctx,
        tx,
        data.SessionId,
    )
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Data:   err.Error(),
        }
    }

    if ready_stats.Ready == ready_stats.Total {
        // everyone is ready
        internaldb_game_state.UpdateGameStateTurnNumber(log, ctx, tx, data.SessionId, 0)
        e.TurnNumber = 0
        e.TempStore["turn_decision_rolls"] = make([]internal.DiceRoll, 0)
        log.Info().Msg("final player has readied up; Start Game")
        e.Broker.Broadcast(log, "GameReadyEvent", "START")
    }

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   "",
    }
}

// RollDice rolls the dice for the current player and stores the result in the
// engine state.
// During turn-order setup, the roll is added to the temporary turn-decision
// list. During normal gameplay, the roll becomes a pending move for that
// player. The result is then broadcast to connected clients.
func RollDice(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(internal.SimpleActionData)

    _, player, players, _, status := getActionPlayers(ctx, log, tx, data.SessionId, data.PlayerId)
    if status != nil {
        return *status
    }

    if e.TurnNumber >= len(players) && e.TurnHasRolled[data.PlayerId] && !e.ExtraRollAllowed[data.PlayerId] {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player has already rolled this turn",
        }
    }

    diceRoll := internal.DiceRoll{
        PlayerId:  data.PlayerId,
        SessionId: data.SessionId,
        DieOne:    rand.IntN(6) + 1,
        DieTwo:    rand.IntN(6) + 1,
    }

    diceRoll.Total = diceRoll.DieOne + diceRoll.DieTwo
    diceRoll.IsDouble = diceRoll.DieOne == diceRoll.DieTwo

    if e.TurnNumber >= 0 && e.TurnNumber < len(players) {
        rolls := e.TempStore["turn_decision_rolls"].([]internal.DiceRoll)
        rolls = append(rolls, diceRoll)

        // sort rolls so highest roll is first
        sort.SliceStable(rolls, func(i, j int) bool {
            if rolls[i].Total == rolls[j].Total {
                if rolls[i].DieOne == rolls[j].DieOne {
                    if rolls[i].DieTwo == rolls[j].DieTwo {
                        return rolls[i].PlayerId < rolls[j].PlayerId
                    }
                    return rolls[i].DieTwo > rolls[j].DieTwo
                }
                return rolls[i].DieOne > rolls[j].DieOne
            }
            return rolls[i].Total > rolls[j].Total
        })

        e.TempStore["turn_decision_rolls"] = rolls
        e.TurnHasRolled[data.PlayerId] = true
        e.PendingRolls[data.PlayerId] = diceRoll
    } else {
        if e.PendingRent != nil && e.PendingRent.FromPlayerId == data.PlayerId {
            return internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg:    "player has a pending rent payment",
            }
        }

        if e.PendingCardDraw != nil && e.PendingCardDraw.PlayerId == data.PlayerId {
            return internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg:    "player has a pending card draw",
            }
        }

        if e.PendingDrawnCard != nil && e.PendingDrawnCard.PlayerId == data.PlayerId {
            return internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg:    "player has a drawn card awaiting resolution",
            }
        }

        if e.PendingBankPayment != nil && e.PendingBankPayment.PlayerId == data.PlayerId {
            return internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg:    "player has a pending bank payment",
            }
        }

        if _, ok := e.PendingRolls[data.PlayerId]; ok {
            return internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg:    "player already has a pending dice roll",
            }
        }

        e.TurnHasRolled[data.PlayerId] = true
        e.ExtraRollAllowed[data.PlayerId] = false

        if player.Jailed > 0 {
            if diceRoll.IsDouble {
                err := internaldb_players.UpdatePlayerJailState(
                    log,
                    ctx,
                    tx,
                    data.PlayerId,
                    data.SessionId,
                    player.GetOutOfJailCards,
                    0,
                )
                if err != nil {
                    return internal.UserActionStatus{
                        Status: http.StatusInternalServerError,
                        Msg:    err.Error(),
                    }
                }

                e.DoubleRollCounts[data.PlayerId] += 1
                diceRoll.ReleasedFromJail = true
                diceRoll.Jailed = 0
                e.PendingRolls[data.PlayerId] = diceRoll
            } else {
                jailedTurns := player.Jailed
                if jailedTurns < 4 {
                    jailedTurns += 1
                    err := internaldb_players.UpdatePlayerJailState(
                        log,
                        ctx,
                        tx,
                        data.PlayerId,
                        data.SessionId,
                        player.GetOutOfJailCards,
                        jailedTurns,
                    )
                    if err != nil {
                        return internal.UserActionStatus{
                            Status: http.StatusInternalServerError,
                            Msg:    err.Error(),
                        }
                    }
                }

                diceRoll.Jailed = jailedTurns
                e.DoubleRollCounts[data.PlayerId] = 0
                e.PendingRolls[data.PlayerId] = diceRoll
            }
        } else {
            if diceRoll.IsDouble {
                e.DoubleRollCounts[data.PlayerId] += 1
                if e.DoubleRollCounts[data.PlayerId] >= 3 {
                    err := internaldb_players.UpdatePlayerPositionAndJailed(log, ctx, tx, data.PlayerId, data.SessionId, 10, 1)
                    if err != nil {
                        return internal.UserActionStatus{
                            Status: http.StatusInternalServerError,
                            Msg:    err.Error(),
                        }
                    }

                    e.ExtraRollAllowed[data.PlayerId] = false
                    diceRoll.SentToJail = true
                    diceRoll.Jailed = 1
                } else {
                    diceRoll.RollAgain = true
                    e.PendingRolls[data.PlayerId] = diceRoll
                }
            } else {
                e.DoubleRollCounts[data.PlayerId] = 0
                e.PendingRolls[data.PlayerId] = diceRoll
            }
        }
    }

    e.Broker.Broadcast(log, "RollDiceEvent", diceRoll)
    events.EmitGameBoardUpdate(log, ctx, e, tx)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   diceRoll,
    }
}

// EndTurn advances the game to the next turn for the current player.
// While the game is still deciding player order, it also assigns final turn
// order once all setup rolls have been completed. After updating turn state,
// it broadcasts the latest board update.
func EndTurn(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {

    data := action.Data.(internal.SimpleActionData)

    _, _, players, _, status := getActionPlayers(ctx, log, tx, data.SessionId, data.PlayerId)
    if status != nil {
        return *status
    }

    if e.PendingBankPayment != nil && e.PendingBankPayment.PlayerId == data.PlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player has a pending bank payment",
        }
    }

    if e.PendingBankPayout != nil && e.PendingBankPayout.PlayerId == data.PlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player has a pending bank payout",
        }
    }

    if e.PendingRent != nil && e.PendingRent.FromPlayerId == data.PlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player has a pending rent payment",
        }
    }

    if e.PendingCardDraw != nil && e.PendingCardDraw.PlayerId == data.PlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player has a pending card draw",
        }
    }

    if e.PendingDrawnCard != nil && e.PendingDrawnCard.PlayerId == data.PlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player has a drawn card awaiting resolution",
        }
    }

    if e.TurnNumber < len(players) {
        if !e.TurnHasRolled[data.PlayerId] {
            return internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg:    "player must roll for turn order",
            }
        }

        delete(e.PendingRolls, data.PlayerId)
    } else {
        if pendingRoll, ok := e.PendingRolls[data.PlayerId]; ok {
            player, err := internaldb_players.GetPlayer(log, ctx, tx, data.PlayerId, data.SessionId)
            if err != nil {
                return internal.UserActionStatus{
                    Status: http.StatusInternalServerError,
                    Msg:    err.Error(),
                }
            }

            if !(player.Jailed > 0 && !pendingRoll.IsDouble && !pendingRoll.ReleasedFromJail) {
                return internal.UserActionStatus{
                    Status: http.StatusBadRequest,
                    Msg:    "player has a pending dice roll",
                }
            }

            delete(e.PendingRolls, data.PlayerId)
        }

        if !e.TurnHasRolled[data.PlayerId] {
            return internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg:    "player must roll to start the turn",
            }
        }

        if e.ExtraRollAllowed[data.PlayerId] {
            return internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg:    "player must roll again",
            }
        }
    }

    if e.PendingPropertyPurchase != nil && e.PendingPropertyPurchase.PlayerId == data.PlayerId && e.PendingPropertyPurchase.SessionId == data.SessionId {
        e.Broker.Broadcast(log, "PropertyPurchaseIgnoredEvent", e.PendingPropertyPurchase)
        e.PendingPropertyPurchase = nil
    }

    err := internaldb_game_state.UpdateGameStateTurnNumber(log, ctx, tx, data.SessionId, e.TurnNumber+1)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }
    // turn update succeeded, update internal counter
    e.TurnNumber += 1

    // all players have rolled to determine their turn order
    if e.TurnNumber == len(players) {
        // give all players their turn order
        for i, roll := range e.TempStore["turn_decision_rolls"].([]internal.DiceRoll) {
            log.Trace().Msgf("%#v", roll)
            err := internaldb_players.UpdatePlayerTurnOrder(
                log,
                ctx,
                tx,
                roll.PlayerId,
                roll.SessionId,
                i,
            )
            if err != nil {
                return internal.UserActionStatus{
                    Status: http.StatusInternalServerError,
                    Msg:    err.Error(),
                }
            }
        }
    }

    // FIXME: Prevent ending turn if there is a pending rent. If player can't pay
    // rent and is bankrupt, handle that here.

    clearPlayerTurnState(e, data.PlayerId)

    events.EmitGameBoardUpdate(log, ctx, e, tx)

    return internal.UserActionStatus{
        Status: http.StatusOK,
    }
}

// MovePlayer moves the current player using their pending dice roll.
// It validates that the player has already rolled, updates their board
// position, clears the pending roll, checks whether rent is now owed, and
// broadcasts either the movement alone or both the movement and a rent-due
// event.
func MovePlayer(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(internal.SimpleActionData)

    _, _, _, turnNumber, status := getActionPlayers(ctx, log, tx, data.SessionId, data.PlayerId)
    if status != nil {
        return *status
    }

    if e.PendingBankPayment != nil && e.PendingBankPayment.PlayerId == data.PlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player has a pending bank payment",
        }
    }

    diceRoll, ok := e.PendingRolls[data.PlayerId]
    if !ok {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player does not have a pending dice roll",
        }
    }

    player, err := internaldb_players.GetPlayer(log, ctx, tx, data.PlayerId, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    if player.Jailed > 0 {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player is in jail",
        }
    }

    newPosition, passedGo := getNewPosition(player.Position, diceRoll.Total)
    err = internaldb_players.UpdatePlayerPosition(log, ctx, tx, data.PlayerId, data.SessionId, newPosition)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    delete(e.PendingRolls, data.PlayerId)

    playerMovement := internal.PlayerMovement{
        PlayerId:    data.PlayerId,
        SessionId:   data.SessionId,
        OldPosition: player.Position,
        NewPosition: newPosition,
        Total:       diceRoll.Total,
        PassedGo:    passedGo,
        TurnNumber:  turnNumber,
    }

    if diceRoll.IsDouble {
        e.ExtraRollAllowed[data.PlayerId] = true
        playerMovement.RollAgain = true
    } else {
        e.ExtraRollAllowed[data.PlayerId] = false
        e.DoubleRollCounts[data.PlayerId] = 0
    }

    return finalizeLanding(ctx, log, e, tx, data.PlayerId, data.SessionId, diceRoll.Total, playerMovement)
}

func DrawCard(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(internal.CardActionData)

    _, _, _, _, status := getActionPlayers(ctx, log, tx, data.SessionId, data.PlayerId)
    if status != nil {
        return *status
    }

    if e.PendingCardDraw == nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there is no pending card draw",
        }
    }

    if e.PendingCardDraw.PlayerId != data.PlayerId || e.PendingCardDraw.SessionId != data.SessionId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "card draw player is incorrect",
        }
    }

    cardId, err := internaldb_event_cards.AssignEventCardDB(log, ctx, tx, data.SessionId, e.PendingCardDraw.CardType, data.PlayerId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    card, err := internaldb_event_cards.GetEventCard(log, ctx, tx, cardId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    e.PendingDrawnCard = &internal.DrawnCard{
        PlayerId:  data.PlayerId,
        SessionId: data.SessionId,
        DiceTotal: e.PendingCardDraw.DiceTotal,
        EventCard: card,
    }
    e.PendingCardDraw = nil

    e.Broker.Broadcast(log, "DrawCardEvent", e.PendingDrawnCard)
    events.EmitGameBoardUpdate(log, ctx, e, tx)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   *e.PendingDrawnCard,
    }
}

func ResolveCard(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(internal.CardActionData)

    _, _, _, _, status := getActionPlayers(ctx, log, tx, data.SessionId, data.PlayerId)
    if status != nil {
        return *status
    }

    if e.PendingDrawnCard == nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there is no drawn card awaiting resolution",
        }
    }

    if e.PendingDrawnCard.PlayerId != data.PlayerId || e.PendingDrawnCard.SessionId != data.SessionId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "card resolve player is incorrect",
        }
    }

    drawnCard := *e.PendingDrawnCard
    delete(e.TempStore, "card_movement")
    effectFunc, exists := CardEffects[drawnCard.Id]
    if exists {
        status := effectFunc(ctx, log, e, tx, data.SessionId, data.PlayerId)
        if status.Status != http.StatusOK {
            return status
        }
    }

    e.Broker.Broadcast(log, "CardResolvedEvent", drawnCard)
    e.PendingDrawnCard = nil

    cardMovement, ok := e.TempStore["card_movement"].(internal.PlayerMovement)
    if ok {
        delete(e.TempStore, "card_movement")
        return finalizeLanding(ctx, log, e, tx, data.PlayerId, data.SessionId, drawnCard.DiceTotal, cardMovement)
    }

    events.EmitGameBoardUpdate(log, ctx, e, tx)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   drawnCard,
    }
}

func PayBank(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(internal.BankPaymentActionData)

    if e.PendingBankPayment == nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there is no pending bank payment",
        }
    }

    _, _, _, _, status := getActionPlayers(ctx, log, tx, data.SessionId, data.PlayerId)
    if status != nil {
        return *status
    }

    if e.PendingBankPayment.PlayerId != data.PlayerId || e.PendingBankPayment.SessionId != data.SessionId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "bank payment payer is incorrect",
        }
    }

    player, err := internaldb_players.GetPlayer(log, ctx, tx, data.PlayerId, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    if player.Money < e.PendingBankPayment.Amount {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player does not have enough money",
        }
    }

    err = internaldb_players.UpdatePlayerMoney(
        log,
        ctx,
        tx,
        data.PlayerId,
        data.SessionId,
        player.Money-e.PendingBankPayment.Amount,
    )
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    bankPayment := internal.BankPayment{
        PlayerId:    data.PlayerId,
        SessionId:   data.SessionId,
        Amount:      e.PendingBankPayment.Amount,
        Reason:      e.PendingBankPayment.Reason,
        PlayerMoney: player.Money - e.PendingBankPayment.Amount,
        Jailed:      player.Jailed,
    }

    if e.PendingBankPayment.Reason == "jail release" {
        err = internaldb_players.UpdatePlayerJailState(
            log,
            ctx,
            tx,
            data.PlayerId,
            data.SessionId,
            player.GetOutOfJailCards,
            0,
        )
        if err != nil {
            return internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg:    err.Error(),
            }
        }

        delete(e.PendingRolls, data.PlayerId)
        e.TurnHasRolled[data.PlayerId] = false
        e.ExtraRollAllowed[data.PlayerId] = false
        e.DoubleRollCounts[data.PlayerId] = 0
        bankPayment.Jailed = 0
        e.Broker.Broadcast(log, "PayToLeaveJailEvent", internal.JailRelease{
            PlayerId:          data.PlayerId,
            SessionId:         data.SessionId,
            Method:            "pay",
            GetOutOfJailCards: player.GetOutOfJailCards,
            PlayerMoney:       bankPayment.PlayerMoney,
            Jailed:            0,
        })
    }

    e.PendingBankPayment = nil

    e.Broker.Broadcast(log, "BankPaymentEvent", bankPayment)
    events.EmitGameBoardUpdate(log, ctx, e, tx)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   bankPayment,
    }
}

func SetBankPayout(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(internal.BankPayoutActionData)

    _, _, _, _, status := getActionPlayers(ctx, log, tx, data.SessionId, data.PlayerId)
    if status != nil {
        return *status
    }

    if data.Amount < 1 {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "amount must be greater than 0",
        }
    }

    return SetPendingBankPayout(log, e, data.PlayerId, data.SessionId, data.Amount, data.Reason)
}

func ReceiveBankPayout(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(internal.BankPaymentActionData)

    if e.PendingBankPayout == nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there is no pending bank payout",
        }
    }

    _, _, _, _, status := getActionPlayers(ctx, log, tx, data.SessionId, data.PlayerId)
    if status != nil {
        return *status
    }

    if e.PendingBankPayout.PlayerId != data.PlayerId || e.PendingBankPayout.SessionId != data.SessionId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "bank payout recipient is incorrect",
        }
    }

    player, err := internaldb_players.GetPlayer(log, ctx, tx, data.PlayerId, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    err = internaldb_players.UpdatePlayerMoney(
        log,
        ctx,
        tx,
        data.PlayerId,
        data.SessionId,
        player.Money+e.PendingBankPayout.Amount,
    )
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    bankPayout := internal.BankPayout{
        PlayerId:    data.PlayerId,
        SessionId:   data.SessionId,
        Amount:      e.PendingBankPayout.Amount,
        Reason:      e.PendingBankPayout.Reason,
        PlayerMoney: player.Money + e.PendingBankPayout.Amount,
    }

    e.PendingBankPayout = nil
    deferredLanding, hasDeferredLanding := getDeferredLanding(e, data.PlayerId)
    if hasDeferredLanding {
        clearDeferredLanding(e, data.PlayerId)
    }

    e.Broker.Broadcast(log, "BankPayoutEvent", bankPayout)

    if hasDeferredLanding {
        deferredLanding.PlayerMovement.PassedGo = false
        setSkipMoveEvent(e, data.PlayerId)
        status := finalizeLanding(
            ctx,
            log,
            e,
            tx,
            deferredLanding.PlayerId,
            deferredLanding.SessionId,
            deferredLanding.DiceTotal,
            deferredLanding.PlayerMovement,
        )
        if status.Status != http.StatusOK {
            return status
        }

        return internal.UserActionStatus{
            Status: http.StatusOK,
            Data:   bankPayout,
        }
    }

    events.EmitGameBoardUpdate(log, ctx, e, tx)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   bankPayout,
    }
}

func ExecutePlayerExchange(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(internal.PlayerExchangeActionData)

    if e.PendingExchange == nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there is no pending player exchange",
        }
    }

    _, _, _, _, status := getActionPlayers(ctx, log, tx, data.SessionId, data.PlayerId)
    if status != nil {
        return *status
    }

    if e.PendingExchange.ActingPlayerId != data.PlayerId || e.PendingExchange.SessionId != data.SessionId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "acting player for exchange is incorrect",
        }
    }

    players, err := internaldb_players.GetPlayersInSession(log, ctx, tx, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{Status: http.StatusInternalServerError, Msg: err.Error()}
    }

    balances := make(map[int]int)
    actingPlayer, err := internaldb_players.GetPlayer(log, ctx, tx, data.PlayerId, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{Status: http.StatusInternalServerError, Msg: err.Error()}
    }

    totalActingAdjustment := 0

    if e.PendingExchange.IsPayingAll {
        if actingPlayer.Money < e.PendingExchange.Amount*(len(players)-1) {
            return internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg:    "player does not have enough money to pay everyone",
            }
        }
    } else {
        for _, p := range players {
            if p.Id != data.PlayerId && p.Money < e.PendingExchange.Amount {
                return internal.UserActionStatus{
                    Status: http.StatusBadRequest,
                    Msg:    fmt.Sprintf("player %s does not have enough money", p.Name),
                }
            }
        }
    }

    for _, p := range players {
        if p.Id == data.PlayerId {
            continue
        }

        if e.PendingExchange.IsPayingAll {
            totalActingAdjustment -= e.PendingExchange.Amount
            err = internaldb_players.UpdatePlayerMoney(log, ctx, tx, p.Id, data.SessionId, p.Money+e.PendingExchange.Amount)
            balances[p.Id] = p.Money + e.PendingExchange.Amount
        } else {
            totalActingAdjustment += e.PendingExchange.Amount
            err = internaldb_players.UpdatePlayerMoney(log, ctx, tx, p.Id, data.SessionId, p.Money-e.PendingExchange.Amount)
            balances[p.Id] = p.Money - e.PendingExchange.Amount
        }

        if err != nil {
            return internal.UserActionStatus{Status: http.StatusInternalServerError, Msg: err.Error()}
        }
    }

    err = internaldb_players.UpdatePlayerMoney(log, ctx, tx, data.PlayerId, data.SessionId, actingPlayer.Money+totalActingAdjustment)
    if err != nil {
        return internal.UserActionStatus{Status: http.StatusInternalServerError, Msg: err.Error()}
    }
    balances[data.PlayerId] = actingPlayer.Money + totalActingAdjustment

    exchange := internal.PlayerExchange{
        ActingPlayerId: data.PlayerId,
        SessionId:      data.SessionId,
        Amount:         e.PendingExchange.Amount,
        Reason:         e.PendingExchange.Reason,
        IsPayingAll:    e.PendingExchange.IsPayingAll,
        Balances:       balances,
    }

    e.PendingExchange = nil
    e.Broker.Broadcast(log, "PlayerExchangeEvent", exchange)
    events.EmitGameBoardUpdate(log, ctx, e, tx)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   exchange,
    }
}

func ProposeTrade(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(internal.TradeActionData)

    _, actingPlayer, _, _, status := getActionPlayers(ctx, log, tx, data.SessionId, data.PlayerId)
    if status != nil {
        return *status
    }

    if data.WithPlayerId == data.PlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player cannot trade with themself",
        }
    }

    if data.OfferedMoney < 0 || data.RequestedMoney < 0 {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "trade money values must be non-negative",
        }
    }

    if data.OfferedMoney == 0 && data.RequestedMoney == 0 && len(data.OfferedPropertyIds) == 0 && len(data.RequestedPropertyIds) == 0 {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "trade must include at least one asset",
        }
    }

    targetPlayer, err := internaldb_players.GetPlayer(log, ctx, tx, data.WithPlayerId, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "target player does not exist",
        }
    }

    if actingPlayer.Bankrupt || targetPlayer.Bankrupt {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "bankrupt players cannot trade",
        }
    }

    offeredProperties, err := validateTradeSelection(log, ctx, tx, data.SessionId, data.PlayerId, data.OfferedPropertyIds)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    err.Error(),
        }
    }

    requestedProperties, err := validateTradeSelection(log, ctx, tx, data.SessionId, data.WithPlayerId, data.RequestedPropertyIds)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    err.Error(),
        }
    }

    pendingTrade := internal.PendingTrade{
        FromPlayerId:        data.PlayerId,
        ToPlayerId:          data.WithPlayerId,
        SessionId:           data.SessionId,
        OfferedMoney:        data.OfferedMoney,
        RequestedMoney:      data.RequestedMoney,
        OfferedProperties:   offeredProperties,
        RequestedProperties: requestedProperties,
    }

    if e.PendingTradeDraft != nil && e.PendingTradeDraft.FromPlayerId == data.PlayerId {
        ClearPendingTradeDraft(log, e)
    }

    tradeStatus := SetPendingTrade(log, e, pendingTrade)
    if tradeStatus.Status != http.StatusOK {
        return tradeStatus
    }

    events.EmitGameBoardUpdate(log, ctx, e, tx)
    return tradeStatus
}

func OpenTradeDraft(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(internal.TradeDraftActionData)

    _, actingPlayer, _, _, status := getActionPlayers(ctx, log, tx, data.SessionId, data.PlayerId)
    if status != nil {
        return *status
    }

    if data.WithPlayerId == data.PlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player cannot trade with themself",
        }
    }

    if e.PendingTrade != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there is already a pending trade",
        }
    }

    targetPlayer, err := internaldb_players.GetPlayer(log, ctx, tx, data.WithPlayerId, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "target player does not exist",
        }
    }

    if actingPlayer.Bankrupt || targetPlayer.Bankrupt {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "bankrupt players cannot trade",
        }
    }

    if e.PendingTradeDraft != nil &&
        e.PendingTradeDraft.FromPlayerId == data.PlayerId &&
        e.PendingTradeDraft.ToPlayerId == data.WithPlayerId &&
        e.PendingTradeDraft.SessionId == data.SessionId {
        return internal.UserActionStatus{
            Status: http.StatusOK,
            Data:   *e.PendingTradeDraft,
        }
    }

    pendingTradeDraft := internal.PendingTradeDraft{
        FromPlayerId: data.PlayerId,
        ToPlayerId:   data.WithPlayerId,
        SessionId:    data.SessionId,
    }

    tradeStatus := SetPendingTradeDraft(log, e, pendingTradeDraft)
    if tradeStatus.Status != http.StatusOK {
        return tradeStatus
    }

    events.EmitGameBoardUpdate(log, ctx, e, tx)
    return tradeStatus
}

func CloseTradeDraft(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(internal.TradeDecisionActionData)

    if e.PendingTradeDraft == nil {
        return internal.UserActionStatus{
            Status: http.StatusOK,
        }
    }

    if e.PendingTradeDraft.SessionId != data.SessionId || e.PendingTradeDraft.FromPlayerId != data.PlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "only the proposing player can close this trade",
        }
    }

    tradeDraft := ClearPendingTradeDraft(log, e)
    events.EmitGameBoardUpdate(log, ctx, e, tx)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   tradeDraft,
    }
}

func AcceptTrade(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(internal.TradeDecisionActionData)

    if e.PendingTrade == nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there is no pending trade",
        }
    }

    if e.PendingTrade.SessionId != data.SessionId || e.PendingTrade.ToPlayerId != data.PlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "only the targeted player can accept this trade",
        }
    }

    fromPlayer, err := internaldb_players.GetPlayer(log, ctx, tx, e.PendingTrade.FromPlayerId, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "trade proposer does not exist",
        }
    }

    toPlayer, err := internaldb_players.GetPlayer(log, ctx, tx, e.PendingTrade.ToPlayerId, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "trade target does not exist",
        }
    }

    if fromPlayer.Bankrupt || toPlayer.Bankrupt {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "bankrupt players cannot trade",
        }
    }

    offeredPropertyIds := make([]int, 0, len(e.PendingTrade.OfferedProperties))
    for _, property := range e.PendingTrade.OfferedProperties {
        offeredPropertyIds = append(offeredPropertyIds, property.PropertyId)
    }

    requestedPropertyIds := make([]int, 0, len(e.PendingTrade.RequestedProperties))
    for _, property := range e.PendingTrade.RequestedProperties {
        requestedPropertyIds = append(requestedPropertyIds, property.PropertyId)
    }

    _, err = validateTradeSelection(log, ctx, tx, data.SessionId, fromPlayer.Id, offeredPropertyIds)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    err.Error(),
        }
    }

    _, err = validateTradeSelection(log, ctx, tx, data.SessionId, toPlayer.Id, requestedPropertyIds)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    err.Error(),
        }
    }

    if fromPlayer.Money < e.PendingTrade.OfferedMoney {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "proposing player does not have enough money",
        }
    }

    if toPlayer.Money < e.PendingTrade.RequestedMoney {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "target player does not have enough money",
        }
    }

    if len(offeredPropertyIds) > 0 {
        err = internaldb_tiles.TransferSpecificOwnedProperties(log, ctx, tx, data.SessionId, offeredPropertyIds, toPlayer.Id)
        if err != nil {
            return internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg:    err.Error(),
            }
        }
    }

    if len(requestedPropertyIds) > 0 {
        err = internaldb_tiles.TransferSpecificOwnedProperties(log, ctx, tx, data.SessionId, requestedPropertyIds, fromPlayer.Id)
        if err != nil {
            return internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg:    err.Error(),
            }
        }
    }

    err = internaldb_players.UpdatePlayerMoney(log, ctx, tx, fromPlayer.Id, data.SessionId, fromPlayer.Money-e.PendingTrade.OfferedMoney+e.PendingTrade.RequestedMoney)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    err = internaldb_players.UpdatePlayerMoney(log, ctx, tx, toPlayer.Id, data.SessionId, toPlayer.Money-e.PendingTrade.RequestedMoney+e.PendingTrade.OfferedMoney)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    trade := buildTradeEvent(e.PendingTrade, true)
    ClearPendingTradeDraft(log, e)
    e.PendingTrade = nil
    e.Broker.Broadcast(log, "TradeAcceptedEvent", trade)
    events.EmitGameBoardUpdate(log, ctx, e, tx)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   trade,
    }
}

func RejectTrade(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(internal.TradeDecisionActionData)

    if e.PendingTrade == nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there is no pending trade",
        }
    }

    if e.PendingTrade.SessionId != data.SessionId || e.PendingTrade.ToPlayerId != data.PlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "only the targeted player can reject this trade",
        }
    }

    trade := buildTradeEvent(e.PendingTrade, false)
    ClearPendingTradeDraft(log, e)
    e.PendingTrade = nil
    e.Broker.Broadcast(log, "TradeRejectedEvent", trade)
    events.EmitGameBoardUpdate(log, ctx, e, tx)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   trade,
    }
}

func CancelTrade(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(internal.TradeDecisionActionData)

    if e.PendingTrade == nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "there is no pending trade",
        }
    }

    if e.PendingTrade.SessionId != data.SessionId || e.PendingTrade.FromPlayerId != data.PlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "only the proposing player can cancel this trade",
        }
    }

    trade := buildTradeEvent(e.PendingTrade, false)
    ClearPendingTradeDraft(log, e)
    e.PendingTrade = nil
    e.Broker.Broadcast(log, "TradeCancelledEvent", trade)
    events.EmitGameBoardUpdate(log, ctx, e, tx)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   trade,
    }
}

func Bankrupt(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(internal.SimpleActionData)

    _, player, players, _, status := getActionPlayers(ctx, log, tx, data.SessionId, data.PlayerId)
    if status != nil {
        return *status
    }

    if e.PendingRent == nil && e.PendingBankPayment == nil {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player does not have a pending payment obligation",
        }
    }

    bankruptcyRank, err := internaldb_players.GetNextBankruptcyRank(log, ctx, tx, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    bankruptcy := internal.Bankruptcy{
        PlayerId:  data.PlayerId,
        SessionId: data.SessionId,
        Rank:      bankruptcyRank,
    }

    if e.PendingRent != nil && e.PendingRent.FromPlayerId == data.PlayerId {
        recipient, err := internaldb_players.GetPlayer(log, ctx, tx, e.PendingRent.ToPlayerId, data.SessionId)
        if err != nil {
            return internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg:    err.Error(),
            }
        }

        err = internaldb_tiles.TransferOwnedProperties(log, ctx, tx, data.SessionId, data.PlayerId, recipient.Id)
        if err != nil {
            return internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg:    err.Error(),
            }
        }

        err = internaldb_players.UpdatePlayerMoney(log, ctx, tx, recipient.Id, data.SessionId, recipient.Money+player.Money)
        if err != nil {
            return internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg:    err.Error(),
            }
        }

        bankruptcy.OwesRent = true
        bankruptcy.RentToId = recipient.Id
        e.PendingRent = nil
    } else if e.PendingBankPayment != nil && e.PendingBankPayment.PlayerId == data.PlayerId {
        err = internaldb_tiles.ReleaseOwnedPropertiesToBank(log, ctx, tx, data.SessionId, data.PlayerId)
        if err != nil {
            return internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg:    err.Error(),
            }
        }

        e.PendingBankPayment = nil
    } else {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player does not have a pending payment obligation",
        }
    }

    err = internaldb_players.UpdatePlayerBankruptcy(log, ctx, tx, data.PlayerId, data.SessionId, true, bankruptcyRank, 0)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    clearPlayerTurnState(e, data.PlayerId)

    err = internaldb_game_state.UpdateGameStateTurnNumber(log, ctx, tx, data.SessionId, e.TurnNumber+1)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }
    e.TurnNumber += 1

    nonBankruptPlayerCount, err := internaldb_players.GetNonBankruptPlayerCount(log, ctx, tx, data.SessionId)
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    if nonBankruptPlayerCount == 1 {
        winner, err := internaldb_players.GetRemainingNonBankruptPlayer(log, ctx, tx, data.SessionId)
        if err != nil {
            return internal.UserActionStatus{
                Status: http.StatusInternalServerError,
                Msg:    err.Error(),
            }
        }

        if winner.Rank == 0 {
            err = internaldb_players.UpdatePlayerRank(log, ctx, tx, winner.Id, data.SessionId, 1)
            if err != nil {
                return internal.UserActionStatus{
                    Status: http.StatusInternalServerError,
                    Msg:    err.Error(),
                }
            }
        }

        bankruptcy.WinnerId = winner.Id
    }

    if e.TurnNumber == len(players) {
        for i, roll := range e.TempStore["turn_decision_rolls"].([]internal.DiceRoll) {
            err := internaldb_players.UpdatePlayerTurnOrder(
                log,
                ctx,
                tx,
                roll.PlayerId,
                roll.SessionId,
                i,
            )
            if err != nil {
                return internal.UserActionStatus{
                    Status: http.StatusInternalServerError,
                    Msg:    err.Error(),
                }
            }
        }
    }

    e.Broker.Broadcast(log, "BankruptcyEvent", bankruptcy)
    events.EmitGameBoardUpdate(log, ctx, e, tx)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   bankruptcy,
    }
}

func ReleaseFromJail(
    ctx context.Context,
    log zerolog.Logger,
    e *internal.MonopolyEngine,
    action *internal.UserActionEvent,
    tx *pgxpool.Tx,
) internal.UserActionStatus {
    data := action.Data.(internal.JailReleaseActionData)

    _, player, _, _, status := getActionPlayers(ctx, log, tx, data.SessionId, data.PlayerId)
    if status != nil {
        return *status
    }

    if player.Jailed == 0 {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player is not in jail",
        }
    }

    if e.PendingBankPayment != nil && e.PendingBankPayment.PlayerId == data.PlayerId {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player has a pending bank payment",
        }
    }

    jailRelease := internal.JailRelease{
        PlayerId:          data.PlayerId,
        SessionId:         data.SessionId,
        Method:            data.Method,
        GetOutOfJailCards: player.GetOutOfJailCards,
        PlayerMoney:       player.Money,
        Jailed:            0,
    }

    diceRoll, ok := e.PendingRolls[data.PlayerId]
    if !ok {
        if data.Method == "pay" {
            return internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg:    "player must roll before paying to get out of jail",
            }
        }

        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player must roll before using a get out of jail card",
        }
    }

    if data.Method == "pay" {
        if diceRoll.DieOne == diceRoll.DieTwo {
            return internal.UserActionStatus{
                Status: http.StatusBadRequest,
                Msg:    "player rolled doubles and does not need to pay to get out of jail",
            }
        }

        if player.Money < 50 {
            pendingBankPaymentStatus := SetPendingBankPayment(log, e, data.PlayerId, data.SessionId, 50, "jail release")
            if pendingBankPaymentStatus.Status != http.StatusOK {
                return pendingBankPaymentStatus
            }

            return internal.UserActionStatus{
                Status: http.StatusOK,
                Data:   pendingBankPaymentStatus.Data,
            }
        }

        pendingBankPaymentStatus := SetPendingBankPayment(log, e, data.PlayerId, data.SessionId, 50, "jail release")
        if pendingBankPaymentStatus.Status != http.StatusOK {
            return pendingBankPaymentStatus
        }

        return internal.UserActionStatus{
            Status: http.StatusOK,
            Data:   pendingBankPaymentStatus.Data,
        }
    }

    if player.GetOutOfJailCards < 1 {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player does not have a get out of jail card",
        }
    }

    if diceRoll.DieOne == diceRoll.DieTwo {
        return internal.UserActionStatus{
            Status: http.StatusBadRequest,
            Msg:    "player rolled doubles and does not need to use a get out of jail card",
        }
    }

    err := internaldb_players.UpdatePlayerJailState(
        log,
        ctx,
        tx,
        data.PlayerId,
        data.SessionId,
        player.GetOutOfJailCards-1,
        0,
    )
    if err != nil {
        return internal.UserActionStatus{
            Status: http.StatusInternalServerError,
            Msg:    err.Error(),
        }
    }

    delete(e.PendingRolls, data.PlayerId)
    e.TurnHasRolled[data.PlayerId] = false
    e.ExtraRollAllowed[data.PlayerId] = false
    e.DoubleRollCounts[data.PlayerId] = 0
    jailRelease.GetOutOfJailCards = player.GetOutOfJailCards - 1

    e.Broker.Broadcast(log, "UseGetOutOfJailCardEvent", jailRelease)
    events.EmitGameBoardUpdate(log, ctx, e, tx)

    return internal.UserActionStatus{
        Status: http.StatusOK,
        Data:   jailRelease,
    }
}
