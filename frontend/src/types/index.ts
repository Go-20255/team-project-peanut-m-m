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
}

export interface GameBoardData {
  tiles: Tile[]
  current_turn: number
  players: Player[]
}

export interface GameState {
  current_turn: number
  tiles: Tile[]
  players: PlayerInfo[]
  current_roll?: DiceRoll | null
  last_move?: PlayerMovement | null
  pending_property_purchase?: PropertyPurchaseAvailable | null
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
