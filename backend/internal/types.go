package internal

import (
	"monopoly-backend/handlers"
	"sync"
)


type UserActionEvent struct {
    // state the event that this action is creating
    Event string
    // provide the data necessary for this event
    Data  any
    // return the status of the action to the caller
    ReturnChan chan UserActionStatus
}

type UserActionStatus struct {
    Status  int
    Msg     string
}

type MonopolyEngine struct {
    SessionId           string
    Broker              *handlers.SseBroker
    UserActionsChan     chan UserActionEvent
    UserActionsChanMu   sync.Mutex
    JoinCode        int             `json:"join_code"`
    TurnNumber      int             `json:"turn_number"`
}

type Player struct {
    Id                  int         `json:"id"`
    Name                string      `json:"name"`
    PlayerOrder         int         `json:"player_order"`
    Money               int         `json:"money"`
    Position            int         `json:"position"`
    GetOutOfJailCards   int         `json:"get_out_of_jail_cards"`
    Jailed              bool        `json:"jailed"`
    SessionId           string      `json:"session_id"`
    InGame              bool        `json:"in_game"`
}

type OwnedProperty struct {
    Id                  int         `json:"id"`
    OwnerPlayerId       int         `json:"owner_player_id"`
    SessionId           int         `json:"session_id"`
    IsMortgaged         bool        `json:"is_mortgaged"`
    Houses              int         `json:"houses"`
    HasHotel            bool        `json:"has_hotel"`
    PropertyInfo        PropertyData`json:"property_info"`
}

type PropertyData struct {
    Id                  int         `json:"id"`
    Name                string      `json:"name"`
    CurrentRent         int         `json:"current_rent"`
    PurchaseCost        int         `json:"purchase_cost"`
    MortgageCost        int         `json:"mortgage_cost"`
    UnmortgageCost      int         `json:"unmortgage_cost"`
    HouseCost           int         `json:"house_cost"`
    HotelCost           int         `json:"hotel_cost"`
    PropertyType        string      `json:"property_type"`
}

type EventCard struct {
    Id                  int         `json:"id"`
    Name                string      `json:"name"`
    Description         string      `json:"description"`
    CardType            string      `json:"card_type"`
}

