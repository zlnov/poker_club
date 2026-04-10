import { Player } from '../club/types';

export interface PlayerStats {
  player_id: number;
  total_games: number;
  total_buy_in: number;
  total_rebuy: number;
  total_invested: number;
  total_chips: number;
  profit: number;
  roi: number;
  itm: number;
}

export interface PlayerWithStats extends Player {
  stats: PlayerStats;
}

// Leaderboard types based on backend DTO
export interface LeaderboardEntry {
  player_id: number;
  player_name: string;
  metric_value: number;
  games_count: number;
}

export interface LeaderboardResponse {
  metric: string;
  period: string;
  entries: LeaderboardEntry[];
  club_id: number;
  club_name: string;
}

export interface GetLeaderboardRequest {
  club_id: number;
  metric: 'profit' | 'roi' | 'winrate';
  period: 'all';
}
