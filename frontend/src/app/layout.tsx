import type { Metadata } from "next";
import { ReactQueryProvider } from "./ReactQueryProvider";
import "./globals.css";

export const metadata: Metadata = {
  title: "Monopoly",
  description: "Multiplayer Monopoly Game",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>
        <ReactQueryProvider>{children}</ReactQueryProvider>
      </body>
    </html>
  );
}
