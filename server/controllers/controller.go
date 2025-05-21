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
	qr             *horizon.HorizonQR

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
	// Maintenantce Table member
	memberCenter                *model.MemberCenterCollection
	memberClassification        *model.MemberClassificationCollection
	memberGender                *model.MemberGenderCollection
	memberGroup                 *model.MemberGroupCollection
	memberOccupation            *model.MemberOccupationCollection
	memberType                  *model.MemberTypeCollection
	memberProfile               *model.MemberProfileCollection
	memberVerification          *model.MemberVerificationCollection
	memberCenterHistory         *model.MemberCenterHistoryCollection
	memberClassificationHistory *model.MemberClassificationHistoryCollection
	memberGenderHistory         *model.MemberGenderHistoryCollection
	memberOccupationHistory     *model.MemberOccupationHistoryCollection
	memberGroupHistory          *model.MemberGroupHistoryCollection
	memberTypeHistory           *model.MemberTypeHistoryCollection
	// End Maintenantce table member
}

func NewController(
	authentication *horizon.HorizonAuthentication,
	storage *horizon.HorizonStorage,
	provider *providers.Providers,
	model *model.Model,
	database *horizon.HorizonDatabase,
	security *horizon.HorizonSecurity,
	qr *horizon.HorizonQR,

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

	memberCenter *model.MemberCenterCollection,
	memberClassification *model.MemberClassificationCollection,
	memberGender *model.MemberGenderCollection,
	memberGroup *model.MemberGroupCollection,
	memberOccupation *model.MemberOccupationCollection,
	memberType *model.MemberTypeCollection,
	memberProfile *model.MemberProfileCollection,
	memberVerification *model.MemberVerificationCollection,
	memberCenterHistory *model.MemberCenterHistoryCollection,
	memberClassificationHistory *model.MemberClassificationHistoryCollection,
	memberGenderHistory *model.MemberGenderHistoryCollection,
	memberOccupationHistory *model.MemberOccupationHistoryCollection,
	memberGroupHistory *model.MemberGroupHistoryCollection,
	memberTypeHistory *model.MemberTypeHistoryCollection,
) (*Controller, error) {
	return &Controller{
		authentication:              authentication,
		storage:                     storage,
		provider:                    provider,
		model:                       model,
		database:                    database,
		security:                    security,
		qr:                          qr,
		branch:                      branch,
		category:                    category,
		contactUs:                   contactUs,
		feedback:                    feedback,
		footstep:                    footstep,
		generatedReport:             generatedReport,
		invitationCode:              invitationCode,
		media:                       media,
		notification:                notification,
		organizationCategory:        organizationCategory,
		organizationDailyUsage:      organizationDailyUsage,
		organization:                organization,
		permissionTemplate:          permissionTemplate,
		subscriptionPlan:            subscriptionPlan,
		userOrganization:            userOrganization,
		user:                        user,
		userRating:                  userRating,
		memberCenter:                memberCenter,
		memberClassification:        memberClassification,
		memberGender:                memberGender,
		memberGroup:                 memberGroup,
		memberOccupation:            memberOccupation,
		memberType:                  memberType,
		memberProfile:               memberProfile,
		memberVerification:          memberVerification,
		memberCenterHistory:         memberCenterHistory,
		memberClassificationHistory: memberClassificationHistory,
		memberGenderHistory:         memberGenderHistory,
		memberOccupationHistory:     memberOccupationHistory,
		memberGroupHistory:          memberGroupHistory,
		memberTypeHistory:           memberTypeHistory,
	}, nil
}

