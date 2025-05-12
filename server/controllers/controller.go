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
	}, nil
}

func (c *Controller) Routes(service *echo.Echo) {
	branchG := service.Group("/branch")
	{
		branchG.GET("/", c.BranchList)
		branchG.GET("/:branch_id", c.BranchGetByID)
		branchG.POST("/organization/:organization_id", c.BranchCreate)
		branchG.PUT("/:branch_id/organization/:organization_id", c.BranchUpdate)
		branchG.DELETE("/:branch_id/organization/:organization_id", c.BranchDelete)
		branchG.GET("/:branch/organization/:organization_id", c.BranchOrganizations)
	}

	categoryG := service.Group("/category")
	{
		categoryG.GET("/", c.CategoryList)
		categoryG.GET("/:category_id", c.CategoryGetByID)
		categoryG.POST("/", c.CategoryCreate)
		categoryG.PUT("/:category_id", c.CategoryUpdate)
		categoryG.DELETE("/:category_id", c.CategoryDelete)
	}

	contactUsG := service.Group("/contact-us")
	{
		contactUsG.GET("/", c.ContactUsList)
		contactUsG.GET("/:contact_us_id", c.ContactUsGetByID)
		contactUsG.POST("/", c.ContactUsCreate)
		contactUsG.PUT("/:contact_us_id", c.ContactUsUpdate)
		contactUsG.DELETE("/:contact_us_id", c.ContactUsDelete)
	}

	feedbackG := service.Group("/feedback")
	{
		feedbackG.GET("/", c.FeedbackList)
		feedbackG.GET("/:feedback_id", c.FeedbackGetByID)
		feedbackG.POST("/", c.FeedbackCreate)
		feedbackG.PUT("/:feedback_id", c.FeedbackUpdate)
		feedbackG.DELETE("/:feedback_id", c.FeedbackDelete)
	}

	footstepG := service.Group("/footstep")
	{
		footstepG.GET("/", c.FootstepList)
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
		generatedReportG.GET("/", c.GeneratedReportList)
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
		invitationCodeG.GET("/", c.InvitationCode)
		invitationCodeG.GET("/:invitation_code_id", c.InvitationCodeGetByID)
		invitationCodeG.POST("/organization/:organization_id/branch/:branch_id", c.InvitationCodeCreate)
		invitationCodeG.PUT("/:invitation_code_id/organization/:organization_id/branch/:branch_id", c.InvitationCodeUpdate)
		invitationCodeG.DELETE("/:invitation_code_id/organization/:organization_id/branch/:branch_id", c.InvitationCodeDelete)
		invitationCodeG.GET("/branch/:branch_id", c.InvitationCodeListByBranch)
		invitationCodeG.GET("/organization/:organization_id", c.InvitationCodeListByOrganization)
		invitationCodeG.GET("/exists/:code", c.InvitationCodeListByOrganizationBranch)
		invitationCodeG.GET("/code/:code", c.InvitationCodeExists)
		invitationCodeG.GET("/verfiy/:code", c.InvitationCodeVerify)
	}

	mediaG := service.Group("/media")
	{
		mediaG.GET("/", c.MediaList)
		mediaG.GET("/:media_id", c.MediaGetByID)
		mediaG.POST("/", c.MediaCreate)
		mediaG.PUT("/:media_id", c.MediaUpdate)
		mediaG.DELETE("/:media_id", c.MediaDelete)
	}

	notificationG := service.Group("notification")
	{
		notificationG.GET("/", c.NotificationList)
		notificationG.GET("/:notification_id", c.NotificationGetByID)
		notificationG.DELETE("/:notification_id", c.NotificationDelete)
		notificationG.GET("/user/:user_id", c.NotificationListByUser)
		notificationG.GET("/user/:user_id/unviewed-count", c.NotificationListByUserUnseenCount)
		notificationG.GET("/user/:user_id/unviewed", c.NotificationListByUserUnviewed)
		notificationG.GET("/user/:user_id/read-all", c.NotificationListByUserReadAll)
	}

	organizationG := service.Group("/organization-category")
	{
		organizationG.GET("/", c.OrganizationCategoryList)
		organizationG.GET("/:organization_category_id", c.OrganizationCategoryGetByID)
		organizationG.POST("/organization/:organization_id", c.OrganizationCategoryCreate)
		organizationG.PUT("/:organization_category_id/organization/:organization_id", c.OrganizationCategoryUpdate)
		organizationG.DELETE("/:organization_category_id", c.OrganizationCategoryDelete)
		organizationG.GET("/category/:category_id", c.OrganizationCategoryListByCategory)
		organizationG.GET("/organizaton/:category_id", c.OrganizationCategoryListByOrganization)
	}

	organizationDailyUsage := service.Group("organization-daily-usage")
	{
		organizationDailyUsage.POST("/", c.OrganizationDailyUsageList)
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

	subscriptionPlanG := service.Group("/subscription-plan")
	{
		subscriptionPlanG.GET("/", c.SubscriptionPlanList)
		subscriptionPlanG.GET("/:subscription_plan_id", c.SubscriptionPlanGetByID)
		subscriptionPlanG.POST("/", c.SubscriptionPlanCreate)
		subscriptionPlanG.PUT("/:subscription_plan_id", c.SubscriptionPlanUpdate)
		subscriptionPlanG.DELETE("/:subscription_plan_id", c.SubscriptionPlanDelete)
	}

}
