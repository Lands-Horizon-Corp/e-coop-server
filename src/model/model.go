package model

import (
	"context"

	"github.com/google/uuid"
	horizon_services "github.com/lands-horizon/horizon-server/services"
	"github.com/lands-horizon/horizon-server/src"
	"gorm.io/gorm"
)

type (
	QRMemberProfile struct {
		Firstname       string `json:"first_name"`
		Lastname        string `json:"last_name"`
		Middlename      string `json:"middle_name"`
		ContactNumber   string `json:"contact_number"`
		MemberProfileID string `json:"member_profile_id"`
		BranchID        string `json:"branch_id"`
		OrganizationID  string `json:"organization_id"`
		Email           string `json:"email"`
	}
	QRInvitationCode struct {
		OrganizationID string `json:"organization_id"`
		BranchID       string `json:"branch_id"`
		UserType       string `json:"user_type"`
		Code           string `json:"code"`
		CurrentUse     int    `json:"current_use"`
		Description    string `json:"description"`
	}

	QRUser struct {
		UserID        string `json:"user_id"`
		Email         string `json:"email"`
		ContactNumber string `json:"contact_number"`
		Username      string `json:"user_name"`
		Name          string `json:"name"`
		Lastname      string `json:"last_name"`
		Firstname     string `json:"first_name"`
		Middlename    string `json:"middle_name"`
	}
	Model struct {
		provider *src.Provider

		// Managers
		Migration                     []any
		BranchManager                 horizon_services.Repository[Branch, BranchResponse, BranchRequest]
		CategoryManager               horizon_services.Repository[Category, CategoryResponse, CategoryRequest]
		ContactUsManager              horizon_services.Repository[ContactUs, ContactUsResponse, ContactUsRequest]
		FeedbackManager               horizon_services.Repository[Feedback, FeedbackResponse, FeedbackRequest]
		FootstepManager               horizon_services.Repository[Footstep, FootstepResponse, any]
		GeneratedReportManager        horizon_services.Repository[GeneratedReport, GeneratedReportResponse, GeneratedReportRequest]
		InvitationCodeManager         horizon_services.Repository[InvitationCode, InvitationCodeResponse, InvitationCodeRequest]
		MediaManager                  horizon_services.Repository[Media, MediaResponse, MediaRequest]
		NotificationManager           horizon_services.Repository[Notification, NotificationResponse, any]
		OrganizationCategoryManager   horizon_services.Repository[OrganizationCategory, OrganizationCategoryResponse, OrganizationCategoryRequest]
		OrganizationDailyUsageManager horizon_services.Repository[OrganizationDailyUsage, OrganizationDailyUsageResponse, OrganizationDailyUsageRequest]
		OrganizationManager           horizon_services.Repository[Organization, OrganizationResponse, OrganizationRequest]
		PermissionTemplateManager     horizon_services.Repository[PermissionTemplate, PermissionTemplateResponse, PermissionTemplateRequest]
		SubscriptionPlanManager       horizon_services.Repository[SubscriptionPlan, SubscriptionPlanResponse, SubscriptionPlanRequest]
		UserOrganizationManager       horizon_services.Repository[UserOrganization, UserOrganizationResponse, UserOrganizationRequest]
		UserManager                   horizon_services.Repository[User, UserResponse, UserRegisterRequest]
		UserRatingManager             horizon_services.Repository[UserRating, UserRatingResponse, UserRatingRequest]

		// Members
		MemberAddressManager horizon_services.Repository[MemberAddress, MemberAddressReponse, MemberAddressRequest]
	}
)

func NewModel(provider *src.Provider) (*Model, error) {
	return &Model{
		provider: provider,
	}, nil
}

// Setting up Validator, Broadcaster, Model, and Automigration
func (c *Model) Start() error {

	// Models
	c.Branch()
	c.Category()
	c.ContactUs()
	c.Feedback()
	c.Footstep()
	c.GeneratedReport()
	c.InvitationCode()
	c.Media()
	c.Notification()
	c.OrganizationCategory()
	c.OrganizationDailyUsage()
	c.Organization()
	c.PermissionTemplate()
	c.SubscriptionPlan()
	c.UserOrganization()
	c.User()
	c.UserRating()

	if err := c.provider.Service.Database.Client().AutoMigrate(c.Migration...); err != nil {
		return err
	}
	return nil
}

func (c *Model) OrganizationSeeder(context context.Context, tx *gorm.DB, userId uuid.UUID, organizationId uuid.UUID, branchId uuid.UUID) error {
	return nil
}

func (c *Model) OrganizationDestroyer(context context.Context, tx *gorm.DB, userId uuid.UUID, organizationId uuid.UUID, branchId uuid.UUID) error {
	return nil
}
