"use client"

export function emitToast(message: string) {
  if (typeof window === "undefined") return

  window.dispatchEvent(
    new CustomEvent("game-toast", {
      detail: { message },
    }),
  )
}
