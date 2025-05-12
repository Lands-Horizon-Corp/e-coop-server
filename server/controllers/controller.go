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
		branchG.GET("/branch/branch_id", c.BranchGetByID)
		branchG.POST("organization/:organization_id", c.BranchCreate)
		branchG.PUT("/:branch_id/organization/:organization_id", c.BranchUpdate)
		branchG.DELETE("/:branch_id/organization/:organization_id", c.BranchDelete)
		branchG.GET("/branch/organization/:organization_id", c.BranchOrganizations)
	}

	categoryG := service.Group("/category")
	{
		categoryG.GET("/", c.CategoryList)
		categoryG.GET("/category_id", c.CategoryGetByID)
		categoryG.POST("/", c.CategoryCreate)
		categoryG.PUT("/category_id", c.CategoryUpdate)
		categoryG.DELETE("/category_id", c.CategoryDelete)
	}

	contactUsG := service.Group("/contact-us")
	{
		contactUsG.GET("/", c.ContactUsList)
		contactUsG.GET("/contact_us_id", c.ContactUsGetByID)
		contactUsG.POST("/", c.ContactUsCreate)
		contactUsG.PUT("/contact_us_id", c.ContactUsUpdate)
		contactUsG.DELETE("/contact_us_id", c.ContactUsDelete)
	}
}
