package handler

import (
	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
	"horizon.com/server/server/provider"
	"horizon.com/server/server/publisher"
	"horizon.com/server/server/repository"
)

type Handler struct {
	authentication *horizon.HorizonAuthentication
	storage        *horizon.HorizonStorage
	provider       *provider.Provider
	repository     *repository.Repository
	model          *model.Model
	publisher      *publisher.Publisher
}

func NewHandler(
	authentication *horizon.HorizonAuthentication,
	storage *horizon.HorizonStorage,
	provider *provider.Provider,
	repository *repository.Repository,
	model *model.Model,
	publisher *publisher.Publisher,
) (*Handler, error) {
	return &Handler{
		authentication: authentication,
		storage:        storage,
		provider:       provider,
		repository:     repository,
		model:          model,
		publisher:      publisher,
	}, nil
}

func (h *Handler) Routes(service *echo.Echo) {
	contactGroup := service.Group("/contact-us")
	{
		contactGroup.GET("", h.ContactUsList)
		contactGroup.GET("/:id", h.ContactUsGet)
		contactGroup.POST("", h.ContactUsCreate)
		contactGroup.PUT("/:id", h.ContactUsUpdate)
		contactGroup.DELETE("/:id", h.ContactUsDelete)
	}

	feedbackGroup := service.Group("/feedback")
	{
		feedbackGroup.GET("", h.FeedbackList)
		feedbackGroup.GET("/:id", h.FeedbackGet)
		feedbackGroup.POST("", h.FeedbackCreate)
		feedbackGroup.PUT("/:id", h.FeedbackUpdate)
		feedbackGroup.DELETE("/:id", h.FeedbackDelete)
	}
	footstepGroup := service.Group("/footstep")
	{
		footstepGroup.GET("/footstep", h.FootstepList)
		footstepGroup.GET("/footstep/:id", h.FootstepGet)
		footstepGroup.DELETE("/footstep/:id", h.FootstepDelete)
	}
	mediaGroup := service.Group("/feedback")
	{
		mediaGroup.GET("", h.MediaList)
		mediaGroup.GET("/:id", h.MediaGet)
		mediaGroup.POST("", h.MediaCreate)
		mediaGroup.PUT("/:id", h.MediaUpdate)
		mediaGroup.DELETE("/:id", h.MediaDelete)
	}

	authenticationGroup := service.Group("/authentication")
	{
		authenticationGroup.GET("/authentication/current", h.UserCurrent)
		authenticationGroup.POST("/authentication/login", h.UserLogin)
		authenticationGroup.POST("/authentication/logout", h.UserLogout)
		authenticationGroup.POST("/authentication/register", h.UserRegister)
		authenticationGroup.POST("/authentication/forgot-password", h.UserForgotPassword)
		authenticationGroup.GET("/authentication/verify-reset-link/:id", h.UserVerifyResetLink)
		authenticationGroup.POST("/authentication/change-password/:id", h.UserChangePassword)
		authenticationGroup.POST("/authentication/apply-contact-number", h.UserApplyContactNumber)
		authenticationGroup.POST("/authentication/verify-contact-number", h.UserVerifyContactNumber)
		authenticationGroup.POST("/authentication/apply-email", h.UserApplyEmail)
		authenticationGroup.POST("/authentication/verify-email", h.UserVerifyEmail)
		authenticationGroup.POST("/authentication/verify-with-email", h.UserVerifyWithEmail)
		authenticationGroup.POST("/authentication/verify-with-email-confirmation", h.UserVerifyWithEmailConfirmation)
		authenticationGroup.POST("/authentication/verify-with-contact", h.UserVerifyWithContactNumber)
		authenticationGroup.POST("/authentication/verify-with-contact-confirmation", h.UserVerifyWithContactNumberConfirmation)
	}

	profileGroup := service.Group("/profile")
	{
		profileGroup.PUT("/password", h.UserSettingsChangePassword)
		profileGroup.PUT("/email", h.UserSettingsChangeEmail)
		profileGroup.PUT("/username", h.UserSettingsChangeUsername)
		profileGroup.PUT("/contact-number", h.UserSettingsChangeContactNumber)
		profileGroup.PUT("/profile-picture", h.UserSettingsChangeProfilePicture)
		profileGroup.PUT("/profile", h.UserSettingsChangeProfile)
		profileGroup.PUT("/general", h.UserSettingsChangeGeneral)
	}

}
