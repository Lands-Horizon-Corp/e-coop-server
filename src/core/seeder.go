package core

import (
	"context"
	crand "crypto/rand"
	"fmt"
	"io/fs"
	"math/big"
	"path/filepath"
	"strings"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/rotisserie/eris"
)

const country_code = "PH"

type City struct {
	Lat float64
	Lng float64
}

var cities = []City{
	{14.5995, 120.9842},
	{10.3157, 123.8854},
	{7.1907, 125.4553},
	{13.4125, 122.5621},
	{11.2400, 125.0055},
}

func createImageMedia(ctx context.Context, service *horizon.HorizonService, imagePaths []string, imageType string) (*types.Media, error) {
	if len(imagePaths) == 0 {
		return nil, eris.New("no image files available for seeding")
	}
	maxInt := big.NewInt(int64(len(imagePaths)))
	nBig, err := crand.Int(crand.Reader, maxInt)
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

	if err := MediaManager(service).Create(ctx, media); err != nil {
		return nil, eris.Wrap(err, "failed to create media record")
	}
	return media, nil
}

func Seed(ctx context.Context, service *horizon.HorizonService, multiplier int32) error {
	if multiplier <= 0 {
		return nil
	}
	imagePaths := []string{}
	imagesDir := "seeder/images"
	supportedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}
	err := filepath.WalkDir(imagesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if supportedExtensions[ext] {
				imagePaths = append(imagePaths, path)
			}
		}
		return nil
	})
	if err != nil {
		return eris.Wrap(err, "failed to scan images directory")
	}
	if len(imagePaths) == 0 {
		return eris.New("no image files found in seeder/images directory")
	}
	if err := GlobalSeeder(ctx, service); err != nil {
		return err
	}
	if err := SeedVALDECO(ctx, service); err != nil {
		return err
	}
	return nil
}

