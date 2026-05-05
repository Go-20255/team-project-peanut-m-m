export interface Player {
  id: number
  name: string
  ready_up_status: boolean
  piece_token: number
  player_order: number
  money: number
  position: number
  get_out_of_jail_cards: number
  jailed: number
  session_id: string
  in_game: boolean
  bankrupt: boolean
  rank: number
}

export interface PlayerInfo {
  player: Player
  owned_properties: OwnedProperty[]
}

export interface PropertyData {
  id: number
  name: string
  rent_id: number
  purchase_cost: number
  mortgage_cost: number
  unmortgage_cost: number
  house_cost: number
  hotel_cost: number
  property_type: string
}

export interface OwnedProperty {
  id: number
  owner_player_id: number
  session_id: number
  current_rent: number
  is_mortgaged: boolean
  houses: number
  has_hotel: boolean
  property_info: PropertyData
}

export interface GameStateUpdate {
  current_turn: number
  players: PlayerInfo[]
  extra_roll_player_id?: number | null
  pending_card_draw?: PendingCardDraw | null
  drawn_card?: DrawnCard | null
  pending_rent?: PendingRent | null
  pending_property_purchase?: PropertyPurchaseAvailable | null
  pending_bank_payment?: PendingBankPayment | null
  pending_bank_payout?: PendingBankPayout | null
  pending_exchange?: PendingPlayerExchange | null
}

export interface GameBoardData {
  tiles: Tile[]
  current_turn: number
  players: PlayerInfo[]
  extra_roll_player_id?: number | null
}

export interface GameState {
  current_turn: number
  tiles: Tile[]
  players: PlayerInfo[]
  extra_roll_player_id?: number | null
  current_roll?: DiceRoll | null
  last_move?: PlayerMovement | null
  pending_card_draw?: PendingCardDraw | null
  drawn_card?: DrawnCard | null
  pending_rent?: PendingRent | null
  pending_property_purchase?: PropertyPurchaseAvailable | null
  pending_bank_payment?: PendingBankPayment | null
  pending_bank_payout?: PendingBankPayout | null
  pending_exchange?: PendingPlayerExchange | null
}

export interface Tile {
  id: number
  name: string
  property_data: PropertyData | null
}

export interface DiceRoll {
  player_id: number
  session_id: string
  die_one: number
  die_two: number
  total: number
  is_double: boolean
  roll_again: boolean
  released_from_jail: boolean
  sent_to_jail: boolean
  jailed: number
}

export interface PlayerMovement {
  player_id: number
  session_id: string
  old_position: number
  new_position: number
  total: number
  passed_go: boolean
  from_card: boolean
  turn_number: number
  rent_due: boolean
  rent_amount: number
  rent_to_id: number
  property_id: number
  roll_again: boolean
}

export interface PropertyPurchaseAvailable {
  player_id: number
  session_id: string
  property_id: number
  purchase_cost: number
  player_money: number
  can_afford: boolean
}

export interface PendingCardDraw {
  player_id: number
  session_id: string
  card_type: string
  tile_name: string
  position: number
  dice_total: number
}

export interface EventCard {
  id: number
  name: string
  description: string
  card_type: string
}

export interface DrawnCard extends EventCard {
  player_id: number
  session_id: string
  dice_total: number
}

export interface PendingBankPayment {
  player_id: number
  session_id: string
  amount: number
  reason: string
}

export interface PendingBankPayout {
  player_id: number
  session_id: string
  amount: number
  reason: string
}

export interface PendingRent {
  from_player_id: number
  to_player_id: number
  session_id: string
  property_id: number
  position: number
  amount: number
  dice_total: number
  is_utility_card: boolean
  is_railroad_card: boolean
}

export interface PendingPlayerExchange {
  acting_player_id: number
  session_id: string
  amount: number
  reason: string
  is_paying_all: boolean
}
