package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server"
	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/modelcore"
	"github.com/google/uuid"
	"github.com/jaswdr/faker"
	"github.com/rotisserie/eris"
	"github.com/schollz/progressbar/v3"
)

type Seeder struct {
	provider    *server.Provider
	modelcore   *modelcore.ModelCore
	faker       faker.Faker
	imagePaths  []string
	progressBar *progressbar.ProgressBar
}

func NewSeeder(provider *server.Provider, modelcore *modelcore.ModelCore) (*Seeder, error) {
	// We'll update the progress bar total when Run() is called with the multiplier
	overallBar := progressbar.NewOptions(100, // Temporary, will be updated in Run()
		progressbar.OptionSetDescription("🌱 Database Seeding Progress"),
		progressbar.OptionSetWidth(70),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: "░",
			BarStart:      "│",
			BarEnd:        "│",
		}),
	)
	seeder := &Seeder{
		provider:    provider,
		modelcore:   modelcore,
		faker:       faker.New(),
		progressBar: overallBar,
	}
	if err := seeder.loadImagePaths(); err != nil {
		return nil, eris.Wrap(err, "failed to load image paths")
	}
	return seeder, nil
}

func (s *Seeder) Run(ctx context.Context, multiplier int32) error {
	if multiplier <= 0 {
		return nil
	}
	// Calculate accurate total operations for progress bar based on actual .Add(1) calls
	numUsers := int(multiplier) * 1            // Users to create
	numOrgsPerUser := int(multiplier) * 1      // Organizations per user
	numBranchesPerOrg := int(multiplier) * 1   // Branches per organization
	numMembersPerBranch := int(multiplier) * 1 // Members per branch
	// Accurate calculation matching actual .Add(1) calls:
	totalOperations :=
		// Initial seeding (3) + post-function adds (4) = 7
		3 + 4 +
			// SeedUsers: 1 per user
			numUsers +
			// SeedOrganization: per user + per org setup + per branch operations + invites
			(numUsers * 1) + // "Processing organizations for user"
			(numUsers * numOrgsPerUser * 3) + // "Setting up org" + "Created org media" + "Created organization"
			(numUsers * numOrgsPerUser * numBranchesPerOrg * 6) + // "Created branch media" + "Created branch" + "Created settings" + "Created owner" + "Setup accounting" + "Created invites"
			// SeedEmployees: initial lists + per org processing + per employee
			2 + // Lists orgs + users
			(numUsers * numOrgsPerUser * 3) + // Per org: processing + branches + potentials
			(numUsers * numOrgsPerUser * numBranchesPerOrg * numUsers) + // Per potential employee (conservative estimate)
			// SeedMemberProfiles: per org + per member operations
			(numUsers * numOrgsPerUser * 1) + // "Processing member profiles for organization"
			(numUsers * numOrgsPerUser * numBranchesPerOrg * numMembersPerBranch * 2) // "Created member" + "Added address"

	// Update progress bar with actual total
	s.progressBar = progressbar.NewOptions(totalOperations,
		progressbar.OptionSetDescription("🌱 Database Seeding Progress"),
		progressbar.OptionSetWidth(70),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: "░",
			BarStart:      "│",
			BarEnd:        "│",
		}),
	)

	s.progressBar.Describe("🏷️  Seeding categories...")
	if err := s.modelcore.CategorySeed(ctx); err != nil {
		return err
	}
	s.provider.Service.Logger.Info("✅ Completed CategorySeed - Progress: 1/Total")
	s.progressBar.Add(1)

	s.progressBar.Describe("💰 Seeding currencies...")
	if err := s.modelcore.CurrencySeed(ctx); err != nil {
		return err
	}
	s.provider.Service.Logger.Info("✅ Completed CurrencySeed - Progress: 2/Total")
	s.progressBar.Add(1)

	s.progressBar.Describe("📋 Seeding subscription plans...")
	if err := s.modelcore.SubscriptionPlanSeed(ctx); err != nil {
		return err
	}
	s.provider.Service.Logger.Info("✅ Completed SubscriptionPlanSeed - Progress: 3/Total")
	s.progressBar.Add(1)

	s.progressBar.Describe(fmt.Sprintf("👤 Creating %d users...", int(multiplier)*1))
	if err := s.SeedUsers(ctx, multiplier); err != nil {
		return err
	}
	s.provider.Service.Logger.Info("✅ Completed SeedUsers - Progress: SeedUsers+/Total")
	s.progressBar.Add(1)

	s.progressBar.Describe(fmt.Sprintf("🏢 Creating organizations & branches (multiplier: %d)...", multiplier))
	if err := s.SeedOrganization(ctx, multiplier); err != nil {
		return err
	}
	s.provider.Service.Logger.Info("✅ Completed SeedOrganization - Progress: SeedOrg+/Total")
	s.progressBar.Add(1)

	s.progressBar.Describe("👥 Assigning employees to organizations...")
	if err := s.SeedEmployees(ctx, multiplier); err != nil {
		return err
	}
	s.provider.Service.Logger.Info("✅ Completed SeedEmployees - Progress: SeedEmp+/Total")
	s.progressBar.Add(1)

	s.progressBar.Describe("📋 Creating member profiles...")
	if err := s.SeedMemberProfiles(ctx, multiplier); err != nil {
		return err
	}
	s.provider.Service.Logger.Info("✅ Completed SeedMemberProfiles - Progress: SeedMembers+/Total")
	s.progressBar.Add(1)

	// Finish overall progress bar
	s.progressBar.Finish()
	return nil
}

