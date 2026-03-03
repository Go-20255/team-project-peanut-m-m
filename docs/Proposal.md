# Project Proposal: PeaNut M&M
## Overall Goal
Our aim is to recreate Monopoly in Go with a multiplayer implementation. Basic Monopoly features will be implemented, from traversing the board to win and loss conditions.

## Use Cases
The primary use case is to be able to play Monopoly with multiple players on different devices. Our goal is to implement Monopoly mechanics such that they work among all players, such as tracking player statistics throughout the game.

## Components
Go Packages
- https://pkg.go.dev/net/http
- https://pkg.go.dev/github.com/gorilla/websocket
- https://pkg.go.dev/github.com/rs/zerolog
- https://pkg.go.dev/github.com/labstack/echo/v5
- https://pkg.go.dev/github.com/natefinch/lumberjack/v3
- https://pkg.go.dev/github.com/jackc/pgx

Component Architecture:
![Monopoly Components](<monopoly components.png>)

## Testing
- Tests for game logic
- API tests
- Integration tests (?)

## MVP/Stretch Goals

MVP
- Players can join the same game lobby
- A game board is displayed (CLI first)
- Turns rotate amongst the players
- Basic game logic is implemented (dice roll, board - movement, property purchasing)

Stretch Goals
- Implement another simple game (with game menu for game selection, like Jackbox)
- Stats tracking for players between games (requires account creation)
- Game share links
- Authentication

## Checkpoint Functionality
- A working singleplayer monopoly prototype.
Works through CLI
- A multiplayer lobby that players can join and initialize the game in. Does not need to be fully functional between players yet, but allowing players to join the same lobby is expected.
