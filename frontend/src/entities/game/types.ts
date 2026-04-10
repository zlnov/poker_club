export type GameType = 'cash_time' | 'cash_open' | 'tournament';
export type MoneyModel = 'real' | 'chip';
export type EventType = 'buy_in' | 'rebuy' | 'chips_set' | 'correction';

export interface Game {
  id: number;
  club_id: number;
  banker_id: number;
  type: GameType;
  money_model: MoneyModel;
  buy_in_amount: number;
  rebuy_allowed: boolean;
  rebuy_amount: number | null;
  max_rebuys_per_player: number | null;
  duration: string | null;
  start_time: string;
  end_time: string | null;
  min_players: number;
  max_players: number;
  ranking_primary: string;
  ranking_secondary: string | null;
  created_at: string;
  updated_at: string;
}

export interface GameParticipant {
  id: number;
  player_id: number;
  player_name: string;
  buy_in_count: number;
  rebuy_count: number;
  chips_end: number;
  place: number | null;
}

export interface GameWithParticipants extends Game {
  participants: GameParticipant[];
}

export interface GameListItem {
  id: number;
  club_id: number;
  type: GameType;
  money_model: MoneyModel;
  buy_in_amount: number;
  start_time: string;
  end_time: string | null;
  participants_count: number;
}

export interface CreateGameRequest {
  club_id: number;
  banker_id: number;
  type: GameType;
  money_model: MoneyModel;
  buy_in_amount: number;
  rebuy_allowed: boolean;
  rebuy_amount: number | null;
  max_rebuys_per_player: number | null;
  duration: string | null;
  start_time: string;
  min_players: number;
  max_players: number;
  ranking_primary: string;
  ranking_secondary: string | null;
}

export interface BuyInRequest {
  game_id: number;
  player_id: number;
}

export interface SetChipsRequest {
  game_id: number;
  player_id: number;
  chips: number;
}

export interface FinishGameRequest {
  game_id: number;
}

export interface BuyInResponse {
  game_id: number;
  player_id: number;
  buy_in_count: number;
  rebuy_count: number;
  chips_end: number;
}

export interface FinishGameResponse {
  game_id: number;
  end_time: string;
  total_invested: number;
  total_chips: number;
}

export interface GetGameDetailsResponse {
  game: Game;
  participants: GameParticipant[];
}

export interface ListGamesResponse {
  games: GameListItem[];
  total: number;
  limit: number;
  offset: number;
}

export interface ListGamesRequest {
  club_id: number;
  status: 'active' | 'finished' | '';
  limit: number;
  offset: number;
}
