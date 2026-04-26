export interface Player {
  id: number;
  name: string;
  player_order: number;
  money: number;
  position: number;
  get_out_of_jail_cards: number;
  jailed: boolean;
  session_id: string;
  in_game: boolean;
  piece_token: number;
}

export interface PropertyData {
  id: number;
  name: string;
  current_rent: number;
  purchase_cost: number;
  mortgage_cost: number;
  unmortgage_cost: number;
  house_cost: number;
  hotel_cost: number;
  property_type: string;
}

export interface OwnedProperty {
  id: number;
  owner_player_id: number;
  session_id: number;
  is_mortgaged: boolean;
  houses: number;
  has_hotel: boolean;
  property_info: PropertyData;
}

export interface GameState {
  players: Player[];
  session_id: string;
  join_code: number;
  turn_number: number;
}

export interface DiceRoll {
  player_id: number;
  session_id: string;
  die_one: number;
  die_two: number;
  total: number;
}

export interface PlayerMovement {
  player_id: number;
  session_id: string;
  old_position: number;
  new_position: number;
  total: number;
  passed_go: boolean;
  turn_number: number;
}
