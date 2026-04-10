import { api } from '@/shared/lib'
import type { Club, ClubMember, CreateClubRequest } from '@/entities/club'

export const getClubs = async (): Promise<Club[]> => {
  const response = await api.get('/clubs')
  return response.data
}

export const getClub = async (id: number): Promise<Club> => {
  const response = await api.get(`/clubs/${id}`)
  return response.data
}

export const createClub = async (payload: CreateClubRequest): Promise<{ id: number }> => {
  const response = await api.post('/clubs', payload)
  return response.data
}

export const getClubMembers = async (clubId: number): Promise<ClubMember[]> => {
  const response = await api.get(`/clubs/${clubId}/members`)
  return response.data.members
}

export const approveMember = async (clubId: number, memberId: number): Promise<void> => {
  await api.post(`/clubs/${clubId}/members/approve`, { member_id: memberId })
}

export const rejectMember = async (clubId: number, memberId: number): Promise<void> => {
  await api.post(`/clubs/${clubId}/members/reject`, { member_id: memberId })
}
