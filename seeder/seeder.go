package seeder

import (
	"context"
	"fmt"
	"io/fs"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/model/model_core"
	"github.com/google/uuid"
	"github.com/jaswdr/faker"
	"github.com/rotisserie/eris"
	"github.com/schollz/progressbar/v3"
)

type Seeder struct {
	provider   *src.Provider
	model_core *model_core.ModelCore
	faker      faker.Faker
	imagePaths []string
}

func NewSeeder(provider *src.Provider, model_core *model_core.ModelCore) (*Seeder, error) {
	seeder := &Seeder{
		provider:   provider,
		model_core: model_core,
		faker:      faker.New(),
	}

	// Load image paths from seeder/images directory
	if err := seeder.loadImagePaths(); err != nil {
		return nil, eris.Wrap(err, "failed to load image paths")
	}

	return seeder, nil
}

// loadImagePaths scans the seeder/images directory and loads all image file paths
func (s *Seeder) loadImagePaths() error {
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
				s.imagePaths = append(s.imagePaths, path)
			}
		}

		return nil
	})

	if err != nil {
		return eris.Wrap(err, "failed to scan images directory")
	}

	if len(s.imagePaths) == 0 {
		return eris.New("no image files found in seeder/images directory")
	}

	s.provider.Service.Logger.Info(fmt.Sprintf("Loaded %d image files from seeder/images directory", len(s.imagePaths)))
	return nil
}

// createImageMedia randomly selects a local image file and creates a media record for it
func (s *Seeder) createImageMedia(ctx context.Context, imageType string) (*model_core.Media, error) {
	if len(s.imagePaths) == 0 {
		return nil, eris.New("no image files available for seeding")
	}

	// Randomly choose one image from the loaded paths
	randomIndex := rand.Intn(len(s.imagePaths))
	imagePath := s.imagePaths[randomIndex]

	// Upload the image from local path
	storage, err := s.provider.Service.Storage.UploadFromPath(ctx, imagePath, func(progress, total int64, storage *horizon.Storage) {})
	if err != nil {
		return nil, eris.Wrapf(err, "failed to upload image from path %s for %s", imagePath, imageType)
	} // Create media record
	media := &model_core.Media{
		FileName:   storage.FileName,
		FileType:   storage.FileType,
		FileSize:   storage.FileSize,
		StorageKey: storage.StorageKey,
		URL:        storage.URL,
		BucketName: storage.BucketName,
		Status:     "completed",
		Progress:   100,
		CreatedAt:  time.Now().UTC(),
		UpdatedAt:  time.Now().UTC(),
	}

	if err := s.model_core.MediaManager.Create(ctx, media); err != nil {
		return nil, eris.Wrap(err, "failed to create media record")
	}

	return media, nil
}

func (s *Seeder) Run(ctx context.Context, multiplier int32) error {
	if multiplier <= 0 {
		s.provider.Service.Logger.Info("Multiplier is 0 or less, skipping database seeding.")
		return nil
	}

	s.provider.Service.Logger.Info("Starting database seeding with multiplier: " + fmt.Sprintf("%d", multiplier))

	// Create overall progress bar for main seeding steps
	totalSteps := 7 // CategorySeed, CurrencySeed, SubscriptionPlanSeed, SeedUsers, SeedOrganization, SeedEmployees, SeedMemberProfiles
	overallBar := progressbar.NewOptions(totalSteps,
		progressbar.OptionSetDescription("Overall seeding progress..."),
		progressbar.OptionSetWidth(60),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "▓",
			SaucerHead:    "▓",
			SaucerPadding: "░",
			BarStart:      "╢",
			BarEnd:        "╟",
		}),
	)

	if err := s.model_core.CategorySeed(ctx); err != nil {
		return err
	}
	overallBar.Add(1)

	if err := s.model_core.CurrencySeed(ctx); err != nil {
		return err
	}
	overallBar.Add(1)

	if err := s.model_core.SubscriptionPlanSeed(ctx); err != nil {
		return err
	}
	overallBar.Add(1)

	if err := s.SeedUsers(ctx, multiplier); err != nil {
		return err
	}
	overallBar.Add(1)

	if err := s.SeedOrganization(ctx, multiplier); err != nil {
		return err
	}
	overallBar.Add(1)

	if err := s.SeedEmployees(ctx, multiplier); err != nil {
		return err
	}
	overallBar.Add(1)

	if err := s.SeedMemberProfiles(ctx, multiplier); err != nil {
		return err
	}
	overallBar.Add(1)

	// Finish overall progress bar
	overallBar.Finish()
	s.provider.Service.Logger.Info("Seeding completed successfully.")
	return nil
}

