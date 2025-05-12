package controllers

import (
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

func NewHandler(
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
