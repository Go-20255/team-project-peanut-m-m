"use client"

import { Player } from "@/types"
import { getTokenIcon, getTokenName } from "@/utils/tokens"

interface FinalRanksPageProps {
  players: Player[]
}

function getRankLabel(rank: number) {
  const value = Math.abs(rank)
  const mod100 = value % 100

  if (mod100 >= 11 && mod100 <= 13) {
    return `${rank}th place`
  }

  switch (value % 10) {
    case 1:
      return `${rank}st place`
    case 2:
      return `${rank}nd place`
    case 3:
      return `${rank}rd place`
    default:
      return `${rank}th place`
  }
}

export default function FinalRanksPage({ players }: FinalRanksPageProps) {
  const rankedPlayers = players
    .slice()
    .sort((a, b) => {
      if (a.rank === 0 && b.rank === 0) {
        return a.name.localeCompare(b.name)
      }

      if (a.rank === 0) {
        return 1
      }

      if (b.rank === 0) {
        return -1
      }

      return a.rank - b.rank
    })

  return (
    <div
      className="w-full h-screen flex items-center justify-center overflow-hidden"
      style={{ backgroundColor: "#FFFFFF" }}
    >
      <div
        style={{
          width: "min(560px, 90vw)",
          border: "2px solid #D0D3D4",
          backgroundColor: "#FFFFFF",
          padding: 24,
          display: "flex",
          flexDirection: "column",
          gap: 16,
        }}
      >
        <div
          style={{
            color: "#F76902",
            fontSize: 28,
            fontWeight: 700,
            textAlign: "center",
          }}
        >
          Final Ranks
        </div>

        <div
          style={{
            display: "flex",
            flexDirection: "column",
            gap: 10,
          }}
        >
          {rankedPlayers.map((player) => (
            <div
              key={player.id}
              style={{
                border: "2px solid #D0D3D4",
                padding: "12px 14px",
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
                gap: 12,
                backgroundColor: player.rank === 1 ? "#FFF3E0" : "#FFFFFF",
              }}
            >
              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: 10,
                  minWidth: 0,
                }}
              >
                <img
                  src={getTokenIcon(player.piece_token)}
                  alt={getTokenName(player.piece_token)}
                  style={{
                    width: "20px",
                    height: "20px",
                    border: "1px solid #000000",
                    borderRadius: "2px",
                    flexShrink: 0,
                  }}
                />
                <div
                  style={{
                    color: "#000000",
                    fontSize: 16,
                    fontWeight: 700,
                  }}
                >
                  {player.name}
                </div>
              </div>

              <div
                style={{
                  color: player.rank === 1 ? "#F76902" : "#7C878E",
                  fontSize: 14,
                  fontWeight: 700,
                  whiteSpace: "nowrap",
                }}
              >
                {player.rank > 0 ? getRankLabel(player.rank) : "Unranked"}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
