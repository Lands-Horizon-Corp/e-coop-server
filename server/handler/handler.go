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
	database       *horizon.HorizonDatabase
	security       *horizon.HorizonSecurity
}

func NewHandler(
	authentication *horizon.HorizonAuthentication,
	storage *horizon.HorizonStorage,
	provider *provider.Provider,
	repository *repository.Repository,
	model *model.Model,
	publisher *publisher.Publisher,
	database *horizon.HorizonDatabase,
	security *horizon.HorizonSecurity,
) (*Handler, error) {
	return &Handler{
		authentication: authentication,
		storage:        storage,
		provider:       provider,
		repository:     repository,
		model:          model,
		publisher:      publisher,
		database:       database,
		security:       security,
	}, nil
}

func (h *Handler) Routes(service *echo.Echo) {
	categoryGroup := service.Group("/category")
	{
		categoryGroup.GET("", h.CategoryList)
		categoryGroup.GET("/:id", h.CategoryGet)
		categoryGroup.POST("", h.CategoryCreate)
		categoryGroup.PUT("/:id", h.CategoryUpdate)
		categoryGroup.DELETE("/:id", h.CategoryDelete)
	}
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
		footstepGroup.GET("/", h.FootstepList)
		footstepGroup.GET("/:id", h.FootstepGet)
		footstepGroup.DELETE("/:id", h.FootstepDelete)
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
		authenticationGroup.GET("/current", h.UserCurrent)
		authenticationGroup.POST("/login", h.UserLogin)
		authenticationGroup.POST("/logout", h.UserLogout)
		authenticationGroup.POST("/register", h.UserRegister)
		authenticationGroup.POST("/forgot-password", h.UserForgotPassword)
		authenticationGroup.GET("/verify-reset-link/:id", h.UserVerifyResetLink)
		authenticationGroup.POST("/change-password/:id", h.UserChangePassword)
		authenticationGroup.POST("/apply-contact-number", h.UserApplyContactNumber)
		authenticationGroup.POST("/verify-contact-number", h.UserVerifyContactNumber)
		authenticationGroup.POST("/apply-email", h.UserApplyEmail)
		authenticationGroup.POST("/verify-email", h.UserVerifyEmail)
		authenticationGroup.POST("/verify-with-email", h.UserVerifyWithEmail)
		authenticationGroup.POST("/verify-with-email-confirmation", h.UserVerifyWithEmailConfirmation)
		authenticationGroup.POST("/verify-with-contact", h.UserVerifyWithContactNumber)
		authenticationGroup.POST("/verify-with-contact-confirmation", h.UserVerifyWithContactNumberConfirmation)
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

	notificationGroup := service.Group("/notification")
	{
		notificationGroup.GET("", h.NotificationList)
		notificationGroup.GET("/:id", h.NotificationGet)
		notificationGroup.PUT("/:id/view", h.NotificationView)
		notificationGroup.DELETE("/:id", h.NotificationDelete)
		notificationGroup.GET("/unviewed-count", h.NotificationUnviewedCount)
	}
	subscriptionPlanGroup := service.Group("/subscription-plan")
	{
		subscriptionPlanGroup.GET("", h.SubscriptionPlanList)
		subscriptionPlanGroup.GET("/:id", h.SubscriptionPlanGet)
		subscriptionPlanGroup.POST("", h.SubscriptionPlanCreate)
		subscriptionPlanGroup.PUT("/:id", h.SubscriptionPlanUpdate)
		subscriptionPlanGroup.DELETE("/:id", h.SubscriptionPlanDelete)
	}

	organizationGroup := service.Group("organization")
	{
		organizationGroup.GET("", h.OrganizationList)
		organizationGroup.GET("/:id", h.OrganizationGet)
		organizationGroup.POST("", h.OrganizationCreate)
		organizationGroup.PUT("/:id", h.OrganizationUpdate)
		organizationGroup.DELETE("/:id", h.OrganizationDelete)
		organizationGroup.POST("/:id/subscription/:subscription-id", h.OrganizationSubscribe)

	}
}
