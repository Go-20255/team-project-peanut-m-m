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
  players: Player[]
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
  session_id: string
  join_code: number
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
}

export interface PlayerMovement {
  player_id: number
  session_id: string
  old_position: number
  new_position: number
  total: number
  passed_go: boolean
  turn_number: number
}
