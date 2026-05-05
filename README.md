# Monopoly Go

## What is this project?

This project aims to recreate as much of Monopoly as possible in a Go API backend.
This backend is responsible for controlling the state of the game and handling multiple
games simultaneously. The frontend for this project was made with Next.js and Node.js.

## Project Architecture
```
.
├── backend
│   ├── Dockerfile                          <---- Dockerfile for building backend docker image
│   ├── go.mod
│   ├── go.sum
│   ├── handlers                            <---- handlers stores the different API handlers
│   │   ├── common
│   │   │   └── common_handler.go
│   │   ├── game_state
│   │   │   ├── gameplay_handlers.go
│   │   │   ├── game_state_handlers.go
│   │   │   └── game_state_live_handler.go
│   │   ├── player
│   │   │   └── player_handlers.go
│   │   ├── property
│   │   │   └── property_handler.go
│   │   └── sse.go
│   ├── internal
│   │   ├── db                              <---- db stores all postgres queries and functions
│   │   │   ├── common.go
│   │   │   ├── event_cards
│   │   │   │   └── event_cards_db.go
│   │   │   ├── game_state
│   │   │   │   └── game_state_db.go
│   │   │   ├── player
│   │   │   │   └── players_db.go
│   │   │   ├── setup_db.go
│   │   │   └── tile
│   │   │       ├── properties_db.go
│   │   │       ├── rent_db.go
│   │   │       └── tiles_db.go
│   │   ├── engine                          <---- engine stores all monopoly engine related code
│   │   │   ├── events                      <---- events stores how we handle each event produced by the API
│   │   │   │   ├── board.go
│   │   │   │   ├── player
│   │   │   │   │   ├── event_effects.go
│   │   │   │   │   ├── player_events.go
│   │   │   │   │   └── rent_events.go
│   │   │   │   ├── property
│   │   │   │   │   ├── property_building_events.go
│   │   │   │   │   └── property_events.go
│   │   │   │   └── turn
│   │   │   │       └── turn_events.go
│   │   │   └── monopoly_engine.go          <---- here is where the main runtime loop for monopoly game resides
│   │   └── types.go                        <---- here we store almost all type definitions used by the backend
│   ├── main.go                             <---- starts the monopoly engines and API server
│   ├── package-lock.json
│   ├── rebuild_ephemeral_postgres.sh
│   └── util                                <---- utility functions for backend
│       ├── logging.go
│       ├── player_jwt.go
│       └── tokens.go
├── bruno                                   <---- bruno files are stored here for API testing
│   └── <bruno files for api testing>
├── docker-compose.yml                      <---- Docker compose for deployment
├── docs
│   ├── Checkpoint.md
│   ├── monopoly components.png
│   └── Proposal.md
├── frontend
│   ├── Dockerfile                          <---- Dockerfile for building frontend docker image
│   ├── eslint.config.mjs
│   ├── next.config.ts
│   ├── package.json                        <---- frontend package settings
│   ├── package-lock.json
│   ├── postcss.config.mjs
│   ├── public                              <---- all public assets and images are stored here
│   │   ├── assets
│   │   │   └── img
│   │   │       ├── deeds
│   │   │       │   └── <images of deeds>
│   │   │       ├── icons
│   │   │       │   └── <icons>
│   │   │       └── tiles
│   │   │           └── <images of tiles>
│   │   ├── file.svg
│   │   ├── globe.svg
│   │   ├── next.svg
│   │   ├── vercel.svg
│   │   └── window.svg
│   ├── README.md
│   ├── src                                 <---- frontend source code
│   │   ├── app                             <---- / (root route)
│   │   │   ├── game                        <---- /game (game board route)
│   │   │   │   └── page.tsx
│   │   │   ├── globals.css
│   │   │   ├── layout.tsx
│   │   │   ├── page.tsx
│   │   │   ├── ReactQueryProvider.tsx
│   │   │   └── select-player               <---- /select-player (player select route) 
│   │   │       └── page.tsx
│   │   ├── components                      <---- components used in pages stored here
│   │   │   └── game
│   │   │       ├── FinalRanksPage.tsx
│   │   │       ├── GameBoard.tsx
│   │   │       ├── PlayerSidebar.tsx
│   │   │       ├── TokenSelector.tsx
│   │   │       └── TradeOverlay.tsx
│   │   ├── hooks                           <---- All API hooks that call to the backend live here
│   │   │   ├── gameEvents.ts               <---- live update game events are handled here
│   │   │   ├── liveUpdates.ts              <---- manages the live update connection with the backend
│   │   │   ├── playerHooks.ts              <---- hooks for /player routes
│   │   │   ├── propertyHooks.ts            <---- hooks for /property routes
│   │   │   └── useGameAPI.ts               <---- hooks for /game routes
│   │   ├── types
│   │   │   └── index.ts                    <---- shared types for the frontend
│   │   └── utils                           <---- utility functions for the frontend
│   │       ├── api.ts
│   │       ├── index.ts
│   │       ├── storage.ts
│   │       ├── toast.ts
│   │       └── tokens.ts
│   ├── tailwind.config.ts
│   └── tsconfig.json
├── .gitignore
├── justfile                                <---- justfile containing setup scripts
├── README.md                               <---- the file you are reading right now :)
├── setup.sh
└── .vscode
    └── settings.json

39 directories, 170 files
```

## Environment Setup 

### Requirements

Linux environment (either with a VM or WSL2 if on Windows) with the following packages:
- Go >= 1.25.7
- Node >= 24.7.0
- Docker
- [Just](https://github.com/casey/just) >= 1.47.1 (Makefile alternative)

### Setup

After installing the above packages, to setup your environment, run the following commands:
1. `just setup-environment`: creates a .internal.env with passwords; installs node modules in frontend
2. `just redeploy-ephemeral-postgres`: deploys an ephemeral postgres server in docker for data storage
3. Open two terminals
    1. First terminal run `just run-backend`: Starts the backend server
    2. Second terminal run `just run-frontend`: Starts the frontend server (http://localhost:3000)

