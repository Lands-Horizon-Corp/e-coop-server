package controllers

import (
	"github.com/labstack/echo/v4"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
	"horizon.com/server/server/providers"
)

type Controller struct {
	authentication *horizon.HorizonAuthentication
	storage        *horizon.HorizonStorage
	provider       *providers.Providers
	model          *model.Model
	database       *horizon.HorizonDatabase
	security       *horizon.HorizonSecurity

	// all collections
	branch                 *model.BranchCollection
	category               *model.CategoryCollection
	contactUs              *model.ContactUsCollection
	feedback               *model.FeedbackCollection
	footstep               *model.FootstepCollection
	generatedReport        *model.GeneratedReportCollection
	invitationCode         *model.InvitationCodeCollection
	media                  *model.MediaCollection
	notification           *model.NotificationCollection
	organizationCategory   *model.OrganizationCategoryCollection
	organizationDailyUsage *model.OrganizationDailyUsageCollection
	organization           *model.OrganizationCollection
	permissionTemplate     *model.PermissionTemplateCollection
	subscriptionPlan       *model.SubscriptionPlanCollection
	userOrganization       *model.UserOrganizationCollection
	user                   *model.UserCollection
	userRating             *model.UserRatingCollection
}

func NewController(
	authentication *horizon.HorizonAuthentication,
	storage *horizon.HorizonStorage,
	provider *providers.Providers,
	model *model.Model,
	database *horizon.HorizonDatabase,
	security *horizon.HorizonSecurity,

	// all collections
	branch *model.BranchCollection,
	category *model.CategoryCollection,
	contactUs *model.ContactUsCollection,
	feedback *model.FeedbackCollection,
	footstep *model.FootstepCollection,
	generatedReport *model.GeneratedReportCollection,
	invitationCode *model.InvitationCodeCollection,
	media *model.MediaCollection,
	notification *model.NotificationCollection,
	organizationCategory *model.OrganizationCategoryCollection,
	organizationDailyUsage *model.OrganizationDailyUsageCollection,
	organization *model.OrganizationCollection,
	permissionTemplate *model.PermissionTemplateCollection,
	subscriptionPlan *model.SubscriptionPlanCollection,
	userOrganization *model.UserOrganizationCollection,
	user *model.UserCollection,
	userRating *model.UserRatingCollection,

) (*Controller, error) {
	return &Controller{
		authentication:         authentication,
		storage:                storage,
		provider:               provider,
		model:                  model,
		database:               database,
		security:               security,
		branch:                 branch,
		category:               category,
		contactUs:              contactUs,
		feedback:               feedback,
		footstep:               footstep,
		generatedReport:        generatedReport,
		invitationCode:         invitationCode,
		media:                  media,
		notification:           notification,
		organizationCategory:   organizationCategory,
		organizationDailyUsage: organizationDailyUsage,
		organization:           organization,
		permissionTemplate:     permissionTemplate,
		subscriptionPlan:       subscriptionPlan,
		userOrganization:       userOrganization,
		user:                   user,
		userRating:             userRating,
	}, nil
}

