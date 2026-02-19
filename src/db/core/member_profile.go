package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/pkg/registry"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

func MemberProfileManager(service *horizon.HorizonService) *registry.Registry[
	types.MemberProfile, types.MemberProfileResponse, types.MemberProfileRequest] {
	return registry.NewRegistry(registry.RegistryParams[types.MemberProfile, types.MemberProfileResponse, types.MemberProfileRequest]{
		Preloads: []string{
			"CreatedBy", "UpdatedBy",

			"Branch.Media", "Organization.Media",
			"Media", "SignatureMedia",
			"User",
			"User.Media",
			"MemberType", "MemberGroup", "MemberGender", "MemberCenter",
			"MemberOccupation", "MemberClassification", "MemberVerifiedByEmployeeUser",
			"MemberVerifiedByEmployeeUser.Media",

			"MemberAddresses",
			"MemberAssets", "MemberAssets.Media",
			"MemberIncomes", "MemberIncomes.Media",
			"MemberExpenses",
			"MemberGovernmentBenefits", "MemberGovernmentBenefits.FrontMedia", "MemberGovernmentBenefits.BackMedia",
			"MemberJointAccounts", "MemberJointAccounts.PictureMedia", "MemberJointAccounts.SignatureMedia",

			"MemberRelativeAccounts", "MemberRelativeAccounts.RelativeMemberProfile", "MemberRelativeAccounts.RelativeMemberProfile.Media",
			"RecruitedByMemberProfile", "RecruitedByMemberProfile.Media",
			"RecruitedMembers", "RecruitedMembers.Media",
			"MemberEducationalAttainments",
			"MemberContactReferences",
			"MemberCloseRemarks",
			"MemberDepartment",
		},
		Database: service.Database.Client(),
		Dispatch: func(topics registry.Topics, payload any) error {
			return service.Broker.Dispatch(topics, payload)
		},
		Resource: func(data *types.MemberProfile) *types.MemberProfileResponse {
			context := context.Background()
			if data == nil {
				return nil
			}

			result, err := service.QR.EncodeQR(context, &types.QRMemberProfile{
				FirstName:       data.FirstName,
				LastName:        data.LastName,
				MiddleName:      data.MiddleName,
				ContactNumber:   data.ContactNumber,
				MemberProfileID: data.ID.String(),
				BranchID:        data.BranchID.String(),
				OrganizationID:  data.OrganizationID.String(),
				FullName:        data.FullName,
			}, "member-qr")
			if err != nil {
				return nil
			}
			return &types.MemberProfileResponse{
				ID:                             data.ID,
				CreatedAt:                      data.CreatedAt.Format(time.RFC3339),
				CreatedByID:                    *data.CreatedByID,
				CreatedBy:                      UserManager(service).ToModel(data.CreatedBy),
				UpdatedAt:                      data.UpdatedAt.Format(time.RFC3339),
				UpdatedByID:                    *data.UpdatedByID,
				UpdatedBy:                      UserManager(service).ToModel(data.UpdatedBy),
				OrganizationID:                 data.OrganizationID,
				Organization:                   OrganizationManager(service).ToModel(data.Organization),
				BranchID:                       data.BranchID,
				Branch:                         BranchManager(service).ToModel(data.Branch),
				MediaID:                        data.MediaID,
				Media:                          MediaManager(service).ToModel(data.Media),
				SignatureMediaID:               data.SignatureMediaID,
				SignatureMedia:                 MediaManager(service).ToModel(data.SignatureMedia),
				UserID:                         data.UserID,
				User:                           UserManager(service).ToModel(data.User),
				MemberTypeID:                   data.MemberTypeID,
				MemberType:                     MemberTypeManager(service).ToModel(data.MemberType),
				MemberGroupID:                  data.MemberGroupID,
				MemberGroup:                    MemberGroupManager(service).ToModel(data.MemberGroup),
				MemberGenderID:                 data.MemberGenderID,
				MemberGender:                   MemberGenderManager(service).ToModel(data.MemberGender),
				MemberCenterID:                 data.MemberCenterID,
				MemberCenter:                   MemberCenterManager(service).ToModel(data.MemberCenter),
				MemberOccupationID:             data.MemberOccupationID,
				MemberOccupation:               MemberOccupationManager(service).ToModel(data.MemberOccupation),
				MemberClassificationID:         data.MemberClassificationID,
				MemberClassification:           MemberClassificationManager(service).ToModel(data.MemberClassification),
				MemberVerifiedByEmployeeUserID: data.MemberVerifiedByEmployeeUserID,
				MemberVerifiedByEmployeeUser:   UserManager(service).ToModel(data.MemberVerifiedByEmployeeUser),
				RecruitedByMemberProfileID:     data.RecruitedByMemberProfileID,
				RecruitedByMemberProfile:       MemberProfileManager(service).ToModel(data.RecruitedByMemberProfile),
				IsClosed:                       data.IsClosed,
				IsMutualFundMember:             data.IsMutualFundMember,
				IsMicroFinanceMember:           data.IsMicroFinanceMember,
				FirstName:                      data.FirstName,
				MiddleName:                     data.MiddleName,
				LastName:                       data.LastName,
				FullName:                       data.FullName,
				Suffix:                         data.Suffix,
				BirthDate:                      data.BirthDate.Format(time.RFC3339),
				Status:                         data.Status,
				Description:                    data.Description,
				Notes:                          data.Notes,
				ContactNumber:                  data.ContactNumber,
				OldReferenceID:                 data.OldReferenceID,
				Passbook:                       data.Passbook,
				BusinessAddress:                data.BusinessAddress,
				BusinessContactNumber:          data.BusinessContactNumber,
				CivilStatus:                    data.CivilStatus,
				QRCode:                         result,
				MemberAddresses:                MemberAddressManager(service).ToModels(data.MemberAddresses),
				MemberAssets:                   MemberAssetManager(service).ToModels(data.MemberAssets),
				MemberIncomes:                  MemberIncomeManager(service).ToModels(data.MemberIncomes),
				MemberExpenses:                 MemberExpenseManager(service).ToModels(data.MemberExpenses),
				MemberGovernmentBenefits:       MemberGovernmentBenefitManager(service).ToModels(data.MemberGovernmentBenefits),
				MemberJointAccounts:            MemberJointAccountManager(service).ToModels(data.MemberJointAccounts),
				MemberRelativeAccounts:         MemberRelativeAccountManager(service).ToModels(data.MemberRelativeAccounts),
				MemberEducationalAttainments:   MemberEducationalAttainmentManager(service).ToModels(data.MemberEducationalAttainments),
				MemberContactReferences:        MemberContactReferenceManager(service).ToModels(data.MemberContactReferences),
				MemberCloseRemarks:             MemberCloseRemarkManager(service).ToModels(data.MemberCloseRemarks),
				RecruitedMembers:               MemberProfileManager(service).ToModels(data.RecruitedMembers),
				MemberDepartmentID:             data.MemberDepartmentID,
				MemberDepartment:               MemberDepartmentManager(service).ToModel(data.MemberDepartment),
				BirthPlace:                     data.BirthPlace,
				Sex:                            data.Sex,
			}
		},

		Created: func(data *types.MemberProfile) registry.Topics {
			return []string{
				"member_profile.create",
				fmt.Sprintf("member_profile.create.%s", data.ID),
				fmt.Sprintf("member_profile.create.branch.%s", data.BranchID),
				fmt.Sprintf("member_profile.create.organization.%s", data.OrganizationID),
			}
		},
		Updated: func(data *types.MemberProfile) registry.Topics {
			return []string{
				"member_profile.update",
				fmt.Sprintf("member_profile.update.%s", data.ID),
				fmt.Sprintf("member_profile.update.branch.%s", data.BranchID),
				fmt.Sprintf("member_profile.update.organization.%s", data.OrganizationID),
			}
		},
		Deleted: func(data *types.MemberProfile) registry.Topics {
			return []string{
				"member_profile.delete",
				fmt.Sprintf("member_profile.delete.%s", data.ID),
				fmt.Sprintf("member_profile.delete.branch.%s", data.BranchID),
				fmt.Sprintf("member_profile.delete.organization.%s", data.OrganizationID),
			}
		},
	})
}