func (s *Seeder) SeedOrganization(ctx context.Context, multiplier int32) error {
	orgs, err := s.modelcore.OrganizationManager.List(ctx)
	if err != nil {
		return err
	}
	if len(orgs) > 0 {
		s.progressBar.Describe("🏢 Organizations already exist, skipping organization creation...")
		// Skip all organization-related adds by advancing the expected count
		numUsers := int(multiplier) * 1
		numOrgsPerUser := int(multiplier) * 1
		numBranchesPerOrg := int(multiplier) * 1
		skipCount := (numUsers * 1) + // per user processing
			(numUsers * numOrgsPerUser * 3) + // per org operations
			(numUsers * numOrgsPerUser * numBranchesPerOrg * 6) // per branch operations
		s.provider.Service.Logger.Info(fmt.Sprintf("⏭️ Skipping %d organization operations (organizations already exist)", skipCount))
		s.progressBar.Add(skipCount)
		return nil
	}

	numOrgsPerUser := int(multiplier) * 1
	users, err := s.modelcore.UserManager.List(ctx)
	if err != nil {
		return err
	}
	subscriptions, err := s.modelcore.SubscriptionPlanManager.List(ctx)
	if err != nil {
		return err
	}
	categories, err := s.modelcore.CategoryManager.List(ctx)
	if err != nil {
		return err
	}

	for _, user := range users {
		s.provider.Service.Logger.Info(fmt.Sprintf("🏢 Processing organizations for user: %s %s", *user.FirstName, *user.LastName))
		s.progressBar.Describe(fmt.Sprintf("🏢 Processing organizations for user: %s %s", *user.FirstName, *user.LastName))
		s.progressBar.Add(1)
		for j := 0; j < numOrgsPerUser; j++ {
			s.provider.Service.Logger.Info(fmt.Sprintf("🏭 Setting up organization %d/%d for user %s", j+1, numOrgsPerUser, *user.FirstName))
			s.progressBar.Describe("🏭 Setting up organization...")
			s.progressBar.Add(1)
			sub := subscriptions[j%len(subscriptions)]
			subscriptionEndDate := time.Now().Add(30 * 24 * time.Hour)
			orgMedia, err := s.createImageMedia(ctx, "Organization")
			if err != nil {
				return eris.Wrap(err, "failed to create organization media")
			}
			s.provider.Service.Logger.Info("📸 Created organization media")
			s.progressBar.Describe("📸 Created organization media")
			s.progressBar.Add(1)
			organization := &modelcore.Organization{
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

			if err := s.modelcore.OrganizationManager.Create(ctx, organization); err != nil {
				return err
			}
			s.provider.Service.Logger.Info(fmt.Sprintf("🏢 Created organization: %s", organization.Name))
			s.progressBar.Describe(fmt.Sprintf("🏢 Created organization: %s", organization.Name))
			s.progressBar.Add(1)

			// Add categories to organization
			for _, category := range categories {
				if err := s.modelcore.OrganizationCategoryManager.Create(ctx, &modelcore.OrganizationCategory{
					CreatedAt:      time.Now().UTC(),
					UpdatedAt:      time.Now().UTC(),
					OrganizationID: &organization.ID,
					CategoryID:     &category.ID,
				}); err != nil {
					return err
				}
			}

			numBranches := int(multiplier) * 1

			for k := 0; k < numBranches; k++ {
				branchMedia, err := s.createImageMedia(ctx, "Organization")
				if err != nil {
					return eris.Wrap(err, "failed to create organization media")
				}
				s.provider.Service.Logger.Info(fmt.Sprintf("📸 Created branch media %d/%d", k+1, numBranches))
				s.progressBar.Describe("📸 Created branch media")
				s.progressBar.Add(1)
				branch := &modelcore.Branch{
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
				if err := s.modelcore.BranchManager.Create(ctx, branch); err != nil {
					return err
				}
				s.provider.Service.Logger.Info(fmt.Sprintf("🏪 Created branch: %s for %s", branch.Name, organization.Name))
				s.progressBar.Describe(fmt.Sprintf("🏪 Created branch: %s for %s", branch.Name, organization.Name))
				s.progressBar.Add(1)
				currency, err := s.modelcore.CurrencyFindByAlpha2(ctx, branch.CountryCode)
				if err != nil {
					return eris.Wrap(err, "failed to find currency for account seeding")
				}
				// Create default branch settings for each branch
				branchSetting := &modelcore.BranchSetting{
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

				if err := s.modelcore.BranchSettingManager.Create(ctx, branchSetting); err != nil {
					return err
				}
				s.provider.Service.Logger.Info(fmt.Sprintf("⚙️ Created settings for branch: %s", branch.Name))
				s.progressBar.Describe(fmt.Sprintf("⚙️ Created settings for branch: %s", branch.Name))
				s.progressBar.Add(1)

				// Create the owner relationship for the user who created the organization
				developerKey, err := s.provider.Service.Security.GenerateUUIDv5(ctx, fmt.Sprintf("%s-%s-%s", user.ID, organization.ID, branch.ID))
				if err != nil {
					return err
				}

				ownerOrganization := &modelcore.UserOrganization{
					CreatedAt:                time.Now().UTC(),
					CreatedByID:              user.ID,
					UpdatedAt:                time.Now().UTC(),
					UpdatedByID:              user.ID,
					BranchID:                 &branch.ID,
					OrganizationID:           organization.ID,
					UserID:                   user.ID,
					UserType:                 modelcore.UserOrganizationTypeOwner,
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
					Status:                   modelcore.UserOrganizationStatusOffline,
					LastOnlineAt:             time.Now().UTC(),
				}

				if err := s.modelcore.UserOrganizationManager.Create(ctx, ownerOrganization); err != nil {
					return err
				}
				s.provider.Service.Logger.Info(fmt.Sprintf("👑 Created owner relationship for %s %s", *user.FirstName, *user.LastName))
				s.progressBar.Describe(fmt.Sprintf("👑 Created owner relationship for %s %s", *user.FirstName, *user.LastName))
				s.progressBar.Add(1)

				// Run organization seeder for accounting setup
				tx := s.provider.Service.Database.Client().Begin()
				if tx.Error != nil {
					tx.Rollback()
					return err
				}
				if err := s.modelcore.OrganizationSeeder(ctx, tx, user.ID, organization.ID, branch.ID); err != nil {
					tx.Rollback()
					return err
				}
				if err := tx.Commit().Error; err != nil {
					return err
				}
				s.provider.Service.Logger.Info("💼 Setup accounting system for organization")
				s.progressBar.Describe("💼 Setup accounting system for organization")
				s.progressBar.Add(1)

				numInvites := int(multiplier) * 1
				for m := 0; m < numInvites; m++ {
					userType := modelcore.UserOrganizationTypeMember
					if m%2 == 0 {
						userType = modelcore.UserOrganizationTypeEmployee
					}
					invitationCode := &modelcore.InvitationCode{
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
					if err := s.modelcore.InvitationCodeManager.Create(ctx, invitationCode); err != nil {
						return err
					}
				}
				s.provider.Service.Logger.Info(fmt.Sprintf("✉️ Created %d invitation codes for %s", numInvites, organization.Name))
				s.progressBar.Describe(fmt.Sprintf("✉️ Created %d invitation codes for %s", numInvites, organization.Name))
				s.progressBar.Add(1)

			}
		}
	}
	return nil
}

func (s *Seeder) SeedEmployees(ctx context.Context, multiplier int32) error {

	// Get all organizations and their branches
	organizations, err := s.modelcore.OrganizationManager.List(ctx)
	if err != nil {
		return err
	}
	s.provider.Service.Logger.Info(fmt.Sprintf("📋 Listed %d organizations for employee processing", len(organizations)))
	s.progressBar.Add(1)
	// Get all users to use as employees
	users, err := s.modelcore.UserManager.List(ctx)
	if err != nil {
		return err
	}
	s.provider.Service.Logger.Info(fmt.Sprintf("👥 Listed %d users for employee assignment", len(users)))
	s.progressBar.Add(1)
	if len(users) == 0 || len(organizations) == 0 {
		s.provider.Service.Logger.Warn("No users or organizations found for employee seeding")
		return nil
	}

	// Create cross-employment: users work as employees in other users' organizations
	for _, org := range organizations {
		s.provider.Service.Logger.Info(fmt.Sprintf("🏢 Processing employees for organization: %s", org.Name))
		s.progressBar.Add(1)
		// Get branches for this organization
		branches, err := s.modelcore.BranchManager.Find(ctx, &modelcore.Branch{
			OrganizationID: org.ID,
		})
		if err != nil {
			continue
		}
		s.provider.Service.Logger.Info(fmt.Sprintf("🏪 Found %d branches for organization: %s", len(branches), org.Name))
		s.progressBar.Add(1)

		// Find potential employees (users who don't own this organization)
		potentialEmployees := make([]*modelcore.User, 0)
		for _, user := range users {
			if user.ID != org.CreatedByID { // Don't make the owner an employee of their own organization
				potentialEmployees = append(potentialEmployees, user)
			}
		}
		s.provider.Service.Logger.Info(fmt.Sprintf("👤 Found %d potential employees for organization: %s", len(potentialEmployees), org.Name))
		s.progressBar.Add(1)

		if len(potentialEmployees) == 0 {
			continue
		}

		employeeIndex := 0
		for _, branch := range branches {
			// Check how many employees already exist for this branch (excluding owner)
			existingEmployees, err := s.modelcore.Employees(ctx, org.ID, branch.ID)
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
				existingAssociation, err := s.modelcore.UserOrganizationManager.Count(ctx, &modelcore.UserOrganization{
					UserID:         selectedUser.ID,
					OrganizationID: org.ID,
					BranchID:       &branch.ID,
				})
				if err != nil || existingAssociation > 0 {
					// Skip if user is already associated with this branch
					continue
				}

				// Check if user can join as employee (ensure they're not already a member of this organization)
				if !s.modelcore.UserOrganizationEmployeeCanJoin(ctx, selectedUser.ID, org.ID, branch.ID) {
					continue
				}

				// Generate developer key
				developerKey, err := s.provider.Service.Security.GenerateUUIDv5(ctx, fmt.Sprintf("emp-%s-%s-%s", selectedUser.ID, org.ID, branch.ID))
				if err != nil {
					return err
				}

				// Create employee user organization record
				employeeOrg := &modelcore.UserOrganization{
					CreatedAt:                time.Now().UTC(),
					CreatedByID:              org.CreatedByID, // Created by the organization owner
					UpdatedAt:                time.Now().UTC(),
					UpdatedByID:              org.CreatedByID,
					BranchID:                 &branch.ID,
					OrganizationID:           org.ID,
					UserID:                   selectedUser.ID,
					UserType:                 modelcore.UserOrganizationTypeEmployee,
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
					Status:                   modelcore.UserOrganizationStatusOffline,
					LastOnlineAt:             time.Now().UTC(),
				}

				if err := s.modelcore.UserOrganizationManager.Create(ctx, employeeOrg); err != nil {
					s.provider.Service.Logger.Error(fmt.Sprintf("Failed to create employee %s %s for branch %s: %v",
						*selectedUser.FirstName, *selectedUser.LastName, branch.Name, err))
					continue
				}
				s.provider.Service.Logger.Info(fmt.Sprintf("✅ Created employee: %s %s for branch %s", *selectedUser.FirstName, *selectedUser.LastName, branch.Name))
				s.progressBar.Add(1)

			}
		}
	}

	return nil
}

func (s *Seeder) SeedUsers(ctx context.Context, multiplier int32) error {
	users, err := s.modelcore.UserManager.List(ctx)
	if err != nil {
		return err
	}
	if len(users) >= 1 {
		s.progressBar.Describe("👤 Users already exist, skipping user creation...")
		// Skip the user creation adds by advancing the expected count
		numUsers := int(multiplier) * 1
		s.provider.Service.Logger.Info(fmt.Sprintf("⏭️ Skipping %d user creations (users already exist)", numUsers))
		s.progressBar.Add(numUsers)
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

		user := &modelcore.User{
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
		if err := s.modelcore.UserManager.Create(ctx, user); err != nil {
			return err
		}
		s.provider.Service.Logger.Info(fmt.Sprintf("✅ Created user: %s %s (%s) - Progress: User %d", *user.FirstName, *user.LastName, user.Email, i+1))
		s.progressBar.Describe(fmt.Sprintf("👤 Created user: %s (%s)", *user.FirstName+" "+*user.LastName, user.Email))
		s.progressBar.Add(1)
	}
	return nil
}

func (s *Seeder) SeedMemberProfiles(ctx context.Context, multiplier int32) error {
	// Check if member profiles already exist
	profiles, err := s.modelcore.MemberProfileManager.List(ctx)
	if err != nil {
		return err
	}
	if len(profiles) > 0 {
		s.progressBar.Describe("👥 Member profiles already exist, skipping member profile creation...")
		// Skip member profile adds
		numUsers := int(multiplier) * 1
		numOrgsPerUser := int(multiplier) * 1
		numBranchesPerOrg := int(multiplier) * 1
		numMembersPerBranch := int(multiplier) * 1
		skipCount := (numUsers * numOrgsPerUser * 1) + // per org processing
			(numUsers * numOrgsPerUser * numBranchesPerOrg * numMembersPerBranch * 2) // per member operations
		s.provider.Service.Logger.Info(fmt.Sprintf("⏭️ Skipping %d member profile operations (profiles already exist)", skipCount))
		s.progressBar.Add(skipCount)
		return nil
	}
	// Get existing organizations and branches to seed member profiles for
	organizations, err := s.modelcore.OrganizationManager.List(ctx)
	if err != nil {
		return err
	}

	// Get existing users to associate with member profiles
	users, err := s.modelcore.UserManager.List(ctx)
	if err != nil {
		return err
	}

	if len(organizations) == 0 || len(users) == 0 {
		s.provider.Service.Logger.Warn("No organizations or users found, skipping member profile seeding")
		return nil
	}

	for _, org := range organizations {
		// Get branches for this organization
		s.provider.Service.Logger.Info(fmt.Sprintf("👥 Processing member profiles for organization: %s", org.Name))
		s.progressBar.Describe(fmt.Sprintf("👥 Processing member profiles for organization: %s", org.Name))
		s.progressBar.Add(1)
		branches, err := s.modelcore.BranchManager.Find(ctx, &modelcore.Branch{
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
				memberProfile := &modelcore.MemberProfile{
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

				if err := s.modelcore.MemberProfileManager.Create(ctx, memberProfile); err != nil {
					s.provider.Service.Logger.Error(fmt.Sprintf("Failed to create member profile %s: %v", fullName, err))
					continue
				}
				s.provider.Service.Logger.Info(fmt.Sprintf("👤 Created member: %s (%s) for branch: %s", fullName, passbook, branch.Name))
				s.progressBar.Describe(fmt.Sprintf("👤 Created member: %s (%s)", fullName, passbook))
				s.progressBar.Add(1)

				// Create sample member address
				memberAddress := &modelcore.MemberAddress{
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

				if err := s.modelcore.MemberAddressManager.Create(ctx, memberAddress); err != nil {
					s.provider.Service.Logger.Error(fmt.Sprintf("Failed to create member address for %s: %v", fullName, err))
				}
				s.provider.Service.Logger.Info(fmt.Sprintf("📍 Added address for member: %s in branch: %s", fullName, branch.Name))
				s.progressBar.Describe(fmt.Sprintf("📍 Added address for member: %s", fullName))
				s.progressBar.Add(1)

			}
		}
	}
	return nil
}