func (s *Seeder) SeedOrganization(ctx context.Context, multiplier int32) error {
	orgs, err := s.model_core.OrganizationManager.List(ctx)
	if err != nil {
		return err
	}
	if len(orgs) > 0 {
		return nil
	}

	numOrgsPerUser := int(multiplier) * 1
	users, err := s.model_core.UserManager.List(ctx)
	if err != nil {
		return err
	}
	subscriptions, err := s.model_core.SubscriptionPlanManager.List(ctx)
	if err != nil {
		return err
	}
	categories, err := s.model_core.CategoryManager.List(ctx)
	if err != nil {
		return err
	}

	// Create progress bar for organization creation
	totalOrgs := len(users) * numOrgsPerUser
	orgBar := progressbar.NewOptions(totalOrgs,
		progressbar.OptionSetDescription("Creating organizations..."),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	for _, user := range users {
		for j := 0; j < numOrgsPerUser; j++ {
			sub := subscriptions[j%len(subscriptions)]
			subscriptionEndDate := time.Now().Add(30 * 24 * time.Hour)
			orgMedia, err := s.createImageMedia(ctx, "Organization")
			if err != nil {
				return eris.Wrap(err, "failed to create organization media")
			}
			organization := &model_core.Organization{
				CreatedAt:                           time.Now().UTC(),
				CreatedByID:                         user.ID,
				UpdatedAt:                           time.Now().UTC(),
				UpdatedByID:                         user.ID,
				Name:                                s.faker.Company().Name(),
				Address:                             ptr(s.faker.Address().Address()),
				Email:                               ptr(s.faker.Internet().Email()),
				ContactNumber:                       ptr(fmt.Sprintf("+6391%08d", s.faker.IntBetween(10000000, 99999999))),
				Description:                         ptr(s.faker.Lorem().Paragraph(3)),
				Color:                               ptr(s.faker.Color().Hex()),
				TermsAndConditions:                  ptr(s.faker.Lorem().Paragraph(5)),
				PrivacyPolicy:                       ptr(s.faker.Lorem().Paragraph(5)),
				CookiePolicy:                        ptr(s.faker.Lorem().Paragraph(5)),
				RefundPolicy:                        ptr(s.faker.Lorem().Paragraph(5)),
				UserAgreement:                       ptr(s.faker.Lorem().Paragraph(5)),
				IsPrivate:                           s.faker.Bool(),
				MediaID:                             &orgMedia.ID,
				CoverMediaID:                        &orgMedia.ID,
				SubscriptionPlanMaxBranches:         sub.MaxBranches,
				SubscriptionPlanMaxEmployees:        sub.MaxEmployees,
				SubscriptionPlanMaxMembersPerBranch: sub.MaxMembersPerBranch,
				SubscriptionPlanID:                  &sub.ID,
				SubscriptionStartDate:               time.Now().UTC(),
				SubscriptionEndDate:                 subscriptionEndDate,
			}
			if err := s.model_core.OrganizationManager.Create(ctx, organization); err != nil {
				return err
			}

			// Add categories to organization
			for _, category := range categories {
				if err := s.model_core.OrganizationCategoryManager.Create(ctx, &model_core.OrganizationCategory{
					CreatedAt:      time.Now().UTC(),
					UpdatedAt:      time.Now().UTC(),
					OrganizationID: &organization.ID,
					CategoryID:     &category.ID,
				}); err != nil {
					return err
				}
			}

			numBranches := int(multiplier) * 1

			// Create progress bar for branches within this organization
			branchBar := progressbar.NewOptions(numBranches,
				progressbar.OptionSetDescription(fmt.Sprintf("Creating branches for %s...", organization.Name)),
				progressbar.OptionSetWidth(40),
				progressbar.OptionShowCount(),
				progressbar.OptionClearOnFinish(),
			)

			for k := range numBranches {
				branchMedia, err := s.createImageMedia(ctx, "Organization")
				if err != nil {
					return eris.Wrap(err, "failed to create organization media")
				}
				branch := &model_core.Branch{
					CreatedAt:      time.Now().UTC(),
					CreatedByID:    user.ID,
					UpdatedAt:      time.Now().UTC(),
					UpdatedByID:    user.ID,
					OrganizationID: organization.ID,
					Type:           []string{"main", "satellite", "branch"}[k%3],
					Name:           s.faker.Company().Name(),
					Email:          s.faker.Internet().Email(),
					Address:        s.faker.Address().Address(),
					Province:       s.faker.Address().State(),
					City:           s.faker.Address().City(),
					Region:         s.faker.Address().State(),
					Barangay:       s.faker.Address().StreetName(),
					PostalCode:     s.faker.Address().PostCode(),
					CountryCode:    "PH",
					ContactNumber:  ptr(fmt.Sprintf("+6391%08d", s.faker.IntBetween(10000000, 99999999))),
					MediaID:        &branchMedia.ID,
					// Random coordinates for Philippines (approximately between 4°-21°N, 116°-127°E)
					Latitude:  ptr(4.0 + float64(s.faker.IntBetween(0, 1700))/100.0),
					Longitude: ptr(116.0 + float64(s.faker.IntBetween(0, 1100))/100.0),
				}
				if err := s.model_core.BranchManager.Create(ctx, branch); err != nil {
					return err
				}
				currency, err := s.model_core.CurrencyFindByAlpha2(ctx, branch.CountryCode)
				if err != nil {
					return eris.Wrap(err, "failed to find currency for account seeding")
				}
				// Create default branch settings for each branch
				branchSetting := &model_core.BranchSetting{
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
					BranchID:  branch.ID,

					// Withdraw Settings
					WithdrawAllowUserInput: true,
					WithdrawPrefix:         s.faker.Lorem().Word(),
					WithdrawORStart:        1,
					WithdrawORCurrent:      1,
					WithdrawOREnd:          999999,
					WithdrawORIteration:    1,
					WithdrawORUnique:       true,
					WithdrawUseDateOR:      false,

					// Deposit Settings
					DepositAllowUserInput: true,
					DepositPrefix:         s.faker.Lorem().Word(),
					DepositORStart:        1,
					DepositORCurrent:      1,
					DepositOREnd:          999999,
					DepositORIteration:    1,
					DepositORUnique:       true,
					DepositUseDateOR:      false,

					// Loan Settings
					LoanAllowUserInput: true,
					LoanPrefix:         s.faker.Lorem().Word(),
					LoanORStart:        1,
					LoanORCurrent:      1,
					LoanOREnd:          999999,
					LoanORIteration:    1,
					LoanORUnique:       true,
					LoanUseDateOR:      false,

					// Check Voucher Settings
					CheckVoucherAllowUserInput: true,
					CheckVoucherPrefix:         s.faker.Lorem().Word(),
					CheckVoucherORStart:        1,
					CheckVoucherORCurrent:      1,
					CheckVoucherOREnd:          999999,
					CheckVoucherORIteration:    1,
					CheckVoucherORUnique:       true,
					CheckVoucherUseDateOR:      false,

					// Default Member Type - can be set later when MemberType is available
					DefaultMemberTypeID:       nil,
					LoanAppliedEqualToBalance: true,
					CurrencyID:                currency.ID,
				}

				if err := s.model_core.BranchSettingManager.Create(ctx, branchSetting); err != nil {
					return err
				}

				// Create the owner relationship for the user who created the organization
				developerKey, err := s.provider.Service.Security.GenerateUUIDv5(ctx, fmt.Sprintf("%s-%s-%s", user.ID, organization.ID, branch.ID))
				if err != nil {
					return err
				}

				ownerOrganization := &model_core.UserOrganization{
					CreatedAt:                time.Now().UTC(),
					CreatedByID:              user.ID,
					UpdatedAt:                time.Now().UTC(),
					UpdatedByID:              user.ID,
					BranchID:                 &branch.ID,
					OrganizationID:           organization.ID,
					UserID:                   user.ID,
					UserType:                 model_core.UserOrganizationTypeOwner,
					Description:              s.faker.Lorem().Sentence(5),
					ApplicationDescription:   s.faker.Lorem().Sentence(3),
					ApplicationStatus:        "accepted",
					DeveloperSecretKey:       developerKey + uuid.NewString() + "-owner-horizon",
					PermissionName:           "Employee",
					PermissionDescription:    "Organization owner with full permissions",
					Permissions:              []string{"read", "write", "manage", "delete", "admin"},
					UserSettingStartOR:       0,
					UserSettingEndOR:         10000,
					UserSettingUsedOR:        0,
					UserSettingStartVoucher:  0,
					UserSettingEndVoucher:    100,
					UserSettingUsedVoucher:   0,
					UserSettingNumberPadding: 7,
					Status:                   model_core.UserOrganizationStatusOffline,
					LastOnlineAt:             time.Now().UTC(),
				}

				if err := s.model_core.UserOrganizationManager.Create(ctx, ownerOrganization); err != nil {
					return err
				}

				// Run organization seeder for accounting setup
				tx := s.provider.Service.Database.Client().Begin()
				if tx.Error != nil {
					tx.Rollback()
					return err
				}
				if err := s.model_core.OrganizationSeeder(ctx, tx, user.ID, organization.ID, branch.ID); err != nil {
					tx.Rollback()
					return err
				}
				if err := tx.Commit().Error; err != nil {
					return err
				}

				numInvites := int(multiplier) * 1
				for m := 0; m < numInvites; m++ {
					userType := model_core.UserOrganizationTypeMember
					if m%2 == 0 {
						userType = model_core.UserOrganizationTypeEmployee
					}
					invitationCode := &model_core.InvitationCode{
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
						Description:    s.faker.Lorem().Sentence(3),
					}
					if err := s.model_core.InvitationCodeManager.Create(ctx, invitationCode); err != nil {
						return err
					}
				}

				s.provider.Service.Logger.Info(fmt.Sprintf("Created organization: %s with branch: %s (Owner: %s %s)",
					organization.Name, branch.Name, *user.FirstName, *user.LastName))

				// Update branch progress bar
				branchBar.Add(1)
			}

			// Finish branch progress bar
			branchBar.Finish()

			// Update organization progress bar
			orgBar.Add(1)
		}
	}

	// Finish organization progress bar
	orgBar.Finish()
	return nil
}

func (s *Seeder) SeedEmployees(ctx context.Context, multiplier int32) error {
	s.provider.Service.Logger.Info("Seeding branch employees...")

	// Get all organizations and their branches
	organizations, err := s.model_core.OrganizationManager.List(ctx)
	if err != nil {
		return err
	}

	// Get all users to use as employees
	users, err := s.model_core.UserManager.List(ctx)
	if err != nil {
		return err
	}

	if len(users) == 0 || len(organizations) == 0 {
		s.provider.Service.Logger.Warn("No users or organizations found for employee seeding")
		return nil
	}

	// Create progress bar for employee creation
	employeeBar := progressbar.NewOptions(len(organizations),
		progressbar.OptionSetDescription("Creating employees..."),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	// Create cross-employment: users work as employees in other users' organizations
	for _, org := range organizations {
		// Get branches for this organization
		branches, err := s.model_core.BranchManager.Find(ctx, &model_core.Branch{
			OrganizationID: org.ID,
		})
		if err != nil {
			continue
		}

		// Find potential employees (users who don't own this organization)
		potentialEmployees := make([]*model_core.User, 0)
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
			existingEmployees, err := s.model_core.Employees(ctx, org.ID, branch.ID)
			if err != nil {
				continue
			}

			numEmployeesToCreate := int(multiplier) * 1

			// Don't create more employees than available users
			if numEmployeesToCreate > len(potentialEmployees) {
				numEmployeesToCreate = len(potentialEmployees)
			}

			// Cap at a reasonable number per branch, e.g., 3 * multiplier, but adjust if needed
			maxPerBranch := 3 * int(multiplier)
			if len(existingEmployees)+numEmployeesToCreate > maxPerBranch {
				numEmployeesToCreate = maxPerBranch - len(existingEmployees)
			}

			if numEmployeesToCreate <= 0 {
				continue
			}

			for i := 0; i < numEmployeesToCreate; i++ {
				// Cycle through potential employees
				selectedUser := potentialEmployees[employeeIndex%len(potentialEmployees)]
				employeeIndex++

				// Check if this user is already associated with this specific branch
				existingAssociation, err := s.model_core.UserOrganizationManager.Count(ctx, &model_core.UserOrganization{
					UserID:         selectedUser.ID,
					OrganizationID: org.ID,
					BranchID:       &branch.ID,
				})
				if err != nil || existingAssociation > 0 {
					// Skip if user is already associated with this branch
					continue
				}

				// Check if user can join as employee (ensure they're not already a member of this organization)
				if !s.model_core.UserOrganizationEmployeeCanJoin(ctx, selectedUser.ID, org.ID, branch.ID) {
					continue
				}

				// Generate developer key
				developerKey, err := s.provider.Service.Security.GenerateUUIDv5(ctx, fmt.Sprintf("emp-%s-%s-%s", selectedUser.ID, org.ID, branch.ID))
				if err != nil {
					return err
				}

				// Create employee user organization record
				employeeOrg := &model_core.UserOrganization{
					CreatedAt:                time.Now().UTC(),
					CreatedByID:              org.CreatedByID, // Created by the organization owner
					UpdatedAt:                time.Now().UTC(),
					UpdatedByID:              org.CreatedByID,
					BranchID:                 &branch.ID,
					OrganizationID:           org.ID,
					UserID:                   selectedUser.ID,
					UserType:                 model_core.UserOrganizationTypeEmployee,
					Description:              s.faker.Lorem().Sentence(5),
					ApplicationDescription:   s.faker.Lorem().Sentence(3),
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
					Status:                   model_core.UserOrganizationStatusOffline,
					LastOnlineAt:             time.Now().UTC(),
				}

				if err := s.model_core.UserOrganizationManager.Create(ctx, employeeOrg); err != nil {
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

		// Update progress bar for each organization processed
		employeeBar.Add(1)
	}

	// Finish progress bar
	employeeBar.Finish()
	s.provider.Service.Logger.Info("Employee seeding completed")
	return nil
}

func (s *Seeder) SeedUsers(ctx context.Context, multiplier int32) error {
	users, err := s.model_core.UserManager.List(ctx)
	if err != nil {
		return err
	}
	if len(users) >= 1 {
		return nil
	}

	// Generate a secure random password base and hash it
	basePassword := "sample-hello-world-12345"
	hashedPassword, err := s.provider.Service.Security.HashPassword(ctx, basePassword)
	if err != nil {
		return err
	}

	// Base number of users is 4, scale with multiplier
	baseNumUsers := 1
	numUsers := int(multiplier) * baseNumUsers

	// Create progress bar for user creation
	userBar := progressbar.NewOptions(numUsers,
		progressbar.OptionSetDescription("Creating users..."),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	for i := range numUsers {
		firstName := s.faker.Person().FirstName()
		middleName := s.faker.Person().LastName()[:1] // Simulate middle initial
		lastName := s.faker.Person().LastName()
		suffix := s.faker.Person().Suffix()
		fullName := fmt.Sprintf("%s %s %s %s", firstName, middleName, lastName, suffix)
		birthdate := time.Now().AddDate(-25-s.faker.IntBetween(0, 40), -s.faker.IntBetween(0, 11), -s.faker.IntBetween(0, 30))
		// Create shared media for all users using local image generation
		userSharedMedia, err := s.createImageMedia(ctx, "User")
		if err != nil {
			return eris.Wrap(err, "failed to create user media")
		}
		var email string
		if i == 0 {
			email = "sample@example.com"
		} else {
			email = s.faker.Internet().Email()
		}

		user := &model_core.User{
			MediaID:           &userSharedMedia.ID,
			Email:             email,
			Password:          hashedPassword,
			Birthdate:         birthdate,
			UserName:          s.faker.Internet().User(),
			FullName:          fullName,
			FirstName:         ptr(firstName),
			MiddleName:        ptr(middleName),
			LastName:          ptr(lastName),
			Suffix:            ptr(suffix),
			ContactNumber:     fmt.Sprintf("+6391%08d", s.faker.IntBetween(10000000, 99999999)),
			IsEmailVerified:   true,
			IsContactVerified: true,
			CreatedAt:         time.Now().UTC(),
			UpdatedAt:         time.Now().UTC(),
		}
		if err := s.model_core.UserManager.Create(ctx, user); err != nil {
			return err
		}

		// Update progress bar
		userBar.Add(1)
	}

	// Finish progress bar
	userBar.Finish()
	return nil
}

func (s *Seeder) SeedMemberProfiles(ctx context.Context, multiplier int32) error {
	// Check if member profiles already exist
	profiles, err := s.model_core.MemberProfileManager.List(ctx)
	if err != nil {
		return err
	}
	if len(profiles) > 0 {
		return nil
	}

	s.provider.Service.Logger.Info("Seeding member profiles...")

	// Get existing organizations and branches to seed member profiles for
	organizations, err := s.model_core.OrganizationManager.List(ctx)
	if err != nil {
		return err
	}

	// Get existing users to associate with member profiles
	users, err := s.model_core.UserManager.List(ctx)
	if err != nil {
		return err
	}

	if len(organizations) == 0 || len(users) == 0 {
		s.provider.Service.Logger.Warn("No organizations or users found, skipping member profile seeding")
		return nil
	}

	// Create progress bar for member profiles
	memberBar := progressbar.NewOptions(len(organizations),
		progressbar.OptionSetDescription("Creating member profiles..."),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	for _, org := range organizations {
		// Get branches for this organization
		branches, err := s.model_core.BranchManager.Find(ctx, &model_core.Branch{
			OrganizationID: org.ID,
		})
		if err != nil {
			continue
		}

		for _, branch := range branches {
			numMembers := int(multiplier) * 1
			if numMembers > len(users) {
				numMembers = len(users)
			}

			for i := 0; i < numMembers; i++ {
				firstName := s.faker.Person().FirstName()
				middleName := s.faker.Person().LastName()[:1]
				lastName := s.faker.Person().LastName()
				fullName := fmt.Sprintf("%s %s %s", firstName, middleName, lastName)

				// Create birthdate (ages 25-65)
				age := 25 + (i % 40)
				birthDate := time.Now().AddDate(-age, 0, 0)

				// Generate passbook number
				passbook := fmt.Sprintf("PB-%s-%04d", branch.Name[:min(3, len(branch.Name))], i+1)

				// Create member profile
				memberProfile := &model_core.MemberProfile{
					CreatedAt:             time.Now().UTC(),
					CreatedByID:           org.CreatedByID,
					UpdatedAt:             time.Now().UTC(),
					UpdatedByID:           org.CreatedByID,
					OrganizationID:        org.ID,
					BranchID:              branch.ID,
					UserID:                &users[i%len(users)].ID, // Rotate through available users
					FirstName:             firstName,
					MiddleName:            middleName,
					LastName:              lastName,
					FullName:              fullName,
					BirthDate:             birthDate,
					Status:                []string{"active", "pending", "inactive"}[i%3],
					Description:           s.faker.Lorem().Paragraph(2),
					Notes:                 s.faker.Lorem().Paragraph(1),
					ContactNumber:         fmt.Sprintf("+6391%08d", s.faker.IntBetween(10000000, 99999999)),
					OldReferenceID:        fmt.Sprintf("REF-%04d", i+1),
					Passbook:              passbook,
					Occupation:            []string{"Farmer", "Teacher", "Driver", "Vendor", "Employee", "Business Owner"}[i%6],
					BusinessAddress:       s.faker.Address().Address(),
					BusinessContactNumber: fmt.Sprintf("+6391%08d", s.faker.IntBetween(10000000, 99999999)),
					CivilStatus:           []string{"married", "single", "widowed", "divorced"}[i%4],
					IsClosed:              false,
					IsMutualFundMember:    i%2 == 0,
					IsMicroFinanceMember:  i%3 == 0,
				}

				if err := s.model_core.MemberProfileManager.Create(ctx, memberProfile); err != nil {
					s.provider.Service.Logger.Error(fmt.Sprintf("Failed to create member profile %s: %v", fullName, err))
					continue
				}

				// Create sample member address
				memberAddress := &model_core.MemberAddress{
					CreatedAt:       time.Now().UTC(),
					UpdatedAt:       time.Now().UTC(),
					CreatedByID:     org.CreatedByID,
					UpdatedByID:     org.CreatedByID,
					OrganizationID:  org.ID,
					BranchID:        branch.ID,
					MemberProfileID: &memberProfile.ID,
					Label:           []string{"home", "work", "other"}[i%3],
					Address:         s.faker.Address().Address(),
					ProvinceState:   s.faker.Address().State(),
					City:            s.faker.Address().City(),
					Barangay:        s.faker.Address().StreetName(),
					PostalCode:      s.faker.Address().PostCode(),
					CountryCode:     "PH",
					Landmark:        s.faker.Lorem().Sentence(2),
				}

				if err := s.model_core.MemberAddressManager.Create(ctx, memberAddress); err != nil {
					s.provider.Service.Logger.Error(fmt.Sprintf("Failed to create member address for %s: %v", fullName, err))
				}

				s.provider.Service.Logger.Info(fmt.Sprintf("Created member profile: %s for branch: %s", fullName, branch.Name))
			}
		}

		// Update progress bar for each organization processed
		memberBar.Add(1)
	}

	// Finish progress bar
	memberBar.Finish()
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
