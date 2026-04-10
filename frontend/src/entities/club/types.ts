export interface Club {
  id: number;
  name: string;
  created_at: string;
  updated_at: string;
}

export interface CreateClubRequest {
  name: string;
}

export interface ClubMember {
  id: number;
  club_id: number;
  player_id: number;
  role: 'admin' | 'member';
  status: 'pending' | 'active' | 'banned';
  player: Player;
}

export interface Player {
  id: number;
  first_name: string;
  last_name: string;
  nickname: string;
  phone_number: string;
  created_at: string;
}

export interface ApproveMemberRequest {
  club_id: number;
  member_id: number;
}

export interface RejectMemberRequest {
  club_id: number;
  member_id: number;
}
