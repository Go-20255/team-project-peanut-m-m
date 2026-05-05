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
    mutationFn: async () => {
      const res = await fetch(`${API_URL}/api/game/property`, {
        method: "POST",
        credentials: "include",
      })
      if (!res.ok) {
        const errorText = await res.text()
        throw new Error(errorText || "Failed to purchase property")
      }
      return res.text()
    },
  })
}

export function useIgnorePropertyPurchase() {
  return useMutation({
    mutationFn: async () => {
      const res = await fetch(`${API_URL}/api/game/property/ignore`, {
        method: "POST",
        credentials: "include",
      })
      if (!res.ok) {
        const errorText = await res.text()
        throw new Error(errorText || "Failed to ignore property")
      }
      return res.text()
    },
  })
}

export function usePurchaseHouse() {
  return useMutation({
    mutationFn: async (propertyId: number) => {
      const body = new URLSearchParams({
        property_id: propertyId.toString(),
      })
      const res = await fetch(`${API_URL}/api/game/property/house`, {
        method: "POST",
        body,
        credentials: "include",
      })
      if (!res.ok) {
        const errorText = await res.text()
        throw new Error(errorText || "Failed to buy house")
      }
      return res.json()
    },
  })
}

export function usePurchaseHotel() {
  return useMutation({
    mutationFn: async (propertyId: number) => {
      const body = new URLSearchParams({
        property_id: propertyId.toString(),
      })
      const res = await fetch(`${API_URL}/api/game/property/hotel`, {
        method: "POST",
        body,
        credentials: "include",
      })
      if (!res.ok) {
        const errorText = await res.text()
        throw new Error(errorText || "Failed to buy hotel")
      }
      return res.json()
    },
  })
}

export function useSellHouse() {
  return useMutation({
    mutationFn: async (propertyId: number) => {
      const body = new URLSearchParams({
        property_id: propertyId.toString(),
      })
      const res = await fetch(`${API_URL}/api/game/property/house/sell`, {
        method: "POST",
        body,
        credentials: "include",
      })
      if (!res.ok) {
        const errorText = await res.text()
        throw new Error(errorText || "Failed to sell house")
      }
      return res.json()
    },
  })
}

export function useSellHotel() {
  return useMutation({
    mutationFn: async (propertyId: number) => {
      const body = new URLSearchParams({
        property_id: propertyId.toString(),
      })
      const res = await fetch(`${API_URL}/api/game/property/hotel/sell`, {
        method: "POST",
        body,
        credentials: "include",
      })
      if (!res.ok) {
        const errorText = await res.text()
        throw new Error(errorText || "Failed to sell hotel")
      }
      return res.json()
    },
  })
}

export function useMortgageProperty() {
  return useMutation({
    mutationFn: async (propertyId: number) => {
      const body = new URLSearchParams({
        property_id: propertyId.toString(),
      })
      const res = await fetch(`${API_URL}/api/game/property/mortgage`, {
        method: "POST",
        body,
        credentials: "include",
      })
      if (!res.ok) {
        const errorText = await res.text()
        throw new Error(errorText || "Failed to mortgage property")
      }
      return res.json()
    },
  })
}

export function useUnmortgageProperty() {
  return useMutation({
    mutationFn: async (propertyId: number) => {
      const body = new URLSearchParams({
        property_id: propertyId.toString(),
      })
      const res = await fetch(`${API_URL}/api/game/property/unmortgage`, {
        method: "POST",
        body,
        credentials: "include",
      })
      if (!res.ok) {
        const errorText = await res.text()
        throw new Error(errorText || "Failed to unmortgage property")
      }
      return res.json()
    },
  })
}
