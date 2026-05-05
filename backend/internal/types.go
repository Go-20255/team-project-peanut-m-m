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
    TempStore         map[string]any
    PendingRolls      map[int]DiceRoll
    PendingRent       *PendingRent
    PendingCardDraw   *PendingCardDraw
    PendingDrawnCard  *DrawnCard
    PendingPropertyPurchase *PendingPropertyPurchase
    PendingBankPayment *PendingBankPayment
    PendingBankPayout  *PendingBankPayout
    PendingExchange    *PendingPlayerExchange
    PendingTrade       *PendingTrade
    TurnHasRolled      map[int]bool
    ExtraRollAllowed   map[int]bool
    DoubleRollCounts   map[int]int
    JoinCode          int `json:"join_code"`
    TurnNumber        int `json:"turn_number"`
}

type GameStateUpdate struct {
    CurrentTurn     int             `json:"current_turn"`
    Players         []PlayerInfo    `json:"players"`
    ExtraRollPlayerId *int          `json:"extra_roll_player_id"`
    PendingCardDraw *PendingCardDraw `json:"pending_card_draw"`
    DrawnCard       *DrawnCard       `json:"drawn_card"`
    PendingRent     *PendingRent     `json:"pending_rent"`
    PendingPropertyPurchase *PendingPropertyPurchase `json:"pending_property_purchase"`
    PendingBankPayment *PendingBankPayment `json:"pending_bank_payment"`
    PendingBankPayout *PendingBankPayout `json:"pending_bank_payout"`
    PendingExchange *PendingPlayerExchange `json:"pending_exchange"`
    PendingTrade *PendingTrade `json:"pending_trade"`
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

type SimpleActionData struct {
    PlayerId    int     `json:"player_id"`
    SessionId   string  `json:"session_id"`
}

type JailReleaseActionData struct {
    PlayerId    int    `json:"player_id"`
    SessionId   string `json:"session_id"`
    Method      string `json:"method"`
}

type BankPaymentActionData struct {
    PlayerId    int    `json:"player_id"`
    SessionId   string `json:"session_id"`
}

type BankPayoutActionData struct {
    PlayerId    int    `json:"player_id"`
    SessionId   string `json:"session_id"`
    Amount      int    `json:"amount"`
    Reason      string `json:"reason"`
}

type CardActionData struct {
    PlayerId    int    `json:"player_id"`
    SessionId   string `json:"session_id"`
}

type RentPaymentActionData struct {
    FromPlayerId int
    ToPlayerId   int
    SessionId    string
    Amount       int
}

type PlayerExchangeActionData struct {
	PlayerId  int    `json:"player_id"`
	SessionId string `json:"session_id"`
}

type PropertyActionData struct {
    PlayerId    int     `json:"player_id"`
    SessionId   string  `json:"session_id"`
    PropertyId int
}

type TradeActionData struct {
    PlayerId            int   `json:"player_id"`
    SessionId           string `json:"session_id"`
    WithPlayerId        int   `json:"with_player_id"`
    OfferedMoney        int   `json:"offered_money"`
    RequestedMoney      int   `json:"requested_money"`
    OfferedPropertyIds  []int `json:"offered_property_ids"`
    RequestedPropertyIds []int `json:"requested_property_ids"`
}

type TradeDecisionActionData struct {
    PlayerId   int    `json:"player_id"`
    SessionId  string `json:"session_id"`
}

type DiceRoll struct {
    PlayerId    int     `json:"player_id"`
    SessionId   string  `json:"session_id"`
    DieOne    int    `json:"die_one"`
    DieTwo    int    `json:"die_two"`
    Total     int    `json:"total"`
    IsDouble  bool   `json:"is_double"`
    RollAgain bool   `json:"roll_again"`
    ReleasedFromJail bool `json:"released_from_jail"`
    SentToJail bool `json:"sent_to_jail"`
    Jailed int `json:"jailed"`
}

type PlayerMovement struct {
    PlayerId    int     `json:"player_id"`
    SessionId   string  `json:"session_id"`
    OldPosition int    `json:"old_position"`
    NewPosition int    `json:"new_position"`
    Total       int    `json:"total"`
    PassedGo    bool   `json:"passed_go"`
    FromCard    bool   `json:"from_card"`
    TurnNumber  int    `json:"turn_number"`
    RentDue     bool   `json:"rent_due"`
    RentAmount  int    `json:"rent_amount"`
    RentToId    int    `json:"rent_to_id"`
    PropertyId  int    `json:"property_id"`
    RollAgain   bool   `json:"roll_again"`
}

type PropertyPurchaseAvailable struct {
    PlayerId     int    `json:"player_id"`
    SessionId    string `json:"session_id"`
    PropertyId   int    `json:"property_id"`
    PurchaseCost int    `json:"purchase_cost"`
    PlayerMoney  int    `json:"player_money"`
    CanAfford    bool   `json:"can_afford"`
}

type PendingPropertyPurchase struct {
    PlayerId     int    `json:"player_id"`
    SessionId    string `json:"session_id"`
    PropertyId   int    `json:"property_id"`
    PurchaseCost int    `json:"purchase_cost"`
    PlayerMoney  int    `json:"player_money"`
    CanAfford    bool   `json:"can_afford"`
}

type PendingCardDraw struct {
    PlayerId    int    `json:"player_id"`
    SessionId   string `json:"session_id"`
    CardType    string `json:"card_type"`
    TileName    string `json:"tile_name"`
    Position    int    `json:"position"`
    DiceTotal   int    `json:"dice_total"`
}

type DrawnCard struct {
    PlayerId     int    `json:"player_id"`
    SessionId    string `json:"session_id"`
    DiceTotal    int    `json:"dice_total"`
    EventCard
}

type PendingRent struct {
    FromPlayerId  int    `json:"from_player_id"`
    ToPlayerId    int    `json:"to_player_id"`
    SessionId     string `json:"session_id"`
    PropertyId    int    `json:"property_id"`
    Position      int    `json:"position"`
    Amount        int    `json:"amount"`
    DiceTotal     int    `json:"dice_total"`
    IsUtilityCard bool   `json:"is_utility_card"`
    IsRailroadCard bool  `json:"is_railroad_card"`
}

type PendingBankPayment struct {
    PlayerId    int    `json:"player_id"`
    SessionId   string `json:"session_id"`
    Amount      int    `json:"amount"`
    Reason      string `json:"reason"`
}

type PendingBankPayout struct {
    PlayerId    int    `json:"player_id"`
    SessionId   string `json:"session_id"`
    Amount      int    `json:"amount"`
    Reason      string `json:"reason"`
}

type DeferredLanding struct {
    PlayerId      int
    SessionId     string
    DiceTotal     int
    PlayerMovement PlayerMovement
}

type PendingPlayerExchange struct {
	ActingPlayerId int    `json:"acting_player_id"`
	SessionId      string `json:"session_id"`
	Amount         int    `json:"amount"`
	Reason         string `json:"reason"`
	IsPayingAll    bool   `json:"is_paying_all"`
}

type TradeProperty struct {
    PropertyId int    `json:"property_id"`
    Name       string `json:"name"`
}

type PendingTrade struct {
    FromPlayerId        int             `json:"from_player_id"`
    ToPlayerId          int             `json:"to_player_id"`
    SessionId           string          `json:"session_id"`
    OfferedMoney        int             `json:"offered_money"`
    RequestedMoney      int             `json:"requested_money"`
    OfferedProperties   []TradeProperty `json:"offered_properties"`
    RequestedProperties []TradeProperty `json:"requested_properties"`
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

type BankPayment struct {
    PlayerId    int    `json:"player_id"`
    SessionId   string `json:"session_id"`
    Amount      int    `json:"amount"`
    Reason      string `json:"reason"`
    PlayerMoney int    `json:"player_money"`
    Jailed      int    `json:"jailed"`
}

type BankPayout struct {
    PlayerId    int    `json:"player_id"`
    SessionId   string `json:"session_id"`
    Amount      int    `json:"amount"`
    Reason      string `json:"reason"`
    PlayerMoney int    `json:"player_money"`
}

type PlayerExchange struct {
	ActingPlayerId int                `json:"acting_player_id"`
	SessionId      string             `json:"session_id"`
	Amount         int                `json:"amount"`
	Reason         string             `json:"reason"`
	IsPayingAll    bool               `json:"is_paying_all"`
	Balances       map[int]int `json:"balances"`
}

type Trade struct {
    FromPlayerId        int             `json:"from_player_id"`
    ToPlayerId          int             `json:"to_player_id"`
    SessionId           string          `json:"session_id"`
    OfferedMoney        int             `json:"offered_money"`
    RequestedMoney      int             `json:"requested_money"`
    OfferedProperties   []TradeProperty `json:"offered_properties"`
    RequestedProperties []TradeProperty `json:"requested_properties"`
    Accepted            bool            `json:"accepted"`
}

type JailRelease struct {
    PlayerId            int    `json:"player_id"`
    SessionId           string `json:"session_id"`
    Method              string `json:"method"`
    GetOutOfJailCards   int    `json:"get_out_of_jail_cards"`
    PlayerMoney         int    `json:"player_money"`
    Jailed              int    `json:"jailed"`
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
    Jailed              int     `json:"jailed"`
    Bankrupt            bool    `json:"bankrupt"`
    Rank                int     `json:"rank"`
    SessionId           string  `json:"session_id"`
    InGame              bool    `json:"in_game"`
}

type Bankruptcy struct {
    PlayerId     int    `json:"player_id"`
    SessionId    string `json:"session_id"`
    Rank         int    `json:"rank"`
    OwesRent     bool   `json:"owes_rent"`
    RentToId     int    `json:"rent_to_id"`
    WinnerId     int    `json:"winner_id"`
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
