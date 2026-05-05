// Token IDs and their corresponding icon information
export const TOKEN_ICONS = {
  0: { name: "Ritchie", icon: "ritchie.png" },
  1: { name: "Roarie", icon: "roarie.png" },
  2: { name: "Ricky the Brick", icon: "brick.png" },
  3: { name: "Go Gopher", icon: "gopher.png" },
} as const

export function getTokenIcon(pieceToken: number): string {
  const token = pieceToken ?? 0
  const icon = TOKEN_ICONS[token as keyof typeof TOKEN_ICONS]
  if (icon) {
    return `/assets/img/icons/${icon.icon}`
  } else {
    console.error(`Invalid piece token: ${token}`)
    return ""
  }
}

export function getTokenName(pieceToken: number): string {
  const token = pieceToken ?? 0
  const icon = TOKEN_ICONS[token as keyof typeof TOKEN_ICONS]
  if (icon) {
    return icon.name
  } else {
    console.error(`Invalid piece token: ${token}`)
    return ""
  }
}
