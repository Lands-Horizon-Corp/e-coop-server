package core

import (
	"context"
	crand "crypto/rand"
	"math/big"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/rotisserie/eris"
)

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

func Seed(ctx context.Context, service *horizon.HorizonService) error {
	if err := GlobalSeeder(ctx, service); err != nil {
		return err
	}
	if err := SeedVALDECO(ctx, service); err != nil {
		return err
	}
	return nil
}

func SeedVALDECO(ctx context.Context, service *horizon.HorizonService) error {
	config := types.OrganizationSeedConfig{
		AdminEmail:         "admin@valdeco.com",
		AdminPassword:      "admin@valdeco123",
		AdminBirthdate:     time.Date(1985, time.January, 1, 0, 0, 0, 0, time.UTC),
		AdminUsername:      "valdeco_admin",
		AdminFullName:      "VALDECO System Administrator",
		AdminFirstName:     "VALDECO",
		AdminMiddleName:    helpers.Ptr("S"),
		AdminLastName:      "Administrator",
		AdminSuffix:        nil,
		AdminContactNumber: "+63 925 511 5772",
		AdminLogoPath:      "seeder/images/valdeco-logo.png",
		OrgName:            "Valenzuela Development Cooperative (VALDECO)",
		OrgAddress:         helpers.Ptr("VALDECO Bldg., Greenleaf Market, Tangke St., Malinta, Valenzuela, Philippines"),
		OrgEmail:           helpers.Ptr("valenzueladevelopmentcoop@gmail.com"),
		OrgContactNumber:   helpers.Ptr("+63 925 511 5772"),
		OrgDescription:     helpers.Ptr(`I. SPIRITUAL AREA In this PANDEMIC crisis where you can’t see who is your real enemy, we the VALDECO Family believes that GOD is in control… "Be still in the presence of the LORD, and wait patiently for him to act" Psalm 37:7. We continuously doing our 30 minutes daily devotion to meditate the Word of GOD asking for strength and guidance. In this activity we include the Prayer Request of our members who visited in our office and wrote thru our "FREE PRAYER BOARD PROGRAM" their prayer request while observing SOCIAL DISTANCING…`),
		OrgColor:           helpers.Ptr("#0066cc"),
		OrgTerms:           helpers.Ptr("Standard cooperative terms and conditions apply."),
		OrgPrivacy:         helpers.Ptr("VALDECO respects member privacy and complies with Data Privacy Act."),
		OrgCookie:          helpers.Ptr("This site uses cookies for essential functionality."),
		OrgRefund:          helpers.Ptr("Refunds are processed according to cooperative bylaws."),
		OrgUserAgreement:   helpers.Ptr("By using VALDECO services you agree to the cooperative's rules."),
		OrgIsPrivate:       false,
		OrgLogoPath:        "seeder/images/valdeco-logo.png",
		OrgProfilePath:     "seeder/images/valdeco-profile.png",
		OrgInstagram:       helpers.Ptr("https://instagram.com/valdecocoop"),
		OrgFacebook:        helpers.Ptr("https://facebook.com/valdecocoop"),
		OrgYoutube:         helpers.Ptr("https://youtube.com/valdecocoop"),
		OrgPersonalWebsite: helpers.Ptr("https://valdecocoop.com"),
		OrgXLink:           helpers.Ptr("https://twitter.com/valdecocoop"),
		SeminarEntries: []types.SeminarEntry{
			{
				MediaPath:   "seeder/images/valdeco/ownership-seminar-2025.jpg",
				Name:        "VALDECO Ownership Seminar 2025",
				Description: "Calling all cooperative members! Join us for a focused and empowering Ownership Seminar on August 15, 2025, at 1:00 PM. 2F Multipurpose Hall, Valdeco Greenleaf Market, Tangke St., Valenzuela City. Deepen your understanding of ownership, discover strategies for sustainable growth, and connect with a thriving community working toward shared success. Don't miss this chance to invest in your future! Pre-Register now!",
			},
		},
		Branches: []types.BranchConfig{
			{
				Name:                   "VALDECO Main Office",
				Type:                   "main",
				Email:                  "valenzueladevelopmentcoop@gmail.com",
				Address:                "VALDECO Bldg., Greenleaf Market, Tangke St., Malinta, Valenzuela, Philippines",
				City:                   "Valenzuela",
				Region:                 "Metro Manila",
				Barangay:               "Malinta",
				PostalCode:             "1440",
				Contact:                "+63 925 511 5772",
				Latitude:               14.6995,
				Longitude:              120.9842,
				TaxID:                  "123456789",
				LogoPath:               "seeder/images/valdeco-logo.png",
				WithdrawAllowUserInput: true,
				WithdrawPrefix:         "VAL",
				WithdrawORStart:        1,
				WithdrawORCurrent:      1,
				WithdrawOREnd:          999999,
				WithdrawORIteration:    1,
			},
			{
				Name:       "VALDECO Malabon Branch",
				Type:       "branch",
				Email:      "emmydoy212324@gmail.com",
				Address:    "189 Gen. Luna St, Malabon, Philippines",
				City:       "Malabon",
				Region:     "Metro Manila",
				Barangay:   "Barangay 1",
				PostalCode: "1470",
				Contact:    "+63 922 234 1493",
				Latitude:   14.6565,
				Longitude:  120.9482,
				TaxID:      "987654321",
				LogoPath:   "seeder/images/valdeco-logo.png",

				WithdrawAllowUserInput: true,
				WithdrawPrefix:         "VAL",
				WithdrawORStart:        1,
				WithdrawORCurrent:      1,
				WithdrawOREnd:          999999,
				WithdrawORIteration:    1,
			},
		},
		CurrencyAlpha2:       "PH",
		SubscriptionDays:     30,
		InvitationMaxUse:     100,
		InvitationExpiration: 60 * 24 * time.Hour,
	}
	return SeedOrganization(ctx, service, config)
}
