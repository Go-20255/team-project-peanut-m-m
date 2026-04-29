#!/bin/bash

# Multi-Player Testing Script for Monopoly
# This script automates creating a game and adding multiple players

API_URL="http://localhost:9876/api"
PLAYER_NAMES=("Alice" "Bob" "Charlie" "Diana")

echo "🎲 Monopoly Multi-Player Test"
echo "=============================="
echo ""

# Step 1: Create a new game
echo "📝 Creating new game session..."
SESSION_RESPONSE=$(curl -s -X POST "$API_URL/game")
SESSION_ID="$SESSION_RESPONSE"

if [ -z "$SESSION_ID" ] || [ ${#SESSION_ID} -lt 5 ]; then
    echo "❌ Failed to create game session"
    echo "Response: $SESSION_RESPONSE"
    exit 1
fi

echo "✅ Game created with Session ID: $SESSION_ID"
echo ""

# Step 2: Create players
echo "👥 Adding players to the game..."
PLAYER_IDS=()

for player_name in "${PLAYER_NAMES[@]}"; do
    echo "  Adding player: $player_name..."
    
    PLAYER_RESPONSE=$(curl -s -X POST "$API_URL/player" \
        -d "player_name=$player_name" \
        -d "session_id=$SESSION_ID")
    
    PLAYER_ID=$(echo "$PLAYER_RESPONSE" | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
    
    if [ -z "$PLAYER_ID" ]; then
        echo "  ❌ Failed to create player $player_name"
        echo "  Response: $PLAYER_RESPONSE"
        continue
    fi
    
    PLAYER_IDS+=("$PLAYER_ID")
    echo "  ✅ $player_name created (ID: $PLAYER_ID)"
done

echo ""
echo "=============================="
echo "🎮 Game Setup Complete!"
echo "=============================="
echo ""
echo "Session ID: $SESSION_ID"
echo "Players Created: ${#PLAYER_IDS[@]} players"
echo ""
echo "To test the game in the UI:"
echo "  1. Open http://localhost:3000 in your browser"
echo "  2. First player: Create Game (or use session: $SESSION_ID)"
echo "  3. Other players: Join with session code"
echo ""
echo "To test directly with API:"
echo "  - Roll dice:  POST $API_URL/game/roll"
echo "  - Move player: POST $API_URL/game/move"
echo "  - Buy property: POST $API_URL/game/property"
echo ""
