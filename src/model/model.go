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
		Migration []any

		BankManager                   horizon_services.Repository[Bank, BankResponse, BankRequest]
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
		MemberAddressManager               horizon_services.Repository[MemberAddress, MemberAddressReponse, MemberAddressRequest]
		MemberAssetManager                 horizon_services.Repository[MemberAsset, MemberAssetResponse, MemberAssetRequest]
		MemberBankCardManager              horizon_services.Repository[MemberBankCard, MemberBankCardResponse, MemberBankCardRequest]
		MemberCenterHistoryManager         horizon_services.Repository[MemberCenterHistory, MemberCenterHistoryResponse, MemberCenterHistoryRequest]
		MemberCenterManager                horizon_services.Repository[MemberCenter, MemberCenterResponse, MemberCenterRequest]
		MemberClassificationManager        horizon_services.Repository[MemberClassification, MemberClassificationResponse, MemberClassificationRequest]
		MemberCloseRemarkManager           horizon_services.Repository[MemberCloseRemark, MemberCloseRemarkResponse, MemberCloseRemarkRequest]
		MemberContactReferenceManager      horizon_services.Repository[MemberContactReference, MemberContactReferenceResponse, MemberContactReferenceRequest]
		MemberDamayanExtensionEntryManager horizon_services.Repository[MemberDamayanExtensionEntry, MemberDamayanExtensionEntryResponse, MemberDamayanExtensionEntryRequest]
		MemberEducationalAttainmentManager horizon_services.Repository[MemberEducationalAttainment, MemberEducationalAttainmentResponse, MemberEducationalAttainmentRequest]
		MemberExpenseManager               horizon_services.Repository[MemberExpense, MemberExpenseResponse, MemberExpenseRequest]
		MemberGenderHistoryManager         horizon_services.Repository[MemberGenderHistory, MemberGenderHistoryResponse, MemberGenderHistoryRequest]
		MemberGenderManager                horizon_services.Repository[MemberGender, MemberGenderResponse, MemberGenderRequest]
		MemberGovernmentBenefitManager     horizon_services.Repository[MemberGovernmentBenefit, MemberGovernmentBenefitResponse, MemberGovernmentBenefitRequest]
		MemberGroupHistoryManager          horizon_services.Repository[MemberGroupHistory, MemberGroupHistoryResponse, MemberGroupHistoryRequest]
		MemberGroupManager                 horizon_services.Repository[MemberGroup, MemberGroupResponse, MemberGroupRequest]
		MemberIncomeManager                horizon_services.Repository[MemberIncome, MemberIncomeResponse, MemberIncomeRequest]
		MemberJointAccountManager          horizon_services.Repository[MemberJointAccount, MemberJointAccountResponse, MemberJointAccountRequest]
		MemberMutualFundHistoryManager     horizon_services.Repository[MemberMutualFundHistory, MemberMutualFundHistoryResponse, MemberMutualFundHistoryRequest]
		MemberOccupationHistoryManager     horizon_services.Repository[MemberOccupationHistory, MemberOccupationHistoryResponse, MemberOccupationHistoryRequest]
		MemberOccupationManager            horizon_services.Repository[MemberOccupation, MemberOccupationResponse, MemberOccupationRequest]
		MemberOtherInformationEntryManager horizon_services.Repository[MemberOtherInformationEntry, MemberOtherInformationEntryResponse, MemberOtherInformationEntryRequest]
		MemberRelativeAccountManager       horizon_services.Repository[MemberRelativeAccount, MemberRelativeAccountResponse, MemberRelativeAccountRequest]
		MemberTypeHistoryManager           horizon_services.Repository[MemberTypeHistory, MemberTypeHistoryResponse, MemberTypeHistoryRequest]
		MemberTypeManager                  horizon_services.Repository[MemberType, MemberTypeResponse, MemberTypeRequest]
		MemberVerificationManager          horizon_services.Repository[MemberVerification, MemberVerificationResponse, MemberVerificationRequest]
		MemberProfileManager               horizon_services.Repository[MemberProfile, MemberProfileResponse, MemberProfileRequest]
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
	c.Bank()
	c.Branch()
	c.Category()
	c.ContactUs()
	c.Feedback()
	c.Footstep()
	c.GeneratedReport()
	c.InvitationCode()
	c.Media()
	c.MemberAddress()
	c.MemberAsset()
	c.MemberBankCard()
	c.MemberCenter()
	c.MemberCenterHistory()
	c.MemberClassification()
	c.MemberCloseRemark()
	c.MemberContactReference()
	c.MemberDamayanExtensionEntry()
	c.MemberEducationalAttainment()
	c.MemberExpense()
	c.MemberGender()
	c.MemberGenderHistory()
	c.MemberGovernmentBenefit()
	c.MemberGroup()
	c.MemberGroupHistory()
	c.MemberIncome()
	c.MemberJointAccount()
	c.MemberMutualFundHistory()
	c.MemberOccupation()
	c.MemberOccupationHistory()
	c.MemberOtherInformationEntry()
	c.MemberProfile()
	c.MemberRelativeAccount()
	c.MemberType()
	c.MemberTypeHistory()
	c.MemberVerification()
	c.Notification()
	c.OrganizationCategory()
	c.OrganizationDailyUsage()
	c.Organization()
	c.PermissionTemplate()
	c.SubscriptionPlan()
	c.User()
	c.UserOrganization()
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
