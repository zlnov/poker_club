import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import * as clubsApi from './api'

export const useClubs = () => {
  return useQuery({
    queryKey: ['clubs'],
    queryFn: clubsApi.getClubs,
  })
}

export const useClub = (clubId: number) => {
  return useQuery({
    queryKey: ['clubs', clubId],
    queryFn: () => clubsApi.getClub(clubId),
    enabled: !!clubId,
  })
}

export const useClubMembers = (clubId: number) => {
  return useQuery({
    queryKey: ['clubs', clubId, 'members'],
    queryFn: () => clubsApi.getClubMembers(clubId),
    enabled: !!clubId,
  })
}

export const useCreateClub = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: clubsApi.createClub,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['clubs'] })
    },
  })
}

export const useApproveMember = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ clubId, memberId }: { clubId: number; memberId: number }) =>
      clubsApi.approveMember(clubId, memberId),
    onSuccess: (_, { clubId }) => {
      queryClient.invalidateQueries({ queryKey: ['clubs', clubId, 'members'] })
    },
  })
}

export const useRejectMember = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ clubId, memberId }: { clubId: number; memberId: number }) =>
      clubsApi.rejectMember(clubId, memberId),
    onSuccess: (_, { clubId }) => {
      queryClient.invalidateQueries({ queryKey: ['clubs', clubId, 'members'] })
    },
  })
}
