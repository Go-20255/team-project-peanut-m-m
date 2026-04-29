// Token IDs and their corresponding icon information
export const TOKEN_ICONS = {
  0: { name: "Paw", icon: "paw.svg" },
  1: { name: "Brick", icon: "brick.svg" },
  2: { name: "Paw", icon: "paw.svg" },
  3: { name: "Brick", icon: "brick.svg" },
} as const;

export function getTokenIcon(pieceToken: number): string {
  const token = pieceToken ?? 0;
  const icon = TOKEN_ICONS[token as keyof typeof TOKEN_ICONS];
  if (icon) {
    return `/assets/img/icons/${icon.icon}`;
  } else {
    console.error(`Invalid piece token: ${token}`);
    return "";
  }
}