func (c *Controller) Routes(service *echo.Echo) {
	branchG := service.Group("/branch")
	{
		branchG.GET("", c.BranchList)
		branchG.GET("/:branch_id", c.BranchGetByID)
		branchG.POST("/organization/:organization_id", c.BranchCreate)
		branchG.PUT("/:branch_id/organization/:organization_id", c.BranchUpdate)
		branchG.DELETE("/:branch_id/organization/:organization_id", c.BranchDelete)
		branchG.GET("/:branch/organization/:organization_id", c.BranchOrganizations)
	}

	categoryG := service.Group("/category")
	{
		categoryG.GET("", c.CategoryList)
		categoryG.GET("/:category_id", c.CategoryGetByID)
		categoryG.POST("", c.CategoryCreate)
		categoryG.PUT("/:category_id", c.CategoryUpdate)
		categoryG.DELETE("/:category_id", c.CategoryDelete)
	}

	contactUsG := service.Group("/contact-us")
	{
		contactUsG.GET("", c.ContactUsList)
		contactUsG.GET("/:contact_us_id", c.ContactUsGetByID)
		contactUsG.POST("", c.ContactUsCreate)
		contactUsG.PUT("/:contact_us_id", c.ContactUsUpdate)
		contactUsG.DELETE("/:contact_us_id", c.ContactUsDelete)
	}

	feedbackG := service.Group("/feedback")
	{
		feedbackG.GET("", c.FeedbackList)
		feedbackG.GET("/:feedback_id", c.FeedbackGetByID)
		feedbackG.POST("", c.FeedbackCreate)
		feedbackG.PUT("/:feedback_id", c.FeedbackUpdate)
		feedbackG.DELETE("/:feedback_id", c.FeedbackDelete)
	}

	footstepG := service.Group("/footstep")
	{
		footstepG.GET("", c.FootstepList)
		footstepG.GET("/:footstep_id", c.FootstepGetByID)
		footstepG.DELETE("/:footstep_id", c.FootstepDelete)
		footstepG.GET("/user/:user_id", c.FootstepListByUser)
		footstepG.GET("/branch/:branch_id", c.FootstepListByBranch)
		footstepG.GET("/organization/:organization_id", c.FootstepListByOrganization)
		footstepG.GET("/organization/:organization_id/branch/:branch_id", c.FootstepListByOrganizationBranch)
		footstepG.GET("/user/:user_id/organization/:organization_id/branch/:branch_id", c.FootstepListByUserOrganizationBranch)
		footstepG.GET("/user/:user_id/branch/:branch_id", c.FootstepUserBranch)
		footstepG.GET("/user/:user_id/organization/:organization_id", c.FootstepListByUserOrganization)
	}

	generatedReportG := service.Group("generated-report")
	{
		generatedReportG.GET("", c.GeneratedReportList)
		generatedReportG.GET("/:generated_report_id", c.GeneratedReportGetByID)
		generatedReportG.DELETE("/:generated_report_id", c.GeneratedReportDelete)
		generatedReportG.GET("/user/:user_id", c.GeneratedReportListByUser)
		generatedReportG.GET("/branch/:branch_id", c.GeneratedReportListByBranch)
		generatedReportG.GET("/organization/:organization_id", c.GeneratedReportListByOrganization)
		generatedReportG.GET("/organization/:organization_id/branch/:branch_id", c.GeneratedReportListByOrganizationBranch)
		generatedReportG.GET("/user/:user_id/organization/:organization_id/branch/:branch_id", c.GeneratedReportListByUserOrganizationBranch)
		generatedReportG.GET("/user/:user_id/branch/:branch_id", c.GeneratedReportUserBranch)
		generatedReportG.GET("/user/:user_id/organization/:organization_id", c.GeneratedReportListByUserOrganization)
	}

	invitationCodeG := service.Group("/invitation-code")
	{
		invitationCodeG.GET("", c.InvitationCode)
		invitationCodeG.GET("/:invitation_code_id", c.InvitationCodeGetByID)
		invitationCodeG.PUT("/:invitation_code_id/", c.InvitationCodeUpdate)
		invitationCodeG.POST("/organization/:organization_id/branch/:branch_id", c.InvitationCodeCreate)
		invitationCodeG.DELETE("/:invitation_code_id/organization/:organization_id/branch/:branch_id", c.InvitationCodeDelete)
		invitationCodeG.GET("/branch/:branch_id", c.InvitationCodeListByBranch)
		invitationCodeG.GET("/organization/:organization_id", c.InvitationCodeListByOrganization)
		invitationCodeG.GET("/exists/:code", c.InvitationCodeListByOrganizationBranch)
		invitationCodeG.GET("/code/:code", c.InvitationCodeExists)
		invitationCodeG.GET("/verfiy/:code", c.InvitationCodeVerify)
	}

	mediaG := service.Group("/media")
	{
		mediaG.GET("", c.MediaList)
		mediaG.GET("/:media_id", c.MediaGetByID)
		mediaG.POST("", c.MediaCreate)
		mediaG.PUT("/:media_id", c.MediaUpdate)
		mediaG.DELETE("/:media_id", c.MediaDelete)
	}

	notificationG := service.Group("notification")
	{
		notificationG.GET("", c.NotificationList)
		notificationG.GET("/:notification_id", c.NotificationGetByID)
		notificationG.DELETE("/:notification_id", c.NotificationDelete)
		notificationG.GET("/user/:user_id", c.NotificationListByUser)
		notificationG.GET("/user/:user_id/unviewed-count", c.NotificationListByUserUnseenCount)
		notificationG.GET("/user/:user_id/unviewed", c.NotificationListByUserUnviewed)
		notificationG.GET("/user/:user_id/read-all", c.NotificationListByUserReadAll)
	}

	organizationCategoryG := service.Group("/organization-category")
	{
		organizationCategoryG.GET("", c.OrganizationCategoryList)
		organizationCategoryG.GET("/:organization_category_id", c.OrganizationCategoryGetByID)
		organizationCategoryG.POST("/organization/:organization_id", c.OrganizationCategoryCreate)
		organizationCategoryG.PUT("/:organization_category_id/organization/:organization_id", c.OrganizationCategoryUpdate)
		organizationCategoryG.DELETE("/:organization_category_id", c.OrganizationCategoryDelete)
		organizationCategoryG.GET("/category/:category_id", c.OrganizationCategoryListByCategory)
		organizationCategoryG.GET("/organizaton/:category_id", c.OrganizationCategoryListByOrganization)
	}

	organizationDailyUsage := service.Group("organization-daily-usage")
	{
		organizationDailyUsage.POST("", c.OrganizationDailyUsageList)
		organizationDailyUsage.GET("/:organization_daily_usage_id", c.OrganizationDailyUsageGetByID)
		organizationDailyUsage.DELETE("/:organization_daily_usage_id", c.OrganizationDailyUsageDelete)
		organizationDailyUsage.GET("/organization/:organization_id", c.OrganizationDailyUsageListByOrganization)
	}

	authenticationG := service.Group("/authentication")
	{
		authenticationG.GET("/current", c.UserCurrent)
		authenticationG.POST("/login", c.UserLogin)
		authenticationG.POST("/logout", c.UserLogout)
		authenticationG.POST("/register", c.UserRegister)
		authenticationG.POST("/forgot-password", c.UserForgotPassword)
		authenticationG.GET("/verify-reset-link/:id", c.UserVerifyResetLink)
		authenticationG.POST("/change-password/:id", c.UserChangePassword)
		authenticationG.POST("/apply-contact-number", c.UserApplyContactNumber)
		authenticationG.POST("/verify-contact-number", c.UserVerifyContactNumber)
		authenticationG.POST("/apply-email", c.UserApplyEmail)
		authenticationG.POST("/verify-email", c.UserVerifyEmail)
		authenticationG.POST("/verify-with-email", c.UserVerifyWithEmail)
		authenticationG.POST("/verify-with-email-confirmation", c.UserVerifyWithEmailConfirmation)
		authenticationG.POST("/verify-with-contact", c.UserVerifyWithContactNumber)
		authenticationG.POST("/verify-with-contact-confirmation", c.UserVerifyWithContactNumberConfirmation)
	}
	profileGroup := service.Group("/profile")
	{
		profileGroup.PUT("/password", c.UserSettingsChangePassword)
		profileGroup.PUT("/email", c.UserSettingsChangeEmail)
		profileGroup.PUT("/username", c.UserSettingsChangeUsername)
		profileGroup.PUT("/contact-number", c.UserSettingsChangeContactNumber)
		profileGroup.PUT("/profile-picture", c.UserSettingsChangeProfilePicture)
		profileGroup.PUT("/profile", c.UserSettingsChangeProfile)
		profileGroup.PUT("/general", c.UserSettingsChangeGeneral)
	}

	subscriptionPlanG := service.Group("/subscription-plan")
	{
		subscriptionPlanG.GET("", c.SubscriptionPlanList)
		subscriptionPlanG.GET("/:subscription_plan_id", c.SubscriptionPlanGetByID)
		subscriptionPlanG.POST("", c.SubscriptionPlanCreate)
		subscriptionPlanG.PUT("/:subscription_plan_id", c.SubscriptionPlanUpdate)
		subscriptionPlanG.DELETE("/:subscription_plan_id", c.SubscriptionPlanDelete)
	}

	userOrganizationG := service.Group("/user-organization")
	{
		userOrganizationG.GET("", c.UserOrganizationGetAll)
		userOrganizationG.GET("/:user_organization_id", c.UserOrganizationGetByID)
		userOrganizationG.PUT("/:user_organization_id", c.UserOrganizationUpdate)
		userOrganizationG.PUT("/:user_organization_id/developer-key-refresh", c.UserOrganizationRegenerateDeveloperKey)
		userOrganizationG.POST("/join/organization/:organization_id/branch/:branch_id", c.UserOrganizationJoin)
		userOrganizationG.POST("/join/invitation-code/:code", c.UserOrganizationJoinByCode)
		userOrganizationG.POST("/leave/organization/:organization_id/branch/:branch_id", c.UserOrganizationLeave)
		userOrganizationG.GET("/user/:user_id", c.UserOrganizationListByUser)
		userOrganizationG.GET("/branch/:branch_id", c.UserOrganizationListByBranch)
		userOrganizationG.GET("/organization/:organization_id", c.UserOrganizationListByOrganization)
		userOrganizationG.GET("/organization/:organization_id/branch/:branch_id", c.UserOrganizationListByOrganizationBranch)
		userOrganizationG.GET("/user/:user_id/branch/:branch_id", c.UserOrganizationListByUserBranch)
		userOrganizationG.GET("/user/:user_id/organization/:organization_id", c.UserOrganizationListByUserOrganization)
		userOrganizationG.GET("/user/:user_id/organization/:organization_id/branch/:branch_id", c.UserOrganizationByUserOrganizationBranch)
		userOrganizationG.GET("/organization/:organization_id/branch/:branch_id/can-join-employee", c.UserOrganizationCanJoinMember)
		userOrganizationG.GET("/organization/:organization_id/branch/:branch_id/can-join-employee", c.UserOrganizationCanJoinEmployee)
	}

	permissionTemplateG := service.Group("/permission-template")
	{
		permissionTemplateG.GET("", c.PermissionTemplateList)
		permissionTemplateG.GET("/:permission_template_id", c.PermissionTemplateGetByID)
		permissionTemplateG.POST("/organization/:organization_id/branch/:branch_id", c.PermissionTemplateCreate)
		permissionTemplateG.PUT("/:permission_template_id", c.PermissionTemplateUpdate)
		permissionTemplateG.DELETE("/:permission_template_id", c.PermissionTemplateDelete)
	}

	organizationG := service.Group("/organization")
	{
		organizationG.GET("", c.OrganizationList)
		organizationG.GET("/:organization_id", c.OrganizationGetByID)
		organizationG.POST("", c.OrganizationCreate)
		organizationG.PUT("/:organization_id", c.OrganizationUpdate)
		organizationG.DELETE("/:organization_id", c.OrganizationDelete)
	}

	userRatingG := service.Group("/user-rating")
	{
		userRatingG.GET("", c.UserRatingList)
		userRatingG.GET("/:rating_id", c.UserRatingGetByID)
		userRatingG.GET("/organization/:organization_id/branch/:branch_id", c.UserRatingCreate)
		userRatingG.DELETE("/:rating_id", c.UserRatingDelete)
		userRatingG.GET("/user-ratee/:user_ratee_id", c.UserRatingListByRatee)
		userRatingG.GET("/user-rater/:user_rater_id", c.UserRatingListByRater)
		userRatingG.GET("/branch/:branch_id", c.UserRatingListByBranch)
		userRatingG.GET("/organization/:organization_id", c.UserRatingListByOrganization)
		userRatingG.GET("/organization/:organization_id/branch/:branch_id", c.UserRatingListByOrganizationBranch)
		userRatingG.GET("/branch/:branch_id/ratee/:ratee_user_id", c.UserRatingListByBranchRatee)
		userRatingG.GET("/branch/:branch_id/rater/:rater_user_id", c.UserRatingListByBranchRater)
		userRatingG.GET("/organization/:organization_id/ratee/:ratee_user_id", c.UserRatingListByOrganizationRatee)
		userRatingG.GET("/organization/:organization_id/rater/:rater_user_id", c.UserRatingListByOrganizationRater)
		userRatingG.GET("/organization/:organization_id/branch/:branch_id/ratee/:ratee_user_id", c.UserRatingListByOrgBranchRatee)
		userRatingG.GET("/organization/:organization_id/branch/:branch_id/rater/:rater_user_id", c.UserRatingListByOrgBranchRater)
	}

}
