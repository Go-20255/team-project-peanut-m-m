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
    Status int
    Msg    string
    Data   any
}

type MonopolyEngine struct {
    SessionId         string
    Broker            *handlers.SseBroker
    UserActionsChan   chan UserActionEvent
    UserActionsChanMu sync.Mutex
    PendingRolls      map[int]DiceRoll
    JoinCode          int `json:"join_code"`
    TurnNumber        int `json:"turn_number"`
}

type GameStateUpdate struct {
    CurrentTurn     int     `json:"current_turn"`
}

type GameBoardData struct {
    Tiles       []Tile                  `json:"tiles"`
    Players     []PlayerInfoUpdate      `json:"players"`
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
}

type PlayerInfoUpdate struct {
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
    Id            int          `json:"id"`
    OwnerPlayerId int          `json:"owner_player_id"`
    SessionId     int          `json:"session_id"`
    IsMortgaged   bool         `json:"is_mortgaged"`
    Houses        int          `json:"houses"`
    HasHotel      bool         `json:"has_hotel"`
    PropertyInfo  PropertyData `json:"property_info"`
}

type PropertyData struct {
    Id             int    `json:"id"`
    Name           string `json:"name"`
    CurrentRent    int    `json:"current_rent"`
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
