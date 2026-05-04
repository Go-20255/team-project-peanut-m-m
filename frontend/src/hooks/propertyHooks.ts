"use client"

import { useMutation, useQuery } from "@tanstack/react-query"

const API_URL = process.env.NEXT_PUBLIC_API_URL

export function useFetchPropertyOwner(property_id: string) {
  return useQuery<{ owner_id: number; owned: boolean }>({
    queryKey: ["fetchCheckPropertyOwner", property_id],
    queryFn: async () => {
      const res = await fetch(
        `${API_URL}/api/game/property?id=${property_id}`,
        {
          method: "GET",
          credentials: "include",
        },
      );
      if (!res.ok) {
        throw new Error(res.statusText);
      }
      return res.json();
    },
  });
}

/**
 * Purchase property
 */
export function usePurchaseProperty() {
  return useMutation({
    mutationFn: async ({
      playerId,
      sessionId,
      propertyId,
    }: {
      playerId: number
      sessionId: string
      propertyId: number
    }) => {
      const formData = new FormData()
      formData.append("player_id", playerId.toString())
      formData.append("session_id", sessionId)
      formData.append("property_id", propertyId.toString())

      const res = await fetch(`${API_URL}/api/game/property`, {
        method: "POST",
        credentials: "include",
        body: formData,
      })
      if (!res.ok) {
        throw new Error("Failed to purchase property")
      }
      return res.json()
    },
  })
}

export function usePurchaseHouse() {
  return useMutation({
    mutationFn: async () => {
      const res = await fetch(`${API_URL}/api/game/property/house`, {
        method: "POST",
        credentials: "include",
      })
      if (!res.ok) {
        throw new Error(res.statusText)
      }
      return res.json()
    },
  })
}

export function usePurchaseHotel() {
  return useMutation({
    mutationFn: async () => {
      const res = await fetch(`${API_URL}/api/game/property/hotel`, {
        method: "POST",
        credentials: "include",
      })
      if (!res.ok) {
        throw new Error(res.statusText)
      }
      return res.json()
    },
  })
}

export function useSellHouse() {
  return useMutation({
    mutationFn: async () => {
      const res = await fetch(`${API_URL}/api/game/property/house/sell`, {
        method: "POST",
        credentials: "include",
      })
      if (!res.ok) {
        throw new Error(res.statusText)
      }
      return res.json()
    },
  })
}

export function useSellHotel() {
  return useMutation({
    mutationFn: async () => {
      const res = await fetch(`${API_URL}/api/game/property/hotel/sell`, {
        method: "POST",
        credentials: "include",
      })
      if (!res.ok) {
        throw new Error(res.statusText)
      }
      return res.json()
    },
  })
}

export function useMortgageProperty() {
  return useMutation({
    mutationFn: async () => {
      const res = await fetch(`${API_URL}/api/game/mortgage`, {
        method: "POST",
        credentials: "include",
      })
      if (!res.ok) {
        throw new Error(res.statusText)
      }
      return res.json()
    },
  })
}

export function useUnmortgageProperty() {
  return useMutation({
    mutationFn: async () => {
      const res = await fetch(`${API_URL}/api/game/unmortgage`, {
        method: "POST",
        credentials: "include",
      })
      if (!res.ok) {
        throw new Error(res.statusText)
      }
      return res.json()
    },
  })
}

