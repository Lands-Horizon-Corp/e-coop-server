package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model"
	"github.com/google/uuid"
)

type Seeder struct {
	provider *src.Provider
	model    *model.Model
}

func NewSeeder(provider *src.Provider, model *model.Model) (*Seeder, error) {
	return &Seeder{
		provider: provider,
		model:    model,
	}, nil
}

func (s *Seeder) Run(ctx context.Context) error {
	s.provider.Service.Logger.Info("Starting database seeding...")
	if err := s.SeedSubscription(ctx); err != nil {
		return err
	}
	if err := s.SeedCategory(ctx); err != nil {
		return err
	}
	if err := s.SeedUsers(ctx); err != nil {
		return err
	}
	if err := s.SeedOrganization(ctx); err != nil {
		return err
	}
	if err := s.SeedEmployees(ctx); err != nil {
		return err
	}
	if err := s.SeedMemberProfiles(ctx); err != nil {
		return err
	}
	s.provider.Service.Logger.Info("Seeding completed successfully.")
	return nil
}

func (s *Seeder) SeedCategory(ctx context.Context) error {
	category, err := s.model.CategoryManager.List(ctx)
	if err != nil {
		return err
	}
	if len(category) >= 1 {
		return nil
	}

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
		if err := s.model.CategoryManager.Create(ctx, &category); err != nil {
			return err
		}
	}
	return nil
}

func (s *Seeder) SeedSubscription(ctx context.Context) error {
	subscriptionPlan, err := s.model.SubscriptionPlanManager.List(ctx)
	if err != nil {
		return err
	}
	if len(subscriptionPlan) >= 1 {
		return nil
	}
	subscriptionPlans := []model.SubscriptionPlan{
		{
			Name:                "Enterprise Plan",
			Description:         "An enterprise-level plan with unlimited features.",
			Cost:                499.99,
			Timespan:            int64(30 * 24 * time.Hour),
			MaxBranches:         20,
			MaxEmployees:        500,
			MaxMembersPerBranch: 50,
			Discount:            15.00, // 15% discount
			YearlyDiscount:      20.00, // 20% yearly discount
			IsRecommended:       false, // set as needed
		},
		{
			Name:                "Pro Plan",
			Description:         "A professional plan with additional features.",
			Cost:                199.99,
			Timespan:            int64(30 * 24 * time.Hour),
			MaxBranches:         10,
			MaxEmployees:        100,
			MaxMembersPerBranch: 10,
			Discount:            10.00, // 10% discount
			YearlyDiscount:      15.00, // 15% yearly discount
			IsRecommended:       false, // set as needed
		},
		{
			Name:                "Starter Plan",
			Description:         "An affordable plan for small organizations just getting started.",
			Cost:                49.99,
			Timespan:            int64(30 * 24 * time.Hour),
			MaxBranches:         2,
			MaxEmployees:        10,
			MaxMembersPerBranch: 2,
			Discount:            2.50, // 2.5% discount
			YearlyDiscount:      5.00, // 5% yearly discount
			IsRecommended:       true, // set as needed
		},
		{
			Name:                "Free Plan",
			Description:         "A basic plan with limited features.",
			Cost:                0.00,
			Timespan:            int64(14 * 24 * time.Hour), // 14 days
			MaxBranches:         1,
			MaxEmployees:        2,
			MaxMembersPerBranch: 5,
			Discount:            0, YearlyDiscount: 0,
			IsRecommended: false,
		},
	}
	for _, subscriptionPlan := range subscriptionPlans {
		if err := s.model.SubscriptionPlanManager.Create(ctx, &subscriptionPlan); err != nil {
			return err // optionally log and continue
		}
	}
	return nil
}