func SeedVALDECO(ctx context.Context, service *horizon.HorizonService) error {
	orgs, err := OrganizationManager(service).List(ctx)
	if err != nil {
		return err
	}
	if len(orgs) > 0 {
		return nil
	}
	var owner *types.User
	basePassword := "admin@valdeco123"
	hashedPassword, err := service.Security.HashPassword(basePassword)
	if err != nil {
		return eris.Wrap(err, "failed to hash password for admin user")
	}

	birthdate := time.Date(1985, time.January, 1, 0, 0, 0, 0, time.UTC)
	logoPath := "seeder/images/valdeco-logo.png"
	userMedia, err := createImageMedia(ctx, service, []string{logoPath}, "User Profile")
	if err != nil {
		return eris.Wrap(err, "failed to create admin user media")
	}
	owner = &types.User{
		MediaID:           &userMedia.ID,
		Email:             "admin@valdeco.com",
		Password:          hashedPassword,
		Birthdate:         &birthdate,
		Username:          "valdeco_admin",
		FullName:          "VALDECO System Administrator",
		FirstName:         helpers.Ptr("VALDECO"),
		MiddleName:        helpers.Ptr("S"),
		LastName:          helpers.Ptr("Administrator"),
		Suffix:            nil,
		ContactNumber:     "+63 925 511 5772",
		IsEmailVerified:   true,
		IsContactVerified: true,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	if err := UserManager(service).Create(ctx, owner); err != nil {
		return eris.Wrap(err, "failed to create admin user")
	}

	subscriptions, err := SubscriptionPlanManager(service).List(ctx)
	if err != nil {
		return err
	}
	if len(subscriptions) == 0 {
		return eris.New("no subscription plan found")
	}
	sub := subscriptions[0]

	categories, err := CategoryManager(service).List(ctx)
	if err != nil {
		return err
	}
	currency, err := CurrencyFindByAlpha2(ctx, service, "PH")
	if err != nil {
		return eris.Wrap(err, "failed to find PHP currency")
	}
	profilePath := "seeder/images/valdeco-profile.png"

	logoMedia, err := createImageMedia(ctx, service, []string{logoPath}, "Organization Logo")
	if err != nil {
		return eris.Wrap(err, "failed to upload logo")
	}
	profileMedia, err := createImageMedia(ctx, service, []string{profilePath}, "Organization Profile")
	if err != nil {
		return eris.Wrap(err, "failed to upload profile image")
	}
	subscriptionEndDate := time.Now().Add(30 * 24 * time.Hour)
	organization := &types.Organization{
		CreatedAt:                           time.Now().UTC(),
		CreatedByID:                         owner.ID,
		UpdatedAt:                           time.Now().UTC(),
		UpdatedByID:                         owner.ID,
		Name:                                "Valenzuela Development Cooperative (VALDECO)",
		Address:                             helpers.Ptr("VALDECO Bldg., Greenleaf Market, Tangke St., Malinta, Valenzuela, Philippines"),
		Email:                               helpers.Ptr("valenzueladevelopmentcoop@gmail.com"),
		ContactNumber:                       helpers.Ptr("+63 925 511 5772"),
		Description:                         helpers.Ptr(`I. SPIRITUAL AREA In this PANDEMIC crisis where you can’t see who is your real enemy, we the VALDECO Family believes that GOD is in control… "Be still in the presence of the LORD, and wait patiently for him to act" Psalm 37:7. We continuously doing our 30 minutes daily devotion to meditate the Word of GOD asking for strength and guidance. In this activity we include the Prayer Request of our members who visited in our office and wrote thru our "FREE PRAYER BOARD PROGRAM" their prayer request while observing SOCIAL DISTANCING…`),
		Color:                               helpers.Ptr("#0066cc"),
		TermsAndConditions:                  helpers.Ptr("Standard cooperative terms and conditions apply."),
		PrivacyPolicy:                       helpers.Ptr("VALDECO respects member privacy and complies with Data Privacy Act."),
		CookiePolicy:                        helpers.Ptr("This site uses cookies for essential functionality."),
		RefundPolicy:                        helpers.Ptr("Refunds are processed according to cooperative bylaws."),
		UserAgreement:                       helpers.Ptr("By using VALDECO services you agree to the cooperative's rules."),
		IsPrivate:                           false,
		MediaID:                             &logoMedia.ID,
		CoverMediaID:                        &profileMedia.ID,
		SubscriptionPlanMaxBranches:         sub.MaxBranches,
		SubscriptionPlanMaxEmployees:        sub.MaxEmployees,
		SubscriptionPlanMaxMembersPerBranch: sub.MaxMembersPerBranch,
		SubscriptionPlanID:                  &sub.ID,
		SubscriptionStartDate:               time.Now().UTC(),
		SubscriptionEndDate:                 subscriptionEndDate,
		InstagramLink:                       helpers.Ptr("https://instagram.com/valdecocoop"),
		FacebookLink:                        helpers.Ptr("https://facebook.com/valdecocoop"),
		YoutubeLink:                         helpers.Ptr("https://youtube.com/valdecocoop"),
		PersonalWebsiteLink:                 helpers.Ptr("https://valdecocoop.com"),
		XLink:                               helpers.Ptr("https://twitter.com/valdecocoop"),
	}

	if err := OrganizationManager(service).Create(ctx, organization); err != nil {
		return eris.Wrap(err, "failed to create VALDECO organization")
	}
	for _, category := range categories {
		if err := OrganizationCategoryManager(service).Create(ctx, &types.OrganizationCategory{
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
			OrganizationID: &organization.ID,
			CategoryID:     &category.ID,
		}); err != nil {
			return eris.Wrap(err, "failed to link organization to category")
		}
	}
	branches := []struct {
		name       string
		type_      string
		email      string
		address    string
		city       string
		region     string
		barangay   string
		postalCode string
		contact    string
		lat        float64
		lng        float64
		taxID      string
	}{
		{
			name:       "VALDECO Main Office",
			type_:      "main",
			email:      "valenzueladevelopmentcoop@gmail.com",
			address:    "VALDECO Bldg., Greenleaf Market, Tangke St., Malinta, Valenzuela, Philippines",
			city:       "Valenzuela",
			region:     "Metro Manila",
			barangay:   "Malinta",
			postalCode: "1440",
			contact:    "+63 925 511 5772",
			lat:        14.6995,
			lng:        120.9842,
			taxID:      "123456789",
		},
		{
			name:       "VALDECO Malabon Branch",
			type_:      "branch",
			email:      "emmydoy212324@gmail.com",
			address:    "189 Gen. Luna St, Malabon, Philippines",
			city:       "Malabon",
			region:     "Metro Manila",
			barangay:   "Barangay 1",
			postalCode: "1470",
			contact:    "+63 922 234 1493",
			lat:        14.6565,
			lng:        120.9482,
			taxID:      "987654321",
		},
	}

	for idx, br := range branches {
		branchMedia, err := createImageMedia(ctx, service, []string{logoPath}, "Branch Logo")
		if err != nil {
			return eris.Wrap(err, "failed to upload branch image")
		}
		branch := &types.Branch{
			CreatedAt:               time.Now().UTC(),
			CreatedByID:             owner.ID,
			UpdatedAt:               time.Now().UTC(),
			UpdatedByID:             owner.ID,
			OrganizationID:          organization.ID,
			Type:                    br.type_,
			Name:                    br.name,
			Email:                   br.email,
			Address:                 br.address,
			Province:                br.region,
			City:                    br.city,
			Region:                  br.region,
			Barangay:                br.barangay,
			PostalCode:              br.postalCode,
			CurrencyID:              &currency.ID,
			ContactNumber:           helpers.Ptr(br.contact),
			MediaID:                 &branchMedia.ID,
			Latitude:                &br.lat,
			Longitude:               &br.lng,
			TaxIdentificationNumber: helpers.Ptr(br.taxID),
		}

		if err := BranchManager(service).Create(ctx, branch); err != nil {
			return eris.Wrapf(err, "failed to create branch %s", br.name)
		}
		branchSetting := &types.BranchSetting{
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			BranchID:  branch.ID,

			WithdrawAllowUserInput: true,
			WithdrawPrefix:         "VAL",
			WithdrawORStart:        1,
			WithdrawORCurrent:      1,
			WithdrawOREnd:          999999,
			WithdrawORIteration:    1,

			DefaultMemberTypeID:   nil,
			DefaultMemberGenderID: nil,
			CurrencyID:            currency.ID,
		}

		if err := BranchSettingManager(service).Create(ctx, branchSetting); err != nil {
			return eris.Wrap(err, "failed to create branch settings")
		}
		if idx == 0 {
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
				Description:            "Founder and owner of VALDECO",
				ApplicationDescription: "Owner of the cooperative",
				ApplicationStatus:      "accepted",
				DeveloperSecretKey:     developerKey + "-owner-horizon",
				PermissionName:         "Owner",
				PermissionDescription:  "Full administrative permissions over the cooperative",
				Permissions:            []string{"read", "write", "manage", "delete", "admin"},
				Status:                 types.UserOrganizationStatusOnline,
				LastOnlineAt:           time.Now().UTC(),
			}

			if err := UserOrganizationManager(service).Create(ctx, ownerOrganization); err != nil {
				return eris.Wrap(err, "failed to create owner association")
			}
			tx, endTx := service.Database.StartTransaction(ctx)
			if err := OrganizationSeeder(ctx, service, tx, owner.ID, organization.ID, branch.ID); err != nil {
				return endTx(err)
			}
			if err := endTx(nil); err != nil {
				return err
			}
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
				Code:           fmt.Sprintf("VALDECO-%s-%d", userType, idx+1),
				ExpirationDate: time.Now().UTC().Add(60 * 24 * time.Hour),
				MaxUse:         100,
				CurrentUse:     0,
				Description:    fmt.Sprintf("Invitation for %s of VALDECO %s", userType, br.name),
			}

			if err := InvitationCodeManager(service).Create(ctx, invitationCode); err != nil {
				return eris.Wrap(err, "failed to create invitation code")
			}
		}
	}
	return nil
}