func (c *Controller) Routes(service *echo.Echo) {
	qrG := service.Group("qr")
	{
		qrG.GET("/:code", c.QRCode)
	}
	branchG := service.Group("/branch")
	{
		branchG.GET("", c.BranchList)
		branchG.GET("/:branch_id", c.BranchGetByID)
		branchG.POST("/user-organization/:user_organization_id", c.BranchCreate)
		branchG.PUT("/user-organization/:user_organization_id", c.BranchUpdate)
		branchG.DELETE("/:branch_id/user-organization/:user_organization_id", c.BranchDelete)
		branchG.GET("/organization/:organization_id", c.BranchOrganizations)
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
		invitationCodeG.GET("/verify/:code", c.InvitationCodeVerify)
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
		authenticationG.POST("/verify-with-password", c.UserVerifyWithPassword)
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
		userOrganizationG.GET("/organization/:organization_id/branch/:branch_id/can-join-member", c.UserOrganizationCanJoinMember)
		userOrganizationG.GET("/organization/:organization_id/branch/:branch_id/can-join-employee", c.UserOrganizationCanJoinEmployee)
		userOrganizationG.GET("/:user_organization_id/switch", c.UserOrganizationSwitch)
		userOrganizationG.GET("/unswitch", c.UserOrganizationUnSwitch)
		userOrganizationG.GET("/:organization_id/seed", c.UserOrganizationSeeder)

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

	memberCenterG := service.Group("/member-center")
	{
		memberCenterG.GET("", c.MemberCenterList)
		memberCenterG.GET("/:member_center_id", c.MemberCenterGetByID)
		memberCenterG.POST("", c.MemberCenterCreate)
		memberCenterG.PUT("/:member_center_id", c.MemberCenterUpdate)
		memberCenterG.DELETE("/:member_center_id", c.MemberCenterDelete)
		memberCenterG.GET("/branch/:branch_id", c.MemberCenterListByBranch)
		memberCenterG.GET("/organization/:organization_id", c.MemberCenterListByOrganization)
		memberCenterG.GET("/organization/:organization_id/branch/:branch_id", c.MemberCenterListByOrganizationBranch)
	}

	memberClassificationG := service.Group("/member-classification")
	{
		memberClassificationG.GET("", c.MemberClassificationList)
		memberClassificationG.GET("/:member_classification_id", c.MemberClassificationGetByID)
		memberClassificationG.POST("", c.MemberClassificationCreate)
		memberClassificationG.PUT("/:member_classification_id", c.MemberClassificationUpdate)
		memberClassificationG.DELETE("/:member_classification_id", c.MemberClassificationDelete)
		memberClassificationG.GET("/branch/:branch_id", c.MemberClassificationListByBranch)
		memberClassificationG.GET("/organization/:organization_id", c.MemberClassificationListByOrganization)
		memberClassificationG.GET("/organization/:organization_id/branch/:branch_id", c.MemberClassificationListByOrganizationBranch)
	}
	memberGenderG := service.Group("/member-gender")
	{
		memberGenderG.GET("", c.MemberGenderList)
		memberGenderG.GET("/:member_gender_id", c.MemberGenderGetByID)
		memberGenderG.POST("", c.MemberGenderCreate)
		memberGenderG.PUT("/:member_gender_id", c.MemberGenderUpdate)
		memberGenderG.DELETE("/:member_gender_id", c.MemberGenderDelete)
		memberGenderG.GET("/branch/:branch_id", c.MemberGenderListByBranch)
		memberGenderG.GET("/organization/:organization_id", c.MemberGenderListByOrganization)
		memberGenderG.GET("/organization/:organization_id/branch/:branch_id", c.MemberGenderListByOrganizationBranch)
	}

	memberGroupG := service.Group("/member-group")
	{
		memberGroupG.GET("", c.MemberGroupList)
		memberGroupG.GET("/:member_group_id", c.MemberGroupGetByID)
		memberGroupG.POST("", c.MemberGroupCreate)
		memberGroupG.PUT("/:member_group_id", c.MemberGroupUpdate)
		memberGroupG.DELETE("/:member_group_id", c.MemberGroupDelete)
		memberGroupG.GET("/branch/:branch_id", c.MemberGroupListByBranch)
		memberGroupG.GET("/organization/:organization_id", c.MemberGroupListByOrganization)
		memberGroupG.GET("/organization/:organization_id/branch/:branch_id", c.MemberGroupListByOrganizationBranch)
	}

	memberOccupationG := service.Group("/member-occupation")
	{
		memberOccupationG.GET("", c.MemberOccupationList)
		memberOccupationG.GET("/:member_occupation_id", c.MemberOccupationGetByID)
		memberOccupationG.POST("", c.MemberOccupationCreate)
		memberOccupationG.PUT("/:member_occupation_id", c.MemberOccupationUpdate)
		memberOccupationG.DELETE("/:member_occupation_id", c.MemberOccupationDelete)
		memberOccupationG.GET("/branch/:branch_id", c.MemberOccupationListByBranch)
		memberOccupationG.GET("/organization/:organization_id", c.MemberOccupationListByOrganization)
		memberOccupationG.GET("/organization/:organization_id/branch/:branch_id", c.MemberOccupationListByOrganizationBranch)
	}

	memberTypeG := service.Group("/member-type")
	{
		memberTypeG.GET("", c.MemberTypeList)
		memberTypeG.GET("/:member_type_id", c.MemberTypeGetByID)
		memberTypeG.POST("", c.MemberTypeCreate)
		memberTypeG.PUT("/:member_type_id", c.MemberTypeUpdate)
		memberTypeG.DELETE("/:member_type_id", c.MemberTypeDelete)
		memberTypeG.GET("/branch/:branch_id", c.MemberTypeListByBranch)
		memberTypeG.GET("/organization/:organization_id", c.MemberTypeListByOrganization)
		memberTypeG.GET("/organization/:organization_id/branch/:branch_id", c.MemberTypeListByOrganizationBranch)
	}

	memberProfileG := service.Group("/member-profile")
	{
		memberProfileG.GET("", c.MemberProfileList)
		memberProfileG.GET("/:member_profile_id", c.MemberProfileGetByID)
		memberProfileG.POST("", c.MemberProfileCreate)
		memberProfileG.PUT("/:member_profile_id", c.MemberProfileUpdate)
		memberProfileG.DELETE("/:member_profile_id", c.MemberProfileDelete)
		memberProfileG.GET("/branch/:branch_id", c.MemberProfileListByBranch)
		memberProfileG.GET("/organization/:organization_id", c.MemberProfileListByOrganization)
		memberProfileG.GET("/organization/:organization_id/branch/:branch_id", c.MemberProfileListByOrganizationBranch)
	}

	memberVerificationG := service.Group("/member-verification")
	{
		memberVerificationG.GET("", c.MemberVerificationList)
		memberVerificationG.GET("/:member_verification_id", c.MemberVerificationGetByID)
		memberVerificationG.DELETE("/:member_verification_id", c.MemberVerificationDelete)
		memberVerificationG.GET("/branch/:branch_id", c.MemberVerificationListByBranch)
		memberVerificationG.GET("/organization/:organization_id", c.MemberVerificationListByOrganization)
		memberVerificationG.GET("/organization/:organization_id/branch/:branch_id", c.MemberVerificationListByOrganizationBranch)
	}

	memberGenderHistoryG := service.Group("/member-gender-history")
	{
		memberGenderHistoryG.GET("", c.MemberGenderHistoryList)
		memberGenderHistoryG.GET("/:member_gender_history_id", c.MemberGenderHistoryGetByID)
		memberGenderHistoryG.DELETE("/:member_gender_history_id", c.MemberGenderHistoryDelete)
		memberGenderHistoryG.GET("/branch/:branch_id", c.MemberGenderHistoryListByBranch)
		memberGenderHistoryG.GET("/organization/:organization_id", c.MemberGenderHistoryListByOrganization)
		memberGenderHistoryG.GET("/organization/:organization_id/branch/:branch_id", c.MemberGenderHistoryListByOrganizationBranch)
	}

	memberTypeHistoryG := service.Group("/member-type-history")
	{
		memberTypeHistoryG.GET("", c.MemberTypeHistoryList)
		memberTypeHistoryG.GET("/:member_type_history_id", c.MemberTypeHistoryGetByID)
		memberTypeHistoryG.DELETE("/:member_type_history_id", c.MemberTypeHistoryDelete)
		memberTypeHistoryG.GET("/branch/:branch_id", c.MemberTypeHistoryListByBranch)
		memberTypeHistoryG.GET("/organization/:organization_id", c.MemberTypeHistoryListByOrganization)
		memberTypeHistoryG.GET("/organization/:organization_id/branch/:branch_id", c.MemberTypeHistoryListByOrganizationBranch)
	}
	memberOccupationHistoryG := service.Group("/member-occupation-history")
	{
		memberOccupationHistoryG.GET("", c.MemberOccupationHistoryList)
		memberOccupationHistoryG.GET("/:member_occupation_history_id", c.MemberOccupationHistoryGetByID)
		memberOccupationHistoryG.DELETE("/:member_occupation_history_id", c.MemberOccupationHistoryDelete)
		memberOccupationHistoryG.GET("/branch/:branch_id", c.MemberOccupationHistoryListByBranch)
		memberOccupationHistoryG.GET("/organization/:organization_id", c.MemberOccupationHistoryListByOrganization)
		memberOccupationHistoryG.GET("/organization/:organization_id/branch/:branch_id", c.MemberOccupationHistoryListByOrganizationBranch)
	}

	memberGroupHistoryG := service.Group("/member-group-history")
	{
		memberGroupHistoryG.GET("", c.MemberGroupHistoryList)
		memberGroupHistoryG.GET("/:member_group_history_id", c.MemberGroupHistoryGetByID)
		memberGroupHistoryG.DELETE("/:member_group_history_id", c.MemberGroupHistoryDelete)
		memberGroupHistoryG.GET("/branch/:branch_id", c.MemberGroupHistoryListByBranch)
		memberGroupHistoryG.GET("/organization/:organization_id", c.MemberGroupHistoryListByOrganization)
		memberGroupHistoryG.GET("/organization/:organization_id/branch/:branch_id", c.MemberGroupHistoryListByOrganizationBranch)
	}

	memberCenterHistoryG := service.Group("/member-center-history")
	{
		memberCenterHistoryG.GET("", c.MemberCenterHistoryList)
		memberCenterHistoryG.GET("/:member_center_history_id", c.MemberCenterHistoryGetByID)
		memberCenterHistoryG.DELETE("/:member_center_history_id", c.MemberCenterHistoryDelete)
		memberCenterHistoryG.GET("/branch/:branch_id", c.MemberCenterHistoryListByBranch)
		memberCenterHistoryG.GET("/organization/:organization_id", c.MemberCenterHistoryListByOrganization)
		memberCenterHistoryG.GET("/organization/:organization_id/branch/:branch_id", c.MemberCenterHistoryListByOrganizationBranch)
	}

	memberClassificationHistoryG := service.Group("/member-classification-history")
	{
		memberClassificationHistoryG.GET("", c.MemberClassificationHistoryList)
		memberClassificationHistoryG.GET("/:member_classification_history_id", c.MemberClassificationHistoryGetByID)
		memberClassificationHistoryG.DELETE("/:member_classification_history_id", c.MemberClassificationHistoryDelete)
		memberClassificationHistoryG.GET("/branch/:branch_id", c.MemberClassificationHistoryListByBranch)
		memberClassificationHistoryG.GET("/organization/:organization_id", c.MemberClassificationHistoryListByOrganization)
		memberClassificationHistoryG.GET("/organization/:organization_id/branch/:branch_id", c.MemberClassificationHistoryListByOrganizationBranch)
	}
}
