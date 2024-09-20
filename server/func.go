package main

import (
    "github.com/gin-gonic/gin"
    "net/http"
    "context"
    "log"
	"fmt"
	"time"
    "encoding/json"
	"github.com/golang-jwt/jwt/v4"
)

func getClientID(c *gin.Context) {
    clientID := "client-" + fmt.Sprintf("%d", time.Now().UnixNano())
    c.JSON(http.StatusOK, gin.H{"client_id": clientID})
}

func receiveClientData(c *gin.Context) {
    var data ClientData

    if err := c.ShouldBindJSON(&data); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    log.Printf("Received data from client %s\n", data.ClientID)
    log.Printf("Network Interfaces: %+v\n", data.NetworkInterfaces)
    log.Printf("System Info: %+v\n", data.SystemInfo)
    log.Printf("Active Ports: %+v\n", data.ActivePorts)

    c.JSON(http.StatusOK, gin.H{"message": "data received"})
}

// oauth2
func handleGoogleAuth(c *gin.Context) {
	var request struct {
		Code string `json:"code"`
	}

    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request, code not provided"})
        return
    }

    code := request.Code
    if code == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Code not provided"})
        return
    }

    //c.JSON(http.StatusOK, gin.H{"code": code})
    // code := c.PostForm("code") // 프론트엔드에서 전달된 authorization code
    // if code == "" {
    //     c.JSON(http.StatusBadRequest, gin.H{"error": "Code not provided"})
    //     return
    // }

    // Context 생성
    ctx := context.Background()
    token, err := googleOAuthConfig.Exchange(ctx, code)
    if err != nil {
        result := fmt.Sprintf("Failed to exchange token (%s)", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": result})
        return
    }

    // token 안에 포함된 id_token 추출
    idToken, ok := token.Extra("id_token").(string)
    if !ok {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "No id_token found in token response"})
        return
    }

    //fmt.Println("idToken : ", idToken)
    // 얻어온 id_token을 반환하거나 사용
    // c.JSON(http.StatusOK, gin.H{
    //     "access_token": token.AccessToken,
    //     "id_token":     idToken,
    //     "expiry":       token.Expiry,
    // })

    // `id_token`을 사용하여 `tokeninfo` 엔드포인트 호출
    tokenInfoURL := fmt.Sprintf("https://oauth2.googleapis.com/tokeninfo?id_token=%s", idToken)
    tokenInfoResp, err := http.Get(tokenInfoURL)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to get token info"})
        return
    }
    defer tokenInfoResp.Body.Close()

    var profile map[string]interface{}
    if err := json.NewDecoder(tokenInfoResp.Body).Decode(&profile); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to decode token info"})
        return
    }

    // 프로필 정보 추출
    id := profile["sub"].(string)
    name := profile["name"].(string)
    email := profile["email"].(string)
    picture := profile["picture"].(string)

    // 디비에 로그인한 유저 정보 등록
    user := User{
        ID:      id,
        Name:    name,
        Email:   email,
        Picture: picture,
    }
    users[user.ID] = user

    // JWT 토큰 생성
    claims := Claims{
        ID:      id,
        Name:    name,
        Email:   email,
        Picture: picture,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
        },
    }

    jwtToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(jwtSecret)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT"})
        return
    }

    http.SetCookie(c.Writer, &http.Cookie{
        Name:     "jwt",
        Value:    jwtToken,
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteNoneMode,
        Path:     "/",
    })

    c.JSON(http.StatusOK, gin.H{"message": "Logged in", "profile": user})
}

// JWT 미들웨어
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 쿠키에서 JWT 토큰 가져오기
        tokenString, err := c.Cookie("jwt")
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization cookie not found"})
            c.Abort()
            return
        }

        // JWT 토큰 검증
        claims := &Claims{}
        token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
            return jwtSecret, nil
        })
        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
            c.Abort()
            return
        }

        // 유효한 토큰일 경우 요청에 사용자 정보 추가
        c.Set("user", claims)
        c.Next()
    }
}

func protectedEndpoint(c *gin.Context) {
	user, _ := c.Get("user")
	c.JSON(http.StatusOK, gin.H{"message": "Protected endpoint", "user": user})
}


func getProfile(c *gin.Context) {
    // 미들웨어에서 설정한 사용자 정보 가져오기
    user, exists := c.Get("user")
    if !exists {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "User information not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"profile": user})
}

// 로그아웃 핸들러
func logout(c *gin.Context) {
    // 컨텍스트에서 사용자 정보 가져오기
    user, exists := c.Get("user")
    if !exists {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "User information not found"})
        return
    }

    userInfo := user.(*User) // 사용자 정보 캐스팅

    // 사용자가 등록한 정보를 삭제 (예: 데이터베이스 삭제)
    delete(users, userInfo.ID)

    // JWT 블랙리스트에 토큰 추가 (토큰 무효화)
    token, err := c.Cookie("jwt")
    if err == nil {
        jwtBlacklist[token] = time.Now().Add(time.Hour * 24) // 블랙리스트 만료 시간 설정 (예: 24시간)
    }

    // 클라이언트 측에서 JWT 쿠키 삭제
    http.SetCookie(c.Writer, &http.Cookie{
        Name:     "jwt",
        Value:    "",
        Expires:  time.Now().Add(-time.Hour), // 만료시간을 과거로 설정하여 삭제
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteNoneMode,
        Path:     "/",
    })

    c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully and user information deleted"})
}