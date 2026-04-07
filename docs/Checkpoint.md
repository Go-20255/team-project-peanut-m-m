# Project Checkpoint: PeaNut M&M

## Summary of what has been completed

- Dockerized PostgreSQL database for game session persistence and storage of game data
- Base monopoly engine which the APIs interface with to perform game actions (movement, property purchases, pull cards, etc)
- Able to host multiple monopoly sessions/games
- API has the following available routes:
    - `GET /api/health`: provides a healthcheck
    - `POST /api/player`: Creates a player for a specified game session
    - `POST /api/game`: Creates a new game and returns a session id that someone can connect to
    - `GET /api/game`: Gets a list of all available game session ids
    - `POST /api/game/join`: API to get session id via a join code
    - `GET /api/game/join/live`: API to connect to live updating Server Sent Events stream (SSE). Allows for live updates to all clients of current state of monopoly game/session
    - `GET /api/game/property`: Checks who owns a property
    - `POST /api/game/property`: Attempt to purchase a property
- Basic website frontend setup to display a game board
- Proper logging of backend where logs are saved to a file
- Simple `justfile` ([just](https://github.com/casey/just)) that allows for easy use of scripts if you don't want to do things via bash scripts


## What still needs to be done

- Updated readme explaining how to start the servers and explains the project
- Rest of basic APIs to control monopoly game via frontend
- Frontend needs to be able to display game board and have user flows of connecting and playing




