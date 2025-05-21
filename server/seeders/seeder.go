package seeders

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"horizon.com/server/horizon"
	"horizon.com/server/server/model"
)

type DatabaseSeeder struct {
	database       *horizon.HorizonDatabase
	authentication *horizon.HorizonAuthentication
	storage        *horizon.HorizonStorage
	security       *horizon.HorizonSecurity

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
	memberCenter         *model.MemberCenterCollection
	memberClassification *model.MemberClassificationCollection
	memberGender         *model.MemberGenderCollection
	memberGroup          *model.MemberGroupCollection
	memberOccupation     *model.MemberOccupationCollection
	memberType           *model.MemberTypeCollection
	// End Maintenantce table member

}

func NewDatabaseSeeder(
	database *horizon.HorizonDatabase,
	authentication *horizon.HorizonAuthentication,
	storage *horizon.HorizonStorage,
	security *horizon.HorizonSecurity,

	// all collections
	branch *model.BranchCollection, // ✅
	category *model.CategoryCollection, // ✅
	contactUs *model.ContactUsCollection, // ✅
	feedback *model.FeedbackCollection, // ✅
	footstep *model.FootstepCollection,
	generatedReport *model.GeneratedReportCollection,
	invitationCode *model.InvitationCodeCollection, // ✅
	media *model.MediaCollection, // ✅
	notification *model.NotificationCollection, // ✅
	organizationCategory *model.OrganizationCategoryCollection, // ✅
	organizationDailyUsage *model.OrganizationDailyUsageCollection,
	organization *model.OrganizationCollection, // ✅
	permissionTemplate *model.PermissionTemplateCollection,
	subscriptionPlan *model.SubscriptionPlanCollection, // ✅
	userOrganization *model.UserOrganizationCollection, // ✅
	user *model.UserCollection, // ✅
	userRating *model.UserRatingCollection,

	// Maintenantce Table member
	memberCenter *model.MemberCenterCollection,
	memberClassification *model.MemberClassificationCollection,
	memberGender *model.MemberGenderCollection,
	memberGroup *model.MemberGroupCollection,
	memberOccupation *model.MemberOccupationCollection,
	memberType *model.MemberTypeCollection,
	// End Maintenantce table member

) (*DatabaseSeeder, error) {
	return &DatabaseSeeder{
		database:       database,
		authentication: authentication,
		storage:        storage,
		security:       security,

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
		userRating:             userRating,

		memberCenter:         memberCenter,
		memberClassification: memberClassification,
		memberGender:         memberGender,
		memberGroup:          memberGroup,
		memberOccupation:     memberOccupation,
		memberType:           memberType,
	}, nil
}

func (ds *DatabaseSeeder) Run() error {
	users, err := ds.user.Manager.List()
	if err != nil {
		return err
	}
	if len(users) <= 0 {
		if err := ds.SeedCategories(); err != nil {
			return err
		}
		fmt.Println("finished seeding categories")
		if err := ds.SeedContactUs(); err != nil {
			return err
		}
		fmt.Println("finished seeding contact us")
		if err := ds.SeedFeedback(); err != nil {
			return err
		}
		fmt.Println("finished seeding feedback")
		if err := ds.SeedUser(); err != nil {
			return err
		}
		fmt.Println("finished seeding user")
		if err := ds.SeedNotification(); err != nil {
			return err
		}
		fmt.Println("finished seeding notification")
		if err := ds.SeedSubscriptionPlan(); err != nil {
			return err
		}
		fmt.Println("finished seeding subscription")
		if err := ds.SeedOrganization(); err != nil {
			return err
		}
		fmt.Println("finished seeding organization")
	}
	return nil
}

