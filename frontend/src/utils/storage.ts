export const storage = {
  getSessionId: () => localStorage.getItem("sessionId"),
  setSessionId: (id: string) => localStorage.setItem("sessionId", id),

  getPlayerId: () => localStorage.getItem("playerId"),
  setPlayerId: (id: string) => localStorage.setItem("playerId", id),

  getPlayerName: () => localStorage.getItem("playerName"),
  setPlayerName: (name: string) => localStorage.setItem("playerName", name),

  getGameCode: () => localStorage.getItem("gameCode"),
  setGameCode: (code: string) => localStorage.setItem("gameCode", code),

  clear: () => {
    localStorage.removeItem("sessionId")
    localStorage.removeItem("playerId")
    localStorage.removeItem("playerName")
    localStorage.removeItem("gameCode")
  },
}
