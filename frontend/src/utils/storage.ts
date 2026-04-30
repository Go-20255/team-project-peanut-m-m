import { GameState, Player } from "@/types"

export const storage = {
  getItem<T>(key: string): T | null {
    const value = localStorage.getItem(key)
    if (!value) return null

    try {
      return JSON.parse(value) as T
    } catch {
      return value as T
    }
  },

  setItem<T>(key: string, value: T): void {
    localStorage.setItem(key, JSON.stringify(value))
  },

  removeItem(key: string): void {
    localStorage.removeItem(key)
  },

  getSessionId: () => storage.getItem<string>("sessionId"),
  setSessionId: (id: string) => storage.setItem("sessionId", id),

  getPlayer: () => storage.getItem<Player>("player"),
  setPlayer: (p: Player) => storage.setItem("player", p),

  getGameState: () => storage.getItem<GameState>("game_state"),
  setGameState: (gs: GameState) => storage.setItem("game_state", gs),

  getGameCode: () => storage.getItem<string>("gameCode"),
  setGameCode: (code: string) => storage.setItem("gameCode", code),

  clear: () => {
    localStorage.removeItem("sessionId")
    localStorage.removeItem("gameCode")
    localStorage.removeItem("player")
  },
}