func (ds *Seeder) SeedOrganization(ctx context.Context) error {
	orgs, err := ds.model.OrganizationManager.List(ctx)
	if err != nil {
		return err
	}
	if len(orgs) > 0 {
		return nil
	}

	const numOrgsPerUser = 2 // Each user will own 2 organizations
	users, err := ds.model.UserManager.List(ctx)
	if err != nil {
		return err
	}
	subscriptions, err := ds.model.SubscriptionPlanManager.List(ctx)
	if err != nil {
		return err
	}
	categories, err := ds.model.CategoryManager.List(ctx)
	if err != nil {
		return err
	}
	imageUrl := "https://yavuzceliker.github.io/sample-images/image-1021.jpg"
	image, err := ds.provider.Service.Storage.UploadFromURL(ctx, imageUrl, func(progress, total int64, storage *horizon.Storage) {})
	if err != nil {
		return err
	}
	media := &model.Media{
		FileName:   image.FileName,
		FileType:   image.FileType,
		FileSize:   image.FileSize,
		StorageKey: image.StorageKey,
		URL:        image.URL,
		BucketName: image.BucketName,
		Status:     "comppleted",
		Progress:   100,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	if err := ds.model.MediaManager.Create(ctx, media); err != nil {
		return err
	}

	// Create organizations for each user (each user owns their organizations)
	for userIndex, user := range users {
		for j := 0; j < numOrgsPerUser; j++ {
			sub := subscriptions[j%len(subscriptions)]
			subscriptionEndDate := time.Now().Add(30 * 24 * time.Hour)

			organization := &model.Organization{
				CreatedAt:                           time.Now().UTC(),
				CreatedByID:                         user.ID,
				UpdatedAt:                           time.Now().UTC(),
				UpdatedByID:                         user.ID,
				Name:                                fmt.Sprintf("%s's Organization %d", *user.FirstName, j+1),
				Address:                             ptr(fmt.Sprintf("%d %s Street, City %d", (userIndex*10)+j+101, *user.FirstName, userIndex+1)),
				Email:                               ptr(fmt.Sprintf("org%d.%d@%s.com", userIndex+1, j+1, *user.FirstName)),
				ContactNumber:                       ptr(fmt.Sprintf("+63917%02d%02d%04d", userIndex+1, j+1, (userIndex*100)+j)),
				Description:                         ptr(fmt.Sprintf("Organization owned by %s %s for testing.", *user.FirstName, *user.LastName)),
				Color:                               ptr("#" + fmt.Sprintf("%06x", 0xFF5733+(userIndex*100)+(j*50))),
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
			if err := ds.model.OrganizationManager.Create(ctx, organization); err != nil {
				return err
			}

			// Add categories to organization
			for _, category := range categories {
				if err := ds.model.OrganizationCategoryManager.Create(ctx, &model.OrganizationCategory{
					CreatedAt:      time.Now().UTC(),
					UpdatedAt:      time.Now().UTC(),
					OrganizationID: &organization.ID,
					CategoryID:     &category.ID,
				}); err != nil {
					return err
				}
			}

			// Create 2-3 branches for each organization
			numBranches := 2 + (j % 2) // 2-3 branches per organization
			for k := 0; k < numBranches; k++ {
				branch := &model.Branch{
					CreatedAt:      time.Now().UTC(),
					CreatedByID:    user.ID,
					UpdatedAt:      time.Now().UTC(),
					UpdatedByID:    user.ID,
					OrganizationID: organization.ID,
					Type:           []string{"main", "satellite", "branch"}[k%3],
					Name:           fmt.Sprintf("%s Branch %d", organization.Name, k+1),
					Email:          fmt.Sprintf("branch%d@%s-org%d.com", k+1, *user.FirstName, j+1),
					Address:        fmt.Sprintf("Branch %d, %s Street, City %d", k+1, *user.FirstName, userIndex+1),
					Province:       "Test Province",
					City:           fmt.Sprintf("Test City %d", userIndex+1),
					Region:         "Test Region",
					Barangay:       fmt.Sprintf("Barangay %d", k+1),
					PostalCode:     fmt.Sprintf("1%d%02d%d", userIndex+1, j+1, k+1),
					CountryCode:    "PH",
					ContactNumber:  ptr(fmt.Sprintf("+63918%02d%02d%04d", userIndex+1, k+1, (userIndex*100)+k)),
					MediaID:        &media.ID,
				}
				if err := ds.model.BranchManager.Create(ctx, branch); err != nil {
					return err
				}

				// Create default branch settings for each branch
				branchSetting := &model.BranchSetting{
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
					BranchID:  branch.ID,

					// Withdraw Settings
					WithdrawAllowUserInput: true,
					WithdrawPrefix:         fmt.Sprintf("WD%d%d", j+1, k+1),
					WithdrawORStart:        1,
					WithdrawORCurrent:      1,
					WithdrawOREnd:          999999,
					WithdrawORIteration:    1,
					WithdrawORUnique:       true,
					WithdrawUseDateOR:      false,

					// Deposit Settings
					DepositAllowUserInput: true,
					DepositPrefix:         fmt.Sprintf("DP%d%d", j+1, k+1),
					DepositORStart:        1,
					DepositORCurrent:      1,
					DepositOREnd:          999999,
					DepositORIteration:    1,
					DepositORUnique:       true,
					DepositUseDateOR:      false,

					// Loan Settings
					LoanAllowUserInput: true,
					LoanPrefix:         fmt.Sprintf("LN%d%d", j+1, k+1),
					LoanORStart:        1,
					LoanORCurrent:      1,
					LoanOREnd:          999999,
					LoanORIteration:    1,
					LoanORUnique:       true,
					LoanUseDateOR:      false,

					// Check Voucher Settings
					CheckVoucherAllowUserInput: true,
					CheckVoucherPrefix:         fmt.Sprintf("CV%d%d", j+1, k+1),
					CheckVoucherORStart:        1,
					CheckVoucherORCurrent:      1,
					CheckVoucherOREnd:          999999,
					CheckVoucherORIteration:    1,
					CheckVoucherORUnique:       true,
					CheckVoucherUseDateOR:      false,

					// Default Member Type - can be set later when MemberType is available
					DefaultMemberTypeID:       nil,
					LoanAppliedEqualToBalance: true,
				}

				if err := ds.model.BranchSettingManager.Create(ctx, branchSetting); err != nil {
					return err
				}

				// Create the owner relationship for the user who created the organization
				developerKey, err := ds.provider.Service.Security.GenerateUUIDv5(ctx, fmt.Sprintf("%s-%s-%s", user.ID, organization.ID, branch.ID))
				if err != nil {
					return err
				}

				ownerOrganization := &model.UserOrganization{
					CreatedAt:                time.Now().UTC(),
					CreatedByID:              user.ID,
					UpdatedAt:                time.Now().UTC(),
					UpdatedByID:              user.ID,
					BranchID:                 &branch.ID,
					OrganizationID:           organization.ID,
					UserID:                   user.ID,
					UserType:                 model.UserOrganizationTypeOwner,
					Description:              fmt.Sprintf("Owner %s of %s", *user.FirstName, organization.Name),
					ApplicationDescription:   "Organization owner - automatically created",
					ApplicationStatus:        "accepted",
					DeveloperSecretKey:       developerKey + uuid.NewString() + "-owner-horizon",
					PermissionName:           "Owner",
					PermissionDescription:    "Organization owner with full permissions",
					Permissions:              []string{"read", "write", "manage", "delete", "admin"},
					UserSettingStartOR:       0,
					UserSettingEndOR:         10000,
					UserSettingUsedOR:        0,
					UserSettingStartVoucher:  0,
					UserSettingEndVoucher:    100,
					UserSettingUsedVoucher:   0,
					UserSettingNumberPadding: 7,
					Status:                   model.UserOrganizationStatusOffline,
					LastOnlineAt:             time.Now().UTC(),
				}

				if err := ds.model.UserOrganizationManager.Create(ctx, ownerOrganization); err != nil {
					return err
				}

				// Run organization seeder for accounting setup
				tx := ds.provider.Service.Database.Client().Begin()
				if tx.Error != nil {
					tx.Rollback()
					return err
				}
				if err := ds.model.OrganizationSeeder(ctx, tx, user.ID, organization.ID, branch.ID); err != nil {
					tx.Rollback()
					return err
				}
				if err := tx.Commit().Error; err != nil {
					return err
				}

				// Create invitation codes for this branch
				for m := 0; m < 5; m++ {
					userType := model.UserOrganizationTypeMember
					if m%2 == 0 {
						userType = model.UserOrganizationTypeEmployee
					}
					invitationCode := &model.InvitationCode{
						CreatedAt:      time.Now().UTC(),
						CreatedByID:    user.ID,
						UpdatedAt:      time.Now().UTC(),
						UpdatedByID:    user.ID,
						OrganizationID: organization.ID,
						BranchID:       branch.ID,
						UserType:       userType,
						Code:           uuid.New().String(),
						ExpirationDate: time.Now().UTC().Add(60 * 24 * time.Hour),
						MaxUse:         50,
						CurrentUse:     m % 25,
						Description:    fmt.Sprintf("Invite for %s Branch %d", organization.Name, k+1),
					}
					if err := ds.model.InvitationCodeManager.Create(ctx, invitationCode); err != nil {
						return err
					}
				}

				ds.provider.Service.Logger.Info(fmt.Sprintf("Created organization: %s with branch: %s (Owner: %s %s)",
					organization.Name, branch.Name, *user.FirstName, *user.LastName))
			}
		}
	}
	return nil
}

func (s *Seeder) SeedEmployees(ctx context.Context) error {
	s.provider.Service.Logger.Info("Seeding branch employees...")

	// Get all organizations and their branches
	organizations, err := s.model.OrganizationManager.List(ctx)
	if err != nil {
		return err
	}

	// Get all users to use as employees
	users, err := s.model.UserManager.List(ctx)
	if err != nil {
		return err
	}

	if len(users) == 0 || len(organizations) == 0 {
		s.provider.Service.Logger.Warn("No users or organizations found for employee seeding")
		return nil
	}

	// Create cross-employment: users work as employees in other users' organizations
	for _, org := range organizations {
		// Get branches for this organization
		branches, err := s.model.BranchManager.Find(ctx, &model.Branch{
			OrganizationID: org.ID,
		})
		if err != nil {
			continue
		}

		// Find potential employees (users who don't own this organization)
		potentialEmployees := make([]*model.User, 0)
		for _, user := range users {
			if user.ID != org.CreatedByID { // Don't make the owner an employee of their own organization
				potentialEmployees = append(potentialEmployees, user)
			}
		}

		if len(potentialEmployees) == 0 {
			continue
		}

		employeeIndex := 0
		for _, branch := range branches {
			// Check how many employees already exist for this branch (excluding owner)
			existingEmployees, err := s.model.Employees(ctx, org.ID, branch.ID)
			if err != nil {
				continue
			}

			// Create 1-3 employees per branch
			numEmployeesToCreate := 1 + (int(branch.ID.ID()) % 3)

			// Don't create more employees than available users
			if numEmployeesToCreate > len(potentialEmployees) {
				numEmployeesToCreate = len(potentialEmployees)
			}

			// Don't create more than 3 employees per branch
			if len(existingEmployees)+numEmployeesToCreate > 3 {
				numEmployeesToCreate = 3 - len(existingEmployees)
			}

			if numEmployeesToCreate <= 0 {
				continue
			}

			for i := 0; i < numEmployeesToCreate; i++ {
				// Cycle through potential employees
				selectedUser := potentialEmployees[employeeIndex%len(potentialEmployees)]
				employeeIndex++

				// Check if this user is already associated with this specific branch
				existingAssociation, err := s.model.UserOrganizationManager.Count(ctx, &model.UserOrganization{
					UserID:         selectedUser.ID,
					OrganizationID: org.ID,
					BranchID:       &branch.ID,
				})
				if err != nil || existingAssociation > 0 {
					// Skip if user is already associated with this branch
					continue
				}

				// Check if user can join as employee (ensure they're not already a member of this organization)
				if !s.model.UserOrganizationEmployeeCanJoin(ctx, selectedUser.ID, org.ID, branch.ID) {
					continue
				}

				// Generate developer key
				developerKey, err := s.provider.Service.Security.GenerateUUIDv5(ctx, fmt.Sprintf("emp-%s-%s-%s", selectedUser.ID, org.ID, branch.ID))
				if err != nil {
					return err
				}

				// Create employee user organization record
				employeeOrg := &model.UserOrganization{
					CreatedAt:                time.Now().UTC(),
					CreatedByID:              org.CreatedByID, // Created by the organization owner
					UpdatedAt:                time.Now().UTC(),
					UpdatedByID:              org.CreatedByID,
					BranchID:                 &branch.ID,
					OrganizationID:           org.ID,
					UserID:                   selectedUser.ID,
					UserType:                 model.UserOrganizationTypeEmployee,
					Description:              fmt.Sprintf("Employee %s %s at %s", *selectedUser.FirstName, *selectedUser.LastName, branch.Name),
					ApplicationDescription:   "Cross-organization employee for testing and demonstration",
					ApplicationStatus:        "accepted",
					DeveloperSecretKey:       developerKey + uuid.NewString() + "-employee-horizon",
					PermissionName:           "Employee",
					PermissionDescription:    "Branch employee with standard operational permissions",
					Permissions:              []string{"read", "create", "update"},
					UserSettingStartOR:       int64((i + 1) * 1000),
					UserSettingEndOR:         int64((i+1)*1000 + 999),
					UserSettingUsedOR:        0,
					UserSettingStartVoucher:  int64((i + 1) * 10),
					UserSettingEndVoucher:    int64((i+1)*10 + 9),
					UserSettingUsedVoucher:   0,
					UserSettingNumberPadding: 7,
					Status:                   model.UserOrganizationStatusOffline,
					LastOnlineAt:             time.Now().UTC(),
				}

				if err := s.model.UserOrganizationManager.Create(ctx, employeeOrg); err != nil {
					s.provider.Service.Logger.Error(fmt.Sprintf("Failed to create employee %s %s for branch %s: %v",
						*selectedUser.FirstName, *selectedUser.LastName, branch.Name, err))
					continue
				}

				s.provider.Service.Logger.Info(fmt.Sprintf("Created employee: %s %s for organization: %s, branch: %s (Owner: %s)",
					*selectedUser.FirstName, *selectedUser.LastName, org.Name, branch.Name,
					func() string {
						for _, u := range users {
							if u.ID == org.CreatedByID {
								return fmt.Sprintf("%s %s", *u.FirstName, *u.LastName)
							}
						}
						return "Unknown"
					}()))
			}
		}
	}

	s.provider.Service.Logger.Info("Employee seeding completed")
	return nil
}

func (ds *Seeder) SeedUsers(ctx context.Context) error {
	users, err := ds.model.UserManager.List(ctx)
	if err != nil {
		return err
	}
	if len(users) >= 1 {
		return nil
	}

	imageUrl := "https://yavuzceliker.github.io/sample-images/image-1021.jpg"
	image, err := ds.provider.Service.Storage.UploadFromURL(ctx, imageUrl, func(progress, total int64, storage *horizon.Storage) {})
	if err != nil {
		return err
	}
	media := &model.Media{
		FileName:   image.FileName,
		FileType:   image.FileType,
		FileSize:   image.FileSize,
		StorageKey: image.StorageKey,
		URL:        image.URL,
		BucketName: image.BucketName,
		Status:     "comppleted",
		Progress:   100,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}
	if err := ds.model.MediaManager.Create(ctx, media); err != nil {
		return err
	}

	// Create a single hashed password for all users
	hashedPassword, err := ds.provider.Service.Security.HashPassword(ctx, "sample-hello-world-12345")
	if err != nil {
		return err
	}

	// User 1
	user1 := &model.User{
		MediaID:           &media.ID,
		Email:             "sample@example.com",
		Password:          hashedPassword,
		Birthdate:         time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
		UserName:          "sampleuser",
		FullName:          "Sample User",
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
	if err := ds.model.UserManager.Create(ctx, user1); err != nil {
		return err
	}

	// User 2
	user2 := &model.User{
		MediaID:           &media.ID,
		Email:             "seconduser@example.com",
		Password:          hashedPassword,
		Birthdate:         time.Date(1992, time.March, 15, 0, 0, 0, 0, time.UTC),
		UserName:          "seconduser",
		FullName:          "Second User",
		FirstName:         ptr("Second"),
		MiddleName:        ptr("M."),
		LastName:          ptr("User"),
		Suffix:            ptr("Sr"),
		ContactNumber:     "+639234567890",
		IsEmailVerified:   true,
		IsContactVerified: true,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	if err := ds.model.UserManager.Create(ctx, user2); err != nil {
		return err
	}

	// User 3
	user3 := &model.User{
		MediaID:           &media.ID,
		Email:             "thirduser@example.com",
		Password:          hashedPassword,
		Birthdate:         time.Date(1988, time.June, 5, 0, 0, 0, 0, time.UTC),
		UserName:          "thirduser",
		FullName:          "Third User",
		FirstName:         ptr("Third"),
		MiddleName:        ptr("J."),
		LastName:          ptr("User"),
		Suffix:            ptr("II"),
		ContactNumber:     "+639345678901",
		IsEmailVerified:   true,
		IsContactVerified: true,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	if err := ds.model.UserManager.Create(ctx, user3); err != nil {
		return err
	}

	// User 4
	user4 := &model.User{
		MediaID:           &media.ID,
		Email:             "fourthuser@example.com",
		Password:          hashedPassword,
		Birthdate:         time.Date(1995, time.December, 20, 0, 0, 0, 0, time.UTC),
		UserName:          "fourthuser",
		FullName:          "Fourth User",
		FirstName:         ptr("Fourth"),
		MiddleName:        ptr("K."),
		LastName:          ptr("User"),
		Suffix:            ptr("III"),
		ContactNumber:     "+639456789012",
		IsEmailVerified:   true,
		IsContactVerified: true,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}
	if err := ds.model.UserManager.Create(ctx, user4); err != nil {
		return err
	}
	return nil
}

func (s *Seeder) SeedMemberProfiles(ctx context.Context) error {
	// Check if member profiles already exist
	profiles, err := s.model.MemberProfileManager.List(ctx)
	if err != nil {
		return err
	}
	if len(profiles) > 0 {
		return nil
	}

	s.provider.Service.Logger.Info("Seeding member profiles...")

	// Get existing organizations and branches to seed member profiles for
	organizations, err := s.model.OrganizationManager.List(ctx)
	if err != nil {
		return err
	}

	// Get existing users to associate with member profiles
	users, err := s.model.UserManager.List(ctx)
	if err != nil {
		return err
	}

	if len(organizations) == 0 || len(users) == 0 {
		s.provider.Service.Logger.Warn("No organizations or users found, skipping member profile seeding")
		return nil
	}

	// Sample member profile data
	sampleMembers := []struct {
		FirstName     string
		MiddleName    string
		LastName      string
		ContactNumber string
		CivilStatus   string
		BusinessAddr  string
	}{
		{"Juan", "Dela", "Cruz", "+639171234567", "married", "123 Rizal Street, Manila"},
		{"Maria", "Santos", "Garcia", "+639182345678", "single", "456 Bonifacio Ave, Quezon City"},
		{"Jose", "Mercado", "Rizal", "+639193456789", "married", "789 Mabini Street, Makati"},
		{"Ana", "Reyes", "Santos", "+639204567890", "widowed", "321 Luna Street, Pasig"},
		{"Pedro", "Villa", "Moreno", "+639215678901", "single", "654 Del Pilar Ave, Taguig"},
		{"Carmen", "Torres", "Valdez", "+639226789012", "married", "987 Aguinaldo Street, Mandaluyong"},
		{"Roberto", "Flores", "Mendoza", "+639237890123", "divorced", "147 Roxas Blvd, Pasay"},
		{"Sofia", "Ramos", "Herrera", "+639248901234", "single", "258 EDSA, Caloocan"},
		{"Miguel", "Castro", "Jimenez", "+639259012345", "married", "369 Commonwealth Ave, Quezon City"},
		{"Isabella", "Vargas", "Romero", "+639260123456", "single", "741 Ortigas Ave, San Juan"},
	}

	for _, org := range organizations {
		// Get branches for this organization
		branches, err := s.model.BranchManager.Find(ctx, &model.Branch{
			OrganizationID: org.ID,
		})
		if err != nil {
			continue
		}

		for _, branch := range branches {
			// Create 3-5 member profiles per branch
			numMembers := 3 + (int(org.ID.ID()) % 3) // 3-5 members per branch
			if numMembers > len(sampleMembers) {
				numMembers = len(sampleMembers)
			}

			for i := 0; i < numMembers; i++ {
				sample := sampleMembers[i]

				// Create birthdate (ages 25-65)
				age := 25 + (i % 40)
				birthDate := time.Now().AddDate(-age, 0, 0)

				// Generate full name
				fullName := fmt.Sprintf("%s %s %s", sample.FirstName, sample.MiddleName, sample.LastName)

				// Generate passbook number
				passbook := fmt.Sprintf("PB-%s-%04d", branch.Name[:min(3, len(branch.Name))], i+1)

				// Create member profile
				memberProfile := &model.MemberProfile{
					CreatedAt:             time.Now().UTC(),
					CreatedByID:           org.CreatedByID,
					UpdatedAt:             time.Now().UTC(),
					UpdatedByID:           org.CreatedByID,
					OrganizationID:        org.ID,
					BranchID:              branch.ID,
					UserID:                &users[i%len(users)].ID, // Rotate through available users
					FirstName:             sample.FirstName,
					MiddleName:            sample.MiddleName,
					LastName:              sample.LastName,
					FullName:              fullName,
					BirthDate:             birthDate,
					Status:                []string{"active", "pending", "inactive"}[i%3],
					Description:           fmt.Sprintf("Seeded member profile for %s", fullName),
					Notes:                 fmt.Sprintf("Generated member for testing purposes in %s branch", branch.Name),
					ContactNumber:         sample.ContactNumber,
					OldReferenceID:        fmt.Sprintf("REF-%04d", i+1),
					Passbook:              passbook,
					Occupation:            []string{"Farmer", "Teacher", "Driver", "Vendor", "Employee", "Business Owner"}[i%6],
					BusinessAddress:       sample.BusinessAddr,
					BusinessContactNumber: fmt.Sprintf("+639%03d%06d", 100+i, 100000+i),
					CivilStatus:           sample.CivilStatus,
					IsClosed:              false,
					IsMutualFundMember:    i%2 == 0,
					IsMicroFinanceMember:  i%3 == 0,
				}

				if err := s.model.MemberProfileManager.Create(ctx, memberProfile); err != nil {
					s.provider.Service.Logger.Error(fmt.Sprintf("Failed to create member profile %s: %v", fullName, err))
					continue
				}

				// Create sample member address
				memberAddress := &model.MemberAddress{
					CreatedAt:       time.Now().UTC(),
					UpdatedAt:       time.Now().UTC(),
					CreatedByID:     org.CreatedByID,
					UpdatedByID:     org.CreatedByID,
					OrganizationID:  org.ID,
					BranchID:        branch.ID,
					MemberProfileID: &memberProfile.ID,
					Label:           "home",
					Address:         sample.BusinessAddr,
					ProvinceState:   "Metro Manila",
					City:            []string{"Manila", "Quezon City", "Makati", "Pasig", "Taguig"}[i%5],
					Barangay:        fmt.Sprintf("Barangay %d", i+1),
					PostalCode:      fmt.Sprintf("1%03d", 100+i),
					CountryCode:     "PH",
					Landmark:        fmt.Sprintf("Near %s Mall", []string{"SM", "Ayala", "Robinson", "Gateway", "Trinoma"}[i%5]),
				}

				if err := s.model.MemberAddressManager.Create(ctx, memberAddress); err != nil {
					s.provider.Service.Logger.Error(fmt.Sprintf("Failed to create member address for %s: %v", fullName, err))
				}

				s.provider.Service.Logger.Info(fmt.Sprintf("Created member profile: %s for branch: %s", fullName, branch.Name))
			}
		}
	}

	s.provider.Service.Logger.Info("Member profile seeding completed")
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func ptr[T any](v T) *T {
	return &v
}
