package internal

import (
    "monopoly-backend/handlers"
    "sync"
)

type UserActionEvent struct {
    // state the event that this action is creating
    Event string
    // provide the data necessary for this event
    Data any
    // return the status of the action to the caller
    ReturnChan chan UserActionStatus
}

type UserActionStatus struct {
    // http status code to return to caller
    Status int
    // String Message to respond with
    Msg    string
    // Data blob to return to caller
    Data   any
}

type MonopolyEngine struct {
    SessionId         string
    Broker            *handlers.SseBroker
    UserActionsChan   chan UserActionEvent
    UserActionsChanMu sync.Mutex
    PendingRolls      map[int]DiceRoll
    PendingRent       *PendingRent
    JoinCode          int `json:"join_code"`
    TurnNumber        int `json:"turn_number"`
}

type GameStateUpdate struct {
    CurrentTurn     int             `json:"current_turn"`
    Players         []PlayerInfo    `json:"players"`
}

type GameBoardData struct {
    Tiles       []Tile                  `json:"tiles"`
    GameStateUpdate
}

type Tile struct {
    Id              int             `json:"id"`
    Name            string          `json:"name"`
    PropertyData    *PropertyData   `json:"property_data"`
}

type RollDiceActionData struct {
    PlayerId  int
    SessionId string
}

type MovePlayerActionData struct {
    PlayerId  int
    SessionId string
}

type RentPaymentActionData struct {
    FromPlayerId int
    ToPlayerId   int
    SessionId    string
    Amount       int
}

type PropertyActionData struct {
    PlayerId   int
    SessionId  string
    PropertyId int
}

type DiceRoll struct {
    PlayerId  int    `json:"player_id"`
    SessionId string `json:"session_id"`
    DieOne    int    `json:"die_one"`
    DieTwo    int    `json:"die_two"`
    Total     int    `json:"total"`
}

type PlayerMovement struct {
    PlayerId    int    `json:"player_id"`
    SessionId   string `json:"session_id"`
    OldPosition int    `json:"old_position"`
    NewPosition int    `json:"new_position"`
    Total       int    `json:"total"`
    PassedGo    bool   `json:"passed_go"`
    TurnNumber  int    `json:"turn_number"`
    RentDue     bool   `json:"rent_due"`
    RentAmount  int    `json:"rent_amount"`
    RentToId    int    `json:"rent_to_id"`
    PropertyId  int    `json:"property_id"`
}

type PendingRent struct {
    FromPlayerId int
    ToPlayerId   int
    SessionId    string
    PropertyId   int
    Position     int
    Amount       int
    DiceTotal    int
}

type RentPayment struct {
    FromPlayerId   int    `json:"from_player_id"`
    ToPlayerId     int    `json:"to_player_id"`
    SessionId      string `json:"session_id"`
    PropertyId     int    `json:"property_id"`
    Amount         int    `json:"amount"`
    FromPlayerMoney int   `json:"from_player_money"`
    ToPlayerMoney  int    `json:"to_player_money"`
}

type PropertyBuildingUpdate struct {
    PlayerId        int  `json:"player_id"`
    SessionId       string `json:"session_id"`
    PropertyId      int  `json:"property_id"`
    Houses          int  `json:"houses"`
    HasHotel        bool `json:"has_hotel"`
    PlayerMoney     int  `json:"player_money"`
    AvailableHouses int  `json:"available_houses"`
    AvailableHotels int  `json:"available_hotels"`
}

type PropertyMortgageUpdate struct {
    PlayerId     int    `json:"player_id"`
    SessionId    string `json:"session_id"`
    PropertyId   int    `json:"property_id"`
    IsMortgaged  bool   `json:"is_mortgaged"`
    PlayerMoney  int    `json:"player_money"`
}

type PlayerInfo struct {
    Player              Player              `json:"player"`
    // properties attached to above player
    OwnedProperties     []OwnedProperty     `json:"owned_properties"`
}

type Player struct {
    Id                  int     `json:"id"`
    Name                string  `json:"name"`
    ReadyUpStatus       bool    `json:"ready_up_status"`
    PieceToken          int     `json:"piece_token"`
    PlayerOrder         int     `json:"player_order"`
    Money               int     `json:"money"`
    Position            int     `json:"position"`
    GetOutOfJailCards   int     `json:"get_out_of_jail_cards"`
    Jailed              bool    `json:"jailed"`
    SessionId           string  `json:"session_id"`
    InGame              bool    `json:"in_game"`
}

type OwnedProperty struct {
    Id              int             `json:"id"`
    OwnerPlayerId   int             `json:"owner_player_id"`
    SessionId       string          `json:"session_id"`
    CurrentRent     int             `json:"current_rent"` // calculated in sql
    IsMortgaged     bool            `json:"is_mortgaged"`
    Houses          int             `json:"houses"`
    HasHotel        bool            `json:"has_hotel"`
    PropertyInfo    PropertyData    `json:"property_info"`
}

type PropertyData struct {
    Id             int    `json:"id"`
    Name           string `json:"name"`
    RentId         int    `json:"rent_id"`
    PurchaseCost   int    `json:"purchase_cost"`
    MortgageCost   int    `json:"mortgage_cost"`
    UnmortgageCost int    `json:"unmortgage_cost"`
    HouseCost      int    `json:"house_cost"`
    HotelCost      int    `json:"hotel_cost"`
    PropertyType   string `json:"property_type"`
}

type EventCard struct {
    Id          int    `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
    CardType    string `json:"card_type"`
}
