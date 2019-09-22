package controllers

import (
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/shootnix/jackie-chat-2/config"
	"github.com/shootnix/jackie-chat-2/constant"
	"github.com/shootnix/jackie-chat-2/entity"
	"github.com/shootnix/jackie-chat-2/logger"
	"net/http"
	"strconv"
	"time"
)

type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type SendMessageReq struct {
	ChatID    int64  `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

func Ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func AuthRequired() gin.HandlerFunc {

	log := logger.GetLogger()
	m := entity.NewMessage()
	m.ChatID = constant.JACKIE_CHAT_DAILY // Jackie Chat Daily
	sender, _ := entity.FindUser("Paolo")

	m.BotID = sender.DefaultBot
	m.UserID = sender.ID

	return func(c *gin.Context) {

		if bearer := c.GetHeader("Authorization"); bearer != "" {

			log.Debug("has authorization header")

			tknStr := bearer[7:len(bearer)]
			claims := &Claims{}

			log.Debug("token = " + tknStr)

			tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
				return config.GetConfig().JWTKey, nil
			})
			if err != nil {

				log.Debug("Got error while checking bearer token: " + err.Error())
				m.Message = "Got someone's wrong bearer token!"
				m.Insert()
				m.Send("telegram")

				if err == jwt.ErrSignatureInvalid {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "Wrong Token"})
					c.Abort()
					return
				}
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				c.Abort()
				return
			}
			if !tkn.Valid {

				log.Debug("Invlalid token")
				m.Message = claims.Username + " has been send invalid auth token!"
				m.Insert()
				m.Send("telegram")

				c.JSON(http.StatusUnauthorized, gin.H{"error": "Wrong Token"})
				c.Abort()
				return
			}
			//c.Set("Username", claims.Username)
			u, err := entity.FindUser(claims.Username)
			if err != nil {
				m.Message = "Can't find user with username " + claims.Username
				m.Insert()
				m.Send("telegram")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
				c.Abort()
				return
			}

			log.Debug("Found user: " + u.Name)

			c.Set("User", u)
		} else {
			m.Message = "Someone trying to reach method without authorization"
			m.Insert()
			m.Send("telegram")

			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization Required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func Auth(c *gin.Context) {

	log := logger.GetLogger()

	m := entity.NewMessage()
	m.ChatID = constant.JACKIE_CHAT_DAILY
	sender, _ := entity.FindUser("Paolo")
	m.BotID = sender.DefaultBot
	m.UserID = sender.ID

	var creds Credentials
	if err := c.ShouldBindJSON(&creds); err != nil {

		log.Debug("[Auth]: Error - " + err.Error())
		m.Message = "Trying to fire up login: " + err.Error()
		m.Insert()
		m.Send("telegram")

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := entity.LoginUser(creds.Username, creds.Password)
	if err != nil {
		m.Message = "Trying to log in user " + creds.Username + ": " + err.Error()
		m.Insert()
		m.Send("telegram")

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Wrong Credentials"})
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		Username: creds.Username,
		StandardClaims: jwt.StandardClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Create the JWT string
	tokenString, err := token.SignedString(config.GetConfig().JWTKey)
	if err != nil {
		m.Message = "error while trying to create token string: " + err.Error()
		m.Insert()
		m.Send("telegram")

		log.Debug("Error = " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unknown Error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func SendMessage(c *gin.Context) {

	log := logger.GetLogger()

	var req SendMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := c.MustGet("User").(*entity.User)
	// Validation
	if req.ChatID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chat_id required"})
		return
	}

	if req.Text == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "text required"})
		return
	}

	parseMode := req.ParseMode
	if parseMode == "" {
		parseMode = "html"
	}

	m := entity.NewMessage()
	m.Message = req.Text
	m.ChatID = req.ChatID
	m.ParseMode = parseMode
	m.BotID = user.DefaultBot
	m.UserID = user.ID
	if err := m.Insert(); err != nil {
		log.Error("Can't insert data into the database: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Server Error"})
		return
	}

	log.Debug("User " + user.Name + " is about to send a message #" + strconv.FormatInt(m.ID, 10))
	_ = m.Send("Telegram")

	c.JSON(http.StatusCreated, gin.H{"id": m.ID})
}

func GetMessageStatus(c *gin.Context) {
	log := logger.GetLogger()

	idStr := c.Query("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id required"})
		return
	}

	ID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	m, err := entity.GetMessage(ID)
	if err != nil {
		log.Debug(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}

	user := c.MustGet("User").(*entity.User)
	if err != nil {
		log.Debug(err.Error())
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	if user.ID != m.UserID {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	status := "NOT DEFINED"
	switch m.IsSuccess {
	case 1:
		status = "SUCCESS"
	case -1:
		status = "PENDING"
	case 0:
		status = "FAIL"
	}

	c.JSON(http.StatusOK, gin.H{"message": status})
}
