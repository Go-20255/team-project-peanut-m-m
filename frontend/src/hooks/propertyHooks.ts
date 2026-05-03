"use client"

import { useMutation } from "@tanstack/react-query"

const API_URL = process.env.NEXT_PUBLIC_API_URL

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
