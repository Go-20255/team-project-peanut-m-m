package util

import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "strings"
    "time"

    "github.com/labstack/echo/v4"
)

const PlayerAuthCookieName = "player_auth_token"

type PlayerJwtClaims struct {
    PlayerId   int    `json:"player_id"`
    PlayerName string `json:"player_name"`
    SessionId  string `json:"session_id"`
    IssuedAt   int64  `json:"iat"`
    ExpiresAt  int64  `json:"exp"`
}

type playerJwtHeader struct {
    Alg string `json:"alg"`
    Typ string `json:"typ"`
}

// CreatePlayerJwt creates a signed JWT for a player session.
// The token includes the player id, player name, session id, issue time,
// and expiration time.
func CreatePlayerJwt(playerId int, playerName string, sessionId string) (string, error) {
    now := time.Now()
    claims := PlayerJwtClaims{
        PlayerId:   playerId,
        PlayerName: playerName,
        SessionId:  sessionId,
        IssuedAt:   now.Unix(),
        ExpiresAt:  now.Add(24 * time.Hour).Unix(),
    }

    return signPlayerJwt(claims)
}

// GetPlayerJwtClaims extracts the player auth token from the request and
// returns its validated claims.
// It supports reading the token from either the auth cookie or the
// Authorization header.
func GetPlayerJwtClaims(c echo.Context) (PlayerJwtClaims, error) {
    token, err := getPlayerJwtToken(c.Request())
    if err != nil {
        return PlayerJwtClaims{}, err
    }

    return parsePlayerJwt(token)
}

// getPlayerJwtToken retrieves the player auth token from an HTTP request.
// It first checks the player auth cookie, then falls back to a Bearer token
// in the Authorization header.
func getPlayerJwtToken(r *http.Request) (string, error) {
    cookie, err := r.Cookie(PlayerAuthCookieName)
    if err == nil && cookie.Value != "" {
        return cookie.Value, nil
    }

    authHeader := r.Header.Get("Authorization")
    if strings.HasPrefix(authHeader, "Bearer ") {
        token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
        if token != "" {
            return token, nil
        }
    }

    return "", fmt.Errorf("missing auth token")
}

// signPlayerJwt creates and signs a JWT using the provided player claims.
// It builds the token manually using an HS256 signature.
func signPlayerJwt(claims PlayerJwtClaims) (string, error) {
    secret, err := getPlayerJwtSecret()
    if err != nil {
        return "", err
    }

    headerBytes, err := json.Marshal(playerJwtHeader{
        Alg: "HS256",
        Typ: "JWT",
    })
    if err != nil {
        return "", err
    }

    claimsBytes, err := json.Marshal(claims)
    if err != nil {
        return "", err
    }

    encodedHeader := base64.RawURLEncoding.EncodeToString(headerBytes)
    encodedClaims := base64.RawURLEncoding.EncodeToString(claimsBytes)
    signingInput := encodedHeader + "." + encodedClaims

    mac := hmac.New(sha256.New, []byte(secret))
    _, err = mac.Write([]byte(signingInput))
    if err != nil {
        return "", err
    }

    signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
    return signingInput + "." + signature, nil
}

// parsePlayerJwt validates a player JWT and returns its claims.
// It checks the token format, header values, signature, claim contents,
// and expiration time before returning the decoded claims.
func parsePlayerJwt(token string) (PlayerJwtClaims, error) {
    secret, err := getPlayerJwtSecret()
    if err != nil {
        return PlayerJwtClaims{}, err
    }

    parts := strings.Split(token, ".")
    if len(parts) != 3 {
        return PlayerJwtClaims{}, fmt.Errorf("invalid auth token")
    }

    headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
    if err != nil {
        return PlayerJwtClaims{}, fmt.Errorf("invalid auth token")
    }

    var header playerJwtHeader
    err = json.Unmarshal(headerBytes, &header)
    if err != nil {
        return PlayerJwtClaims{}, fmt.Errorf("invalid auth token")
    }

    if header.Alg != "HS256" || header.Typ != "JWT" {
        return PlayerJwtClaims{}, fmt.Errorf("invalid auth token")
    }

    signingInput := parts[0] + "." + parts[1]
    mac := hmac.New(sha256.New, []byte(secret))
    _, err = mac.Write([]byte(signingInput))
    if err != nil {
        return PlayerJwtClaims{}, err
    }

    expectedSignature := mac.Sum(nil)
    signatureBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
    if err != nil {
        return PlayerJwtClaims{}, fmt.Errorf("invalid auth token")
    }

    if !hmac.Equal(signatureBytes, expectedSignature) {
        return PlayerJwtClaims{}, fmt.Errorf("invalid auth token")
    }

    claimsBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
    if err != nil {
        return PlayerJwtClaims{}, fmt.Errorf("invalid auth token")
    }

    var claims PlayerJwtClaims
    err = json.Unmarshal(claimsBytes, &claims)
    if err != nil {
        return PlayerJwtClaims{}, fmt.Errorf("invalid auth token")
    }

    if claims.PlayerId <= 0 || claims.PlayerName == "" || claims.SessionId == "" || claims.IssuedAt <= 0 || claims.ExpiresAt <= claims.IssuedAt {
        return PlayerJwtClaims{}, fmt.Errorf("invalid auth token")
    }

    if time.Now().Unix() >= claims.ExpiresAt {
        return PlayerJwtClaims{}, fmt.Errorf("auth token has expired")
    }

    return claims, nil
}

// getPlayerJwtSecret returns the secret used to sign and verify player JWTs.
// It reads the value from environment configuration and returns an error if
// the secret is missing.
func getPlayerJwtSecret() (string, error) {
    secret := os.Getenv("POSTGRES_PASSWORD")
    if secret == "" {
        return "", fmt.Errorf("missing jwt signing secret")
    }

    return secret, nil
}