func MemberProfileCurrentBranch(context context.Context, service *horizon.HorizonService, organizationID uuid.UUID,
	branchID uuid.UUID) ([]*types.MemberProfile, error) {
	return MemberProfileManager(service).Find(context, &types.MemberProfile{
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func MemberProfileDelete(context context.Context, service *horizon.HorizonService, tx *gorm.DB, memberProfileID uuid.UUID) error {
	memberEducationalAttainments, err := MemberEducationalAttainmentManager(service).Find(context, &types.MemberEducationalAttainment{
		MemberProfileID: memberProfileID,
	})
	if err != nil {
		return err
	}
	for _, value := range memberEducationalAttainments {
		if err := MemberEducationalAttainmentManager(service).DeleteWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}

	memberCloseRemarks, err := MemberCloseRemarkManager(service).Find(context, &types.MemberCloseRemark{
		MemberProfileID: &memberProfileID,
	})
	if err != nil {
		return err
	}
	for _, value := range memberCloseRemarks {
		if err := MemberCloseRemarkManager(service).DeleteWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}

	memberAddresses, err := MemberAddressManager(service).Find(context, &types.MemberAddress{
		MemberProfileID: &memberProfileID,
	})
	if err != nil {
		return err
	}
	for _, value := range memberAddresses {
		if err := MemberAddressManager(service).DeleteWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}

	memberContactReferences, err := MemberContactReferenceManager(service).Find(context, &types.MemberContactReference{
		MemberProfileID: memberProfileID,
	})
	if err != nil {
		return err
	}
	for _, value := range memberContactReferences {
		if err := MemberContactReferenceManager(service).DeleteWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}

	memberAssets, err := MemberAssetManager(service).Find(context, &types.MemberAsset{
		MemberProfileID: &memberProfileID,
	})
	if err != nil {
		return err
	}
	for _, value := range memberAssets {
		if err := MemberAssetManager(service).DeleteWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}

	memberIncomes, err := MemberIncomeManager(service).Find(context, &types.MemberIncome{
		MemberProfileID: memberProfileID,
	})
	if err != nil {
		return err
	}
	for _, value := range memberIncomes {
		if err := MemberIncomeManager(service).DeleteWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}
	memberExpenses, err := MemberExpenseManager(service).Find(context, &types.MemberExpense{
		MemberProfileID: memberProfileID,
	})
	if err != nil {
		return err
	}
	for _, value := range memberExpenses {
		if err := MemberExpenseManager(service).DeleteWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}
	memberGovernmentBenefits, err := MemberGovernmentBenefitManager(service).Find(context, &types.MemberGovernmentBenefit{
		MemberProfileID: memberProfileID,
	})
	if err != nil {
		return err
	}
	for _, value := range memberGovernmentBenefits {
		if err := MemberGovernmentBenefitManager(service).DeleteWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}
	memberJointAccounts, err := MemberJointAccountManager(service).Find(context, &types.MemberJointAccount{
		MemberProfileID: memberProfileID,
	})
	if err != nil {
		return err
	}
	for _, value := range memberJointAccounts {
		if err := MemberJointAccountManager(service).DeleteWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}
	memberRelativeAccounts, err := MemberRelativeAccountManager(service).Find(context, &types.MemberRelativeAccount{
		MemberProfileID: memberProfileID,
	})
	if err != nil {
		return err
	}
	for _, value := range memberRelativeAccounts {
		if err := MemberRelativeAccountManager(service).DeleteWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}

	memberGenderHistories, err := MemberGenderHistoryManager(service).Find(context, &types.MemberGenderHistory{
		MemberProfileID: memberProfileID,
	})
	if err != nil {
		return err
	}
	for _, value := range memberGenderHistories {
		if err := MemberGenderHistoryManager(service).DeleteWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}

	memberCenterHistories, err := MemberCenterHistoryManager(service).Find(context, &types.MemberCenterHistory{
		MemberProfileID: memberProfileID,
	})
	if err != nil {
		return err
	}
	for _, value := range memberCenterHistories {
		if err := MemberCenterHistoryManager(service).DeleteWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}

	memberTypeHistories, err := MemberTypeHistoryManager(service).Find(context, &types.MemberTypeHistory{
		MemberProfileID: memberProfileID,
	})
	if err != nil {
		return err
	}
	for _, value := range memberTypeHistories {
		if err := MemberTypeHistoryManager(service).DeleteWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}

	memberClassificationHistories, err := MemberClassificationHistoryManager(service).Find(context, &types.MemberClassificationHistory{
		MemberProfileID: memberProfileID,
	})
	if err != nil {
		return err
	}
	for _, value := range memberClassificationHistories {
		if err := MemberClassificationHistoryManager(service).DeleteWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}

	memberOccupationHistories, err := MemberOccupationHistoryManager(service).Find(context, &types.MemberOccupationHistory{
		MemberProfileID: memberProfileID,
	})
	if err != nil {
		return err
	}
	for _, value := range memberOccupationHistories {
		if err := MemberOccupationHistoryManager(service).DeleteWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}

	memberGroupHistories, err := MemberGroupHistoryManager(service).Find(context, &types.MemberGroupHistory{
		MemberProfileID: memberProfileID,
	})
	if err != nil {
		return err
	}
	for _, value := range memberGroupHistories {
		if err := MemberGroupHistoryManager(service).DeleteWithTx(context, tx, value.ID); err != nil {
			return err
		}
	}

	return MemberProfileManager(service).DeleteWithTx(context, tx, memberProfileID)
}

func MemberProfileFindUserByID(ctx context.Context, service *horizon.HorizonService,
	userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) (*types.MemberProfile, error) {
	return MemberProfileManager(service).FindOne(ctx, &types.MemberProfile{
		UserID:         &userID,
		OrganizationID: organizationID,
		BranchID:       branchID,
	})
}

func memberProfileSeed(context context.Context, service *horizon.HorizonService, tx *gorm.DB,
	userID uuid.UUID, organizationID uuid.UUID, branchID uuid.UUID) error {
	now := time.Now().UTC()

	branch, err := BranchManager(service).GetByID(context, branchID)
	if err != nil {
		return eris.Wrapf(err, "failed to get branch by ID: %s", branchID)
	}
	organization, err := OrganizationManager(service).GetByID(context, organizationID)
	if err != nil {
		return eris.Wrapf(err, "failed to get organization by ID: %s", organizationID)
	}

	firstName := "Company: " + organization.Name

	middleName := ""

	lastName := branch.Name

	fullName := fmt.Sprintf("%s %s %s", firstName, middleName, lastName)
	passbook := fmt.Sprintf("PB-%s-0001", branch.Name[:min(3, len(branch.Name))])
	memberProfile := &types.MemberProfile{
		CreatedAt:             now,
		CreatedByID:           &userID,
		UpdatedAt:             now,
		UpdatedByID:           &userID,
		OrganizationID:        organizationID,
		BranchID:              branchID,
		MediaID:               branch.MediaID,
		UserID:                &userID,
		FirstName:             firstName,
		MiddleName:            middleName,
		LastName:              lastName,
		FullName:              fullName,
		BirthDate:             &now,
		Status:                "active",
		Description:           fmt.Sprintf("Founding member of %s", organization.Name),
		Notes:                 fmt.Sprintf("Organization founder and branch creator for %s", branch.Name),
		ContactNumber:         *branch.ContactNumber,
		OldReferenceID:        "FOUNDER-001",
		Passbook:              passbook,
		Occupation:            "Organization Founder",
		BusinessAddress:       branch.Address,
		BusinessContactNumber: *branch.ContactNumber,
		CivilStatus:           "single",
		IsClosed:              false,
		IsMutualFundMember:    true,
		IsMicroFinanceMember:  true,
	}

	if err := MemberProfileManager(service).CreateWithTx(context, tx, memberProfile); err != nil {
		return eris.Wrapf(err, "failed to create founder member profile %s", memberProfile.FullName)
	}

	return nil
}

func MemberProfileDestroy(ctx context.Context, service *horizon.HorizonService, tx *gorm.DB, id uuid.UUID) error {
	memberProfile, err := MemberProfileManager(service).GetByID(ctx, id)
	if err != nil {
		return eris.Wrapf(err, "failed to get MemberProfile by ID: %s", id)
	}
	memberAddresses, err := MemberAddressManager(service).Find(ctx, &types.MemberAddress{
		MemberProfileID: &memberProfile.ID,
		BranchID:        memberProfile.BranchID,
		OrganizationID:  memberProfile.OrganizationID,
	})
	if err != nil {
		return eris.Wrap(err, "failed to find member addresses")
	}
	for _, memberAddress := range memberAddresses {
		if err := MemberAddressManager(service).DeleteWithTx(ctx, tx, memberAddress.ID); err != nil {
			return eris.Wrapf(err, "failed to delete member address: %s", memberAddress.ID)
		}
	}

	memberAssets, err := MemberAssetManager(service).Find(ctx, &types.MemberAsset{
		MemberProfileID: &memberProfile.ID,
		BranchID:        memberProfile.BranchID,
		OrganizationID:  memberProfile.OrganizationID,
	})
	if err != nil {
		return eris.Wrap(err, "failed to find member assets")
	}
	for _, memberAsset := range memberAssets {
		if err := MemberAssetManager(service).DeleteWithTx(ctx, tx, memberAsset.ID); err != nil {
			return eris.Wrapf(err, "failed to delete member asset: %s", memberAsset.ID)
		}
	}

	memberIncomes, err := MemberIncomeManager(service).Find(ctx, &types.MemberIncome{
		MemberProfileID: memberProfile.ID,
		BranchID:        memberProfile.BranchID,
		OrganizationID:  memberProfile.OrganizationID,
	})
	if err != nil {
		return eris.Wrap(err, "failed to find member incomes")
	}
	for _, memberIncome := range memberIncomes {
		if err := MemberIncomeManager(service).DeleteWithTx(ctx, tx, memberIncome.ID); err != nil {
			return eris.Wrapf(err, "failed to delete member income: %s", memberIncome.ID)
		}
	}

	memberExpenses, err := MemberExpenseManager(service).Find(ctx, &types.MemberExpense{
		MemberProfileID: memberProfile.ID,
		BranchID:        memberProfile.BranchID,
		OrganizationID:  memberProfile.OrganizationID,
	})
	if err != nil {
		return eris.Wrap(err, "failed to find member expenses")
	}
	for _, memberExpense := range memberExpenses {
		if err := MemberExpenseManager(service).DeleteWithTx(ctx, tx, memberExpense.ID); err != nil {
			return eris.Wrapf(err, "failed to delete member expense: %s", memberExpense.ID)
		}
	}

	memberBenefits, err := MemberGovernmentBenefitManager(service).Find(ctx, &types.MemberGovernmentBenefit{
		MemberProfileID: memberProfile.ID,
		BranchID:        memberProfile.BranchID,
		OrganizationID:  memberProfile.OrganizationID,
	})
	if err != nil {
		return eris.Wrap(err, "failed to find member government benefits")
	}
	for _, memberBenefit := range memberBenefits {
		if err := MemberGovernmentBenefitManager(service).DeleteWithTx(ctx, tx, memberBenefit.ID); err != nil {
			return eris.Wrapf(err, "failed to delete member government benefit: %s", memberBenefit.ID)
		}
	}
	memberJointAccounts, err := MemberJointAccountManager(service).Find(ctx, &types.MemberJointAccount{
		MemberProfileID: memberProfile.ID,
		BranchID:        memberProfile.BranchID,
		OrganizationID:  memberProfile.OrganizationID,
	})
	if err != nil {
		return eris.Wrap(err, "failed to find member joint accounts")
	}
	for _, memberJointAccount := range memberJointAccounts {
		if err := MemberJointAccountManager(service).DeleteWithTx(ctx, tx, memberJointAccount.ID); err != nil {
			return eris.Wrapf(err, "failed to delete member joint account: %s", memberJointAccount.ID)
		}
	}

	memberRelativeAccounts, err := MemberRelativeAccountManager(service).Find(ctx, &types.MemberRelativeAccount{
		MemberProfileID: memberProfile.ID,
		BranchID:        memberProfile.BranchID,
		OrganizationID:  memberProfile.OrganizationID,
	})
	if err != nil {
		return eris.Wrap(err, "failed to find member relative accounts")
	}
	for _, memberRelativeAccount := range memberRelativeAccounts {
		if err := MemberRelativeAccountManager(service).DeleteWithTx(ctx, tx, memberRelativeAccount.ID); err != nil {
			return eris.Wrapf(err, "failed to delete member relative account: %s", memberRelativeAccount.ID)
		}
	}
	memberEducations, err := MemberEducationalAttainmentManager(service).Find(ctx, &types.MemberEducationalAttainment{
		MemberProfileID: memberProfile.ID,
		BranchID:        memberProfile.BranchID,
		OrganizationID:  memberProfile.OrganizationID,
	})
	if err != nil {
		return eris.Wrap(err, "failed to find member educational attainments")
	}
	for _, memberEducation := range memberEducations {
		if err := MemberEducationalAttainmentManager(service).DeleteWithTx(ctx, tx, memberEducation.ID); err != nil {
			return eris.Wrapf(err, "failed to delete member educational attainment: %s", memberEducation.ID)
		}
	}
	memberContacts, err := MemberContactReferenceManager(service).Find(ctx, &types.MemberContactReference{
		MemberProfileID: memberProfile.ID,
		BranchID:        memberProfile.BranchID,
		OrganizationID:  memberProfile.OrganizationID,
	})
	if err != nil {
		return eris.Wrap(err, "failed to find member contact references")
	}
	for _, memberContact := range memberContacts {
		if err := MemberContactReferenceManager(service).DeleteWithTx(ctx, tx, memberContact.ID); err != nil {
			return eris.Wrapf(err, "failed to delete member contact reference: %s", memberContact.ID)
		}
	}
	if err := MemberProfileDelete(ctx, service, tx, memberProfile.ID); err != nil {
		return eris.Wrapf(err, "failed to delete member profile: %s", memberProfile.ID)
	}
	return err
}
