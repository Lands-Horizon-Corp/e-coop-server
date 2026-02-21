package seeder

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/db/core"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/rotisserie/eris"
)

func Seed(ctx context.Context, service *horizon.HorizonService) error {
	if err := core.GlobalSeeder(ctx, service); err != nil {
		return err
	}
	if err := SeedVALDECO(ctx, service); err != nil {
		return err
	}
	return nil
}

func createImageMedia(ctx context.Context, service *horizon.HorizonService, imagePaths []string, imageType string) (*types.Media, error) {
	if len(imagePaths) == 0 {
		return nil, eris.New("no image files available for seeding")
	}
	maxInt := big.NewInt(int64(len(imagePaths)))
	nBig, err := rand.Int(rand.Reader, maxInt)
	if err != nil {
		return nil, eris.Wrap(err, "failed to generate secure random index for image selection")
	}
	randomIndex := int(nBig.Int64())
	imagePath := imagePaths[randomIndex]

	storage, err := service.Storage.UploadFromPath(ctx, imagePath, func(_ int64, _ int64, _ *horizon.Storage) {})
	if err != nil {
		return nil, eris.Wrapf(err, "failed to upload image from path %s for %s", imagePath, imageType)
	}
	media := &types.Media{
		FileName:   storage.FileName,
		FileType:   storage.FileType,
		FileSize:   storage.FileSize,
		StorageKey: storage.StorageKey,
		BucketName: storage.BucketName,
		Status:     "completed",
		Progress:   100,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	if err := core.MediaManager(service).Create(ctx, media); err != nil {
		return nil, eris.Wrap(err, "failed to create media record")
	}
	return media, nil
}

func SeedOrganization(ctx context.Context, service *horizon.HorizonService, config types.OrganizationSeedConfig) error {
	orgs, err := core.OrganizationManager(service).List(ctx)
	if err != nil {
		return eris.Wrap(err, "")
	}
	if len(orgs) > 0 {
		return nil
	}
	hashedPassword, err := service.Security.HashPassword(config.AdminPassword)
	if err != nil {
		return eris.Wrap(err, "failed to hash password for admin user")
	}
	userMedia, err := createImageMedia(ctx, service, []string{config.AdminLogoPath}, "User Profile")
	if err != nil {
		return eris.Wrap(err, "failed to create admin user media")
	}
	owner := &types.User{
		MediaID:           &userMedia.ID,
		Email:             config.AdminEmail,
		Password:          hashedPassword,
		Birthdate:         &config.AdminBirthdate,
		Username:          config.AdminUsername,
		FullName:          config.AdminFullName,
		FirstName:         &config.AdminFirstName,
		MiddleName:        config.AdminMiddleName,
		LastName:          &config.AdminLastName,
		Suffix:            config.AdminSuffix,
		ContactNumber:     config.AdminContactNumber,
		IsEmailVerified:   true,
		IsContactVerified: true,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	if err := core.UserManager(service).Create(ctx, owner); err != nil {
		return eris.Wrap(err, "failed to create admin user")
	}
	subscriptions, err := core.SubscriptionPlanManager(service).List(ctx)
	if err != nil {
		return eris.Wrap(err, "")
	}
	currency, err := core.CurrencyFindByAlpha2(ctx, service, config.CurrencyAlpha2)
	if err != nil {
		return eris.Wrap(err, "failed to find currency")
	}
	sub := types.SubscriptionPlan{}
	for _, s := range subscriptions {
		if s.CurrencyID != nil && *s.CurrencyID == currency.ID {
			sub = *s
			break
		}
	}
	if len(subscriptions) == 0 {
		return eris.New("no subscription plan found")
	}
	categories, err := core.CategoryManager(service).List(ctx)
	if err != nil {
		return eris.Wrap(err, "")
	}
	logoMedia, err := createImageMedia(ctx, service, []string{config.OrgLogoPath}, "Organization Logo")
	if err != nil {
		return eris.Wrap(err, "failed to upload logo")
	}
	profileMedia, err := createImageMedia(ctx, service, []string{config.OrgProfilePath}, "Organization Profile")
	if err != nil {
		return eris.Wrap(err, "failed to upload profile image")
	}
	subscriptionEndDate := time.Now().AddDate(0, 0, config.SubscriptionDays)
	organization := &types.Organization{
		CreatedAt:                           time.Now().UTC(),
		CreatedByID:                         owner.ID,
		UpdatedAt:                           time.Now().UTC(),
		UpdatedByID:                         owner.ID,
		Name:                                config.OrgName,
		Address:                             config.OrgAddress,
		Email:                               config.OrgEmail,
		ContactNumber:                       config.OrgContactNumber,
		Description:                         config.OrgDescription,
		Color:                               config.OrgColor,
		TermsAndConditions:                  config.OrgTerms,
		PrivacyPolicy:                       config.OrgPrivacy,
		CookiePolicy:                        config.OrgCookie,
		RefundPolicy:                        config.OrgRefund,
		UserAgreement:                       config.OrgUserAgreement,
		IsPrivate:                           config.OrgIsPrivate,
		MediaID:                             &logoMedia.ID,
		CoverMediaID:                        &profileMedia.ID,
		SubscriptionPlanMaxBranches:         sub.MaxBranches,
		SubscriptionPlanMaxEmployees:        sub.MaxEmployees,
		SubscriptionPlanMaxMembersPerBranch: sub.MaxMembersPerBranch,
		SubscriptionPlanID:                  &sub.ID,
		SubscriptionStartDate:               time.Now().UTC(),
		SubscriptionEndDate:                 subscriptionEndDate,
		InstagramLink:                       config.OrgInstagram,
		FacebookLink:                        config.OrgFacebook,
		YoutubeLink:                         config.OrgYoutube,
		PersonalWebsiteLink:                 config.OrgPersonalWebsite,
		XLink:                               config.OrgXLink,
	}
	if err := core.OrganizationManager(service).Create(ctx, organization); err != nil {
		return eris.Wrap(err, "failed to create organization")
	}
	for _, entry := range config.SeminarEntries {
		media, err := createImageMedia(ctx, service, []string{entry.MediaPath}, entry.Name)
		if err != nil {
			return eris.Wrapf(err, "failed to upload seminar image for %s", entry.Name)
		}
		orgMedia := &types.OrganizationMedia{
			MediaID:        media.ID,
			Name:           entry.Name,
			Description:    &entry.Description,
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
			OrganizationID: organization.ID,
		}
		if err := core.OrganizationMediaManager(service).Create(ctx, orgMedia); err != nil {
			return eris.Wrapf(err, "failed to create organization media for seminar %s", entry.Name)
		}
	}
	for _, category := range categories {
		if err := core.OrganizationCategoryManager(service).Create(ctx, &types.OrganizationCategory{
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
			OrganizationID: &organization.ID,
			CategoryID:     &category.ID,
		}); err != nil {
			return eris.Wrap(err, "failed to link organization to category")
		}
	}
	for idx, br := range config.Branches {
		branchMedia, err := createImageMedia(ctx, service, []string{br.LogoPath}, "Branch Logo")
		if err != nil {
			return eris.Wrap(err, "failed to upload branch image")
		}
		branch := &types.Branch{
			CreatedAt:               time.Now().UTC(),
			CreatedByID:             owner.ID,
			UpdatedAt:               time.Now().UTC(),
			UpdatedByID:             owner.ID,
			OrganizationID:          organization.ID,
			Type:                    br.Type,
			Name:                    br.Name,
			Email:                   br.Email,
			Address:                 br.Address,
			Province:                br.Region,
			City:                    br.City,
			Region:                  br.Region,
			Barangay:                br.Barangay,
			PostalCode:              br.PostalCode,
			CurrencyID:              &currency.ID,
			ContactNumber:           &br.Contact,
			MediaID:                 &branchMedia.ID,
			Latitude:                &br.Latitude,
			Longitude:               &br.Longitude,
			TaxIdentificationNumber: &br.TaxID,
		}
		if err := core.BranchManager(service).Create(ctx, branch); err != nil {
			return eris.Wrapf(err, "failed to create branch %s", br.Name)
		}
		branchSetting := &types.BranchSetting{
			CreatedAt:              time.Now().UTC(),
			UpdatedAt:              time.Now().UTC(),
			BranchID:               branch.ID,
			WithdrawAllowUserInput: br.WithdrawAllowUserInput,
			WithdrawPrefix:         br.WithdrawPrefix,
			WithdrawORStart:        br.WithdrawORStart,
			WithdrawORCurrent:      br.WithdrawORCurrent,
			WithdrawOREnd:          br.WithdrawOREnd,
			WithdrawORIteration:    br.WithdrawORIteration,
			CurrencyID:             currency.ID,
		}
		if err := core.BranchSettingManager(service).Create(ctx, branchSetting); err != nil {
			return eris.Wrap(err, "failed to create branch settings")
		}
		developerKey, err := service.Security.GenerateUUIDv5(fmt.Sprintf("%s-%s-%s", owner.ID, organization.ID, branch.ID))
		if err != nil {
			return eris.Wrap(err, "failed to generate developer key")
		}
		ownerOrganization := &types.UserOrganization{
			CreatedAt:              time.Now().UTC(),
			CreatedByID:            owner.ID,
			UpdatedAt:              time.Now().UTC(),
			UpdatedByID:            owner.ID,
			BranchID:               &branch.ID,
			OrganizationID:         organization.ID,
			UserID:                 owner.ID,
			UserType:               types.UserOrganizationTypeOwner,
			Description:            "Founder and owner of the organization",
			ApplicationDescription: "Owner of the cooperative",
			ApplicationStatus:      "accepted",
			DeveloperSecretKey:     developerKey + "-owner-horizon",
			PermissionName:         "Owner",
			PermissionDescription:  "Full administrative permissions over the cooperative",
			Permissions:            []string{"read", "write", "manage", "delete", "admin"},
			Status:                 types.UserOrganizationStatusOnline,
			LastOnlineAt:           time.Now().UTC(),
		}
		if err := core.UserOrganizationManager(service).Create(ctx, ownerOrganization); err != nil {
			return eris.Wrap(err, "failed to create owner association")
		}
		tx, endTx := service.Database.StartTransaction(ctx)
		if err := core.OrganizationSeeder(ctx, service, tx, owner.ID, organization.ID, branch.ID); err != nil {
			return endTx(err)
		}
		if err := endTx(nil); err != nil {
			return err
		}
		invitationCodes := []types.UserOrganizationType{
			types.UserOrganizationTypeMember,
			types.UserOrganizationTypeEmployee,
		}
		for _, userType := range invitationCodes {
			invitationCode := &types.InvitationCode{
				CreatedAt:      time.Now().UTC(),
				CreatedByID:    owner.ID,
				UpdatedAt:      time.Now().UTC(),
				UpdatedByID:    owner.ID,
				OrganizationID: organization.ID,
				BranchID:       branch.ID,
				UserType:       userType,
				Code:           fmt.Sprintf("%s-%s-%d", config.OrgName, userType, idx+1),
				ExpirationDate: time.Now().UTC().Add(config.InvitationExpiration),
				MaxUse:         config.InvitationMaxUse,
				CurrentUse:     0,
				Description:    fmt.Sprintf("Invitation for %s of %s %s", userType, config.OrgName, br.Name),
			}
			if err := core.InvitationCodeManager(service).Create(ctx, invitationCode); err != nil {
				return eris.Wrap(err, "failed to create invitation code")
			}
		}
		feed := &types.Feed{
			Description:    "🎉 Welcome! Today the coop system has been created. Let's grow together!",
			CreatedAt:      time.Now().UTC(),
			CreatedByID:    owner.ID,
			UpdatedAt:      time.Now().UTC(),
			UpdatedByID:    owner.ID,
			BranchID:       branch.ID,
			OrganizationID: organization.ID,
		}
		if err := core.FeedManager(service).Create(ctx, feed); err != nil {
			return eris.Wrap(err, "Failed to create feed record")
		}
		if err := core.FeedMediaManager(service).Create(ctx, &types.FeedMedia{
			FeedID:         feed.ID,
			MediaID:        logoMedia.ID,
			OrganizationID: organization.ID,
			BranchID:       branch.ID,
			CreatedByID:    owner.ID,
			UpdatedByID:    owner.ID,
		}); err != nil {
			return eris.Wrap(err, "Failed to associate media with feed")
		}
	}

	return nil
}