func (ds *DatabaseSeeder) SeedOrganization() error {
	users, err := ds.user.Manager.List()
	if err != nil {
		return err
	}

	// Ensure we have at least 5 users for our 5 organizations
	const numOrgsPerUser = 5

	subscriptions, err := ds.subscriptionPlan.Manager.List()
	if err != nil {
		return err
	}
	categories, err := ds.category.Manager.List()
	if err != nil {
		return err
	}

	imageUrl := "https://files.slack.com/files-tmb/T08P9M6T257-F08S7UMSZ0V-e535d7688e/image_720.png"
	image, err := ds.storage.UploadFromURL(imageUrl, func(progress, total int64, storage *horizon.Storage) {})
	if err != nil {
		return err
	}

	media := &model.Media{
		FileName:   "organization_image",
		FileType:   "image/png",
		FileSize:   image.FileSize,
		StorageKey: image.StorageKey,
		URL:        image.URL,
		BucketName: image.BucketName,
		Status:     horizon.StorageStatusCompleted,
		Progress:   100,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	if err := ds.media.Manager.Create(media); err != nil {
		return err
	}

	for _, user := range users {
		// Define user types for each organization
		orgUserTypes := []string{"member", "owner", "employee", "owner", "employee"}

		for j := 0; j < numOrgsPerUser; j++ {
			sub := subscriptions[j%len(subscriptions)]
			subscriptionEndDate := time.Now().Add(30 * 24 * time.Hour)

			i := j + 1 // Adjust index to start from 1
			organization := &model.Organization{
				CreatedAt:                           time.Now().UTC(),
				CreatedByID:                         user.ID,
				UpdatedAt:                           time.Now().UTC(),
				UpdatedByID:                         user.ID,
				Name:                                fmt.Sprintf("Org %d - %s", i, *user.FirstName),
				Address:                             ptr(fmt.Sprintf("%d Main Street, Testville", i+101)),
				Email:                               ptr(fmt.Sprintf("org%d@example.com", i)),
				ContactNumber:                       ptr(fmt.Sprintf("+63917%05d%04d", i+100, i)),
				Description:                         ptr("A seeded example organization for testing."),
				Color:                               ptr("#" + fmt.Sprintf("%06x", 0xFF5733+i*50)),
				TermsAndConditions:                  ptr("These are seeded terms and conditions..."),
				PrivacyPolicy:                       ptr("Seeded privacy policy content..."),
				CookiePolicy:                        ptr("Seeded cookie policy content..."),
				RefundPolicy:                        ptr("Seeded refund policy content..."),
				UserAgreement:                       ptr("Seeded user agreement content..."),
				IsPrivate:                           false,
				MediaID:                             &media.ID,
				CoverMediaID:                        &media.ID,
				SubscriptionPlanMaxBranches:         sub.MaxBranches,
				SubscriptionPlanMaxEmployees:        sub.MaxEmployees,
				SubscriptionPlanMaxMembersPerBranch: sub.MaxMembersPerBranch,
				SubscriptionPlanID:                  &sub.ID,
				SubscriptionStartDate:               time.Now().UTC(),
				SubscriptionEndDate:                 subscriptionEndDate,
			}
			if err := ds.organization.Manager.Create(organization); err != nil {
				return err
			}

			// Add categories
			for _, category := range categories {
				if err := ds.organizationCategory.Manager.Create(&model.OrganizationCategory{
					CreatedAt:      time.Now().UTC(),
					UpdatedAt:      time.Now().UTC(),
					OrganizationID: &organization.ID,
					CategoryID:     &category.ID,
				}); err != nil {
					return err
				}
			}

			// Get user type for this organization
			orgUserType := orgUserTypes[j]

			// Create 3 branches for each organization
			for k := range 3 {
				branch := &model.Branch{
					CreatedAt:      time.Now().UTC(),
					CreatedByID:    user.ID,
					UpdatedAt:      time.Now().UTC(),
					UpdatedByID:    user.ID,
					OrganizationID: organization.ID,
					Type:           []string{"main", "satellite"}[k%2],
					Name:           fmt.Sprintf("Branch %d - %s", k+1, organization.Name),
					Email:          fmt.Sprintf("branch%d.%d@organization.com", i, k+1),
					Address:        fmt.Sprintf("Branch %d Street, City", k+1),
					Province:       "Test Province",
					City:           "Test City",
					Region:         "Test Region",
					Barangay:       "Test Barangay",
					PostalCode:     fmt.Sprintf("11%03d", k+1),
					CountryCode:    "PH",
					ContactNumber:  ptr(fmt.Sprintf("+63918%05d%04d", k+1, i)),
					MediaID:        &media.ID,
				}
				if err := ds.branch.Manager.Create(branch); err != nil {
					return err
				}

				// Create user organization relationship with specific type
				userOrganization := &model.UserOrganization{
					CreatedAt:              time.Now().UTC(),
					CreatedByID:            user.ID,
					UpdatedAt:              time.Now().UTC(),
					UpdatedByID:            user.ID,
					BranchID:               &branch.ID,
					OrganizationID:         organization.ID,
					UserID:                 user.ID,
					UserType:               orgUserType,
					Description:            fmt.Sprintf("User %s added as %s", *user.FirstName, orgUserType),
					ApplicationDescription: "Seeded application for testing",
					ApplicationStatus:      "accepted",
					DeveloperSecretKey:     ds.security.GenerateToken(user.ID.String()),
					PermissionName:         orgUserType,
					PermissionDescription:  "Auto-generated role assignment",
					Permissions:            []string{"read", "write", "manage"},
				}
				if err := ds.userOrganization.Manager.Create(userOrganization); err != nil {
					return err
				}

				// Seed related data
				ds.memberCenter.Seeder(user.ID, organization.ID, branch.ID)
				ds.memberClassification.Seeder(user.ID, organization.ID, branch.ID)
				ds.memberGender.Seeder(user.ID, organization.ID, branch.ID)
				ds.memberGroup.Seeder(user.ID, organization.ID, branch.ID)
				ds.memberOccupation.Seeder(user.ID, organization.ID, branch.ID)
				ds.memberType.Seeder(user.ID, organization.ID, branch.ID)

				// Create invitation codes
				for m := range 5 {
					invitationCode := &model.InvitationCode{
						CreatedAt:      time.Now().UTC(),
						CreatedByID:    user.ID,
						UpdatedAt:      time.Now().UTC(),
						UpdatedByID:    user.ID,
						OrganizationID: organization.ID,
						BranchID:       branch.ID,
						UserType:       "user",
						Code:           uuid.New().String(),
						ExpirationDate: time.Now().UTC().Add(60 * 24 * time.Hour),
						MaxUse:         50,
						CurrentUse:     m % 25,
						Description:    fmt.Sprintf("Invite for Branch %d, user %d", k+1, i),
					}
					if err := ds.invitationCode.Manager.Create(invitationCode); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (ds *DatabaseSeeder) SeedUser() error {
	imageUrl := "https://files.slack.com/files-tmb/T08P9M6T257-F08S7UMSZ0V-e535d7688e/image_720.png"
	image, err := ds.storage.Upload(imageUrl, func(progress, total int64, storage *horizon.Storage) {})
	if err != nil {
		return err
	}
	media := &model.Media{
		FileName:   "picture",
		FileType:   "image/png",
		FileSize:   image.FileSize,
		StorageKey: image.StorageKey,
		URL:        image.URL,
		BucketName: image.BucketName,
		Status:     horizon.StorageStatusCompleted,
		Progress:   100,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	if err := ds.media.Manager.Create(media); err != nil {
		return err
	}
	hashedPwd, err := ds.authentication.Password("sample-hello-world-12345")
	if err != nil {
		return err
	}
	user1 := &model.User{
		MediaID:           &media.ID,
		Email:             "sample@example.com",
		Password:          hashedPwd,
		Birthdate:         time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
		UserName:          "sampleuser",
		FullName:          ptr("Sample User"),
		FirstName:         ptr("Sample"),
		MiddleName:        ptr("T."),
		LastName:          ptr("User"),
		Suffix:            ptr("J"),
		ContactNumber:     "+639123456789",
		IsEmailVerified:   true,
		IsContactVerified: true,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	if err := ds.user.Manager.Create(user1); err != nil {
		return err
	}
	user2 := &model.User{
		MediaID:           &media.ID,
		Email:             "user2@example.com",
		Password:          hashedPwd,
		Birthdate:         time.Date(1988, time.July, 15, 0, 0, 0, 0, time.UTC),
		UserName:          "usertwo",
		FullName:          ptr("User Two"),
		FirstName:         ptr("User"),
		MiddleName:        ptr("B."),
		LastName:          ptr("Two"),
		Suffix:            ptr("Jr"),
		ContactNumber:     "+639222222222",
		IsEmailVerified:   true,
		IsContactVerified: true,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	if err := ds.user.Manager.Create(user2); err != nil {
		return err
	}

	return nil
}
func (ds *DatabaseSeeder) SeedCategories() error {
	categories := []model.Category{
		{
			Name:        "Loaning",
			Description: "Loan-related cooperative services",
			Color:       "#FF5733",
			Icon:        "loan",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Membership",
			Description: "Member registration and benefits",
			Color:       "#33C1FF",
			Icon:        "user-group",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Team Building",
			Description: "Events and programs to strengthen teamwork",
			Color:       "#33FF6F",
			Icon:        "team",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Farming",
			Description: "Agricultural and farming initiatives",
			Color:       "#A3D633",
			Icon:        "tractor",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Technology",
			Description: "Tech support and infrastructure",
			Color:       "#8E44AD",
			Icon:        "chip",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Education",
			Description: "Training and educational programs",
			Color:       "#FFC300",
			Icon:        "book-open",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
		{
			Name:        "Livelihood",
			Description: "Community livelihood support",
			Color:       "#2ECC71",
			Icon:        "briefcase",
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		},
	}

	for _, category := range categories {
		if err := ds.category.Manager.Create(&category); err != nil {
			return err
		}
	}
	return nil
}
func (ds *DatabaseSeeder) SeedContactUs() error {
	contacts := []model.ContactUs{
		{
			FirstName:     "Juan",
			LastName:      "Dela Cruz",
			Email:         "juan@example.com",
			ContactNumber: "+639101112131",
			Description:   "Inquiry about membership benefits.",
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		},
		{
			FirstName:     "Maria",
			LastName:      "Santos",
			Email:         "maria@example.com",
			ContactNumber: "+639202122232",
			Description:   "Looking for information on loaning services.",
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		},
		{
			FirstName:     "Jose",
			LastName:      "Rizal",
			Email:         "jose.rizal@example.com",
			ContactNumber: "+639303132333",
			Description:   "Feedback regarding recent farming initiative.",
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		},
	}

	for _, contact := range contacts {
		if err := ds.contactUs.Manager.Create(&contact); err != nil {
			return err // optionally log and continues
		}
	}
	return nil
}
func (ds *DatabaseSeeder) SeedFeedback() error {
	imageUrl := "https://files.slack.com/files-tmb/T08P9M6T257-F08S7UMSZ0V-e535d7688e/image_720.png"
	image, err := ds.storage.Upload(imageUrl, func(progress, total int64, storage *horizon.Storage) {})
	if err != nil {
		return err
	}
	media := &model.Media{
		FileName:   "picture",
		FileType:   "image/png",
		FileSize:   image.FileSize,
		StorageKey: image.StorageKey,
		URL:        image.URL,
		BucketName: image.BucketName,
		Status:     horizon.StorageStatusCompleted,
		Progress:   100,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	if err := ds.media.Manager.Create(media); err != nil {
		return err
	}
	feedbacks := []model.Feedback{
		{
			MediaID:      &media.ID,
			Email:        "feedback1@example.com",
			Description:  "Great service, very helpful staff!",
			FeedbackType: "general",
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		},
		{
			MediaID:      &media.ID,
			Email:        "feedback2@example.com",
			Description:  "Please improve your loan application turnaround time.",
			FeedbackType: "complaint",
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		},
	}

	for _, f := range feedbacks {
		if err := ds.feedback.Manager.Create(&f); err != nil {
			return err
		}
	}

	return nil
}
func (ds *DatabaseSeeder) SeedNotification() error {

	users, err := ds.user.Manager.List()
	if err != nil {
		return err
	}

	if len(users) < 2 {
		return fmt.Errorf("expected at least 2 users to seed notifications")
	}

	notifications := []model.Notification{
		{
			UserID:           users[0].ID,
			Title:            "Welcome to the platform!",
			Description:      "Hello Sample User, we’re glad to have you onboard.",
			NotificationType: "welcome",
			IsViewed:         false,
			CreatedAt:        time.Now().UTC(),
			UpdatedAt:        time.Now().UTC(),
		},
		{
			UserID:           users[0].ID,
			Title:            "Your profile is complete",
			Description:      "Thanks for completing your profile, Sample.",
			NotificationType: "info",
			IsViewed:         false,
			CreatedAt:        time.Now().UTC(),
			UpdatedAt:        time.Now().UTC(),
		},
		{
			UserID:           users[1].ID,
			Title:            "New features available",
			Description:      "User Two, check out the latest updates we just released!",
			NotificationType: "update",
			IsViewed:         false,
			CreatedAt:        time.Now().UTC(),
			UpdatedAt:        time.Now().UTC(),
		},
		{
			UserID:           users[1].ID,
			Title:            "Action Required",
			Description:      "Please verify your additional contact details.",
			NotificationType: "alert",
			IsViewed:         false,
			CreatedAt:        time.Now().UTC(),
			UpdatedAt:        time.Now().UTC(),
		},
	}

	for _, n := range notifications {
		if err := ds.notification.Manager.Create(&n); err != nil {
			return err
		}
	}

	return nil
}
func (ds *DatabaseSeeder) SeedSubscriptionPlan() error {
	// Sample subscription plans
	subscriptionPlans := []model.SubscriptionPlan{
		{
			Name:                "Basic Plan",
			Description:         "A basic plan with limited features.",
			Cost:                99.99,
			Timespan:            12, // 12 months
			MaxBranches:         5,
			MaxEmployees:        50,
			MaxMembersPerBranch: 5,
			Discount:            5.00,  // 5% discount
			YearlyDiscount:      10.00, // 10% yearly discount
		},
		{
			Name:                "Pro Plan",
			Description:         "A professional plan with additional features.",
			Cost:                199.99,
			Timespan:            12, // 12 months
			MaxBranches:         10,
			MaxEmployees:        100,
			MaxMembersPerBranch: 10,
			Discount:            10.00, // 10% discount
			YearlyDiscount:      15.00, // 15% yearly discount
		},
		{
			Name:                "Enterprise Plan",
			Description:         "An enterprise-level plan with unlimited features.",
			Cost:                499.99,
			Timespan:            12, // 12 months
			MaxBranches:         20,
			MaxEmployees:        500,
			MaxMembersPerBranch: 50,
			Discount:            15.00, // 15% discount
			YearlyDiscount:      20.00, // 20% yearly discount
		},
	}
	for _, subscriptionPlan := range subscriptionPlans {
		if err := ds.subscriptionPlan.Manager.Create(&subscriptionPlan); err != nil {
			return err // optionally log and continue
		}
	}
	return nil
}

func ptr[T any](v T) *T {
	return &v
}
