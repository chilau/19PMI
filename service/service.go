package service

import (
	"19PMI/19PMI/cmd/api/swaggerui"
	"19PMI/19PMI/config"
	"19PMI/19PMI/logs"
	"19PMI/19PMI/user"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"io/ioutil"
	"net/http"
	"sync"
)

type ErrorResponse struct {
	Message string `json:"error"`
}

type CreateUserResponse struct {
	UserId string `json:"userId"`
}

type GetUserResponse struct {
	User *user.User `json:"user"`
}

type Response struct {
	Message string `json:"message"`
}

type WebService struct {
	logger *logs.Logger
}

func (s *WebService) Run() (err error) {
	ginEngine := gin.New()
	ginEngine.Use(gin.Recovery())

	ginEngine.POST("/users", s.createUser)
	ginEngine.GET("/users/:userId", s.getUser)
	ginEngine.PUT("/users", s.updateUser)
	ginEngine.DELETE("/users/:userId", s.removeUser)

	swaggerui.SwaggerInfo.Title = "Test API"
	swaggerui.SwaggerInfo.Description = "Some description"
	swaggerui.SwaggerInfo.BasePath = ""
	swaggerui.SwaggerInfo.Version = "1.0"
	swaggerui.SwaggerInfo.Schemes = []string{"http", "https"}

	ginEngine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	appConfig := config.GetApplicationConfiguration()

	err = ginEngine.Run(
		fmt.Sprintf(
			":%d",
			appConfig.ServerConfig.Port,
		),
	)

	return
}

func (s *WebService) createUser(ginContext *gin.Context) {
	bodyReader := ginContext.Request.Body
	body, _ := ioutil.ReadAll(bodyReader)

	s.logger.Info().Msg("create user request received")

	var newUser user.User
	err := json.Unmarshal(body, &newUser)
	if err != nil {
		ginContext.JSON(
			http.StatusBadRequest,
			&ErrorResponse{
				Message: "Unexpected JSON format",
			},
		)
	}

	manager := user.GetUsersManager()
	userId, err := manager.Add(&newUser)

	if err != nil {
		ginContext.JSON(
			http.StatusInternalServerError,
			&ErrorResponse{
				Message: fmt.Sprintf(
					"user not created. Error:%s",
					err.Error(),
				),
			},
		)

		return
	}

	ginContext.JSON(
		http.StatusOK,
		&CreateUserResponse{
			UserId: userId,
		},
	)
}

func (s *WebService) getUser(ginContext *gin.Context) {
	var userID = ginContext.Params.ByName("userId")

	manager := user.GetUsersManager()
	requiredUser, err := manager.Get(userID)

	if err != nil {
		ginContext.JSON(
			http.StatusInternalServerError,
			&ErrorResponse{
				Message: err.Error(),
			},
		)

		return
	}

	ginContext.JSON(
		http.StatusOK,
		&GetUserResponse{
			User: requiredUser,
		},
	)
}

func (s *WebService) updateUser(ginContext *gin.Context) {
	bodyReader := ginContext.Request.Body
	body, _ := ioutil.ReadAll(bodyReader)

	s.logger.Info().Msg("update user request received")

	var requiredUser user.User
	err := json.Unmarshal(body, &requiredUser)
	if err != nil {
		ginContext.JSON(
			http.StatusBadRequest,
			&ErrorResponse{
				Message: "Unexpected JSON format",
			},
		)
	}

	manager := user.GetUsersManager()
	ok := manager.Update(&requiredUser)

	if !ok {
		ginContext.JSON(
			http.StatusInternalServerError,
			&Response{
				Message: fmt.Sprintf(
					"user not updated. UserId: %s",
					requiredUser.Id,
				),
			},
		)

		return
	}

	ginContext.JSON(
		http.StatusOK,
		&Response{
			Message: fmt.Sprintf(
				"user updated successfully. UserId: %s",
				requiredUser.Id,
			),
		},
	)
}

func (s *WebService) removeUser(ginContext *gin.Context) {
	var userID = ginContext.Params.ByName("userId")

	manager := user.GetUsersManager()
	ok := manager.Remove(userID)

	if !ok {
		ginContext.JSON(
			http.StatusOK,
			&Response{
				Message: fmt.Sprintf(
					"user not removed. UserId: %s",
					userID,
				),
			},
		)

		return
	}

	ginContext.JSON(
		http.StatusOK,
		&Response{
			Message: fmt.Sprintf(
				"user removed successfully. UserId: %s",
				userID,
			),
		},
	)
}

/**
 * Setup singleton
 */
var mutex = &sync.Mutex{}
var instance *WebService

func GetWebService() *WebService {
	if instance == nil {
		mutex.Lock()
		if instance == nil {
			instance = new(WebService)

			loggerInstance := logs.GetLogger().SetSource("webService")
			instance.logger = &loggerInstance
		}
		mutex.Unlock()
	}

	return instance
}
