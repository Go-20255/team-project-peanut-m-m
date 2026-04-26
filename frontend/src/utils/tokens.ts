// Token IDs and their corresponding icon information
export const TOKEN_ICONS = {
  0: { name: "Paw", icon: "paw.svg" },
  1: { name: "Brick", icon: "brick.svg" },
  2: { name: "Paw", icon: "paw.svg" },
  3: { name: "Brick", icon: "brick.svg" },
  4: { name: "Paw", icon: "paw.svg" },
  5: { name: "Brick", icon: "brick.svg" },
  6: { name: "Paw", icon: "paw.svg" },
  7: { name: "Brick", icon: "brick.svg" },
} as const;

export function getTokenIcon(pieceToken: number): string {
  const token = pieceToken ?? 0;
  const icon = TOKEN_ICONS[token as keyof typeof TOKEN_ICONS];
  console.debug(`getTokenIcon(${pieceToken}) → ${icon?.icon}`);
  return icon ? `/assets/img/icons/${icon.icon}` : "/assets/img/icons/paw.svg";
}

export function getTokenName(pieceToken: number): string {
  const token = pieceToken ?? 0;
  const icon = TOKEN_ICONS[token as keyof typeof TOKEN_ICONS];
  return icon ? icon.name : "Unknown Token";
}
