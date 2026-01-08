package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/google/uuid"
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

func (c *Core) Seed(ctx context.Context, multiplier int32) error {
	if multiplier <= 0 {
		return nil
	}
	if err := c.loadImagePaths(); err != nil {
		return eris.Wrap(err, "failed to load image paths")
	}
	if err := c.GlobalSeeder(ctx); err != nil {
		return err
	}
	if err := c.SeedUsers(ctx, multiplier); err != nil {
		return err
	}
	if err := c.SeedOrganization(ctx, multiplier); err != nil {
		return err
	}
	if err := c.SeedEmployees(ctx, multiplier); err != nil {
		return err
	}
	if err := c.SeedMemberProfiles(ctx, multiplier); err != nil {
		return err
	}
	return nil
}

func (s *Core) SeedOrganization(ctx context.Context, multiplier int32) error {
	orgs, err := s.OrganizationManager().List(ctx)
	if err != nil {
		return err
	}
	if len(orgs) > 0 {
		return nil
	}

	numOrgsPerUser := int(multiplier) * 1
	users, err := s.UserManager().List(ctx)
	if err != nil {
		return err
	}
	subscriptions, err := s.SubscriptionPlanManager().List(ctx)
	if err != nil {
		return err
	}
	categories, err := s.CategoryManager().List(ctx)
	if err != nil {
		return err
	}

	for _, user := range users {
		for j := range numOrgsPerUser {

			sub := subscriptions[j%len(subscriptions)]
			subscriptionEndDate := time.Now().Add(30 * 24 * time.Hour)
			orgMedia, err := s.createImageMedia(ctx, "Organization")
			if err != nil {
				return eris.Wrap(err, "failed to create organization media")
			}
			organization := &Organization{
				CreatedAt:                           time.Now().UTC(),
				CreatedByID:                         user.ID,
				UpdatedAt:                           time.Now().UTC(),
				UpdatedByID:                         user.ID,
				Name:                                s.faker.Company().Name(),
				Address:                             handlers.Ptr(s.faker.Address().Address()),
				Email:                               handlers.Ptr(s.faker.Internet().Email()),
				ContactNumber:                       handlers.Ptr(fmt.Sprintf("+6391%08d", s.faker.IntBetween(10000000, 99999999))),
				Description:                         handlers.Ptr(s.faker.Lorem().Paragraph(3)),
				Color:                               handlers.Ptr(s.faker.Color().Hex()),
				TermsAndConditions:                  handlers.Ptr(s.faker.Lorem().Paragraph(5)),
				PrivacyPolicy:                       handlers.Ptr(s.faker.Lorem().Paragraph(5)),
				CookiePolicy:                        handlers.Ptr(s.faker.Lorem().Paragraph(5)),
				RefundPolicy:                        handlers.Ptr(s.faker.Lorem().Paragraph(5)),
				UserAgreement:                       handlers.Ptr(s.faker.Lorem().Paragraph(5)),
				IsPrivate:                           s.faker.Bool(),
				MediaID:                             &orgMedia.ID,
				CoverMediaID:                        &orgMedia.ID,
				SubscriptionPlanMaxBranches:         sub.MaxBranches,
				SubscriptionPlanMaxEmployees:        sub.MaxEmployees,
				SubscriptionPlanMaxMembersPerBranch: sub.MaxMembersPerBranch,
				SubscriptionPlanID:                  &sub.ID,
				SubscriptionStartDate:               time.Now().UTC(),
				SubscriptionEndDate:                 subscriptionEndDate,
				InstagramLink:                       handlers.Ptr(s.faker.Internet().URL()),
				FacebookLink:                        handlers.Ptr(s.faker.Internet().URL()),
				YoutubeLink:                         handlers.Ptr(s.faker.Internet().URL()),
				PersonalWebsiteLink:                 handlers.Ptr(s.faker.Internet().URL()),
				XLink:                               handlers.Ptr(s.faker.Internet().URL()),
			}

			if err := s.OrganizationManager().Create(ctx, organization); err != nil {
				return err
			}
			for _, category := range categories {
				if err := s.OrganizationCategoryManager().Create(ctx, &OrganizationCategory{
					CreatedAt:      time.Now().UTC(),
					UpdatedAt:      time.Now().UTC(),
					OrganizationID: &organization.ID,
					CategoryID:     &category.ID,
				}); err != nil {
					return err
				}
			}

			numBranches := int(multiplier) * 1

			for k := range numBranches {
				branchMedia, err := s.createImageMedia(ctx, "Organization")
				if err != nil {
					return eris.Wrap(err, "failed to create organization media")
				}

				currency, err := s.CurrencyFindByAlpha2(ctx, country_code)
				if err != nil {
					return eris.Wrap(err, "failed to find currency for account seeding")
				}
				c := cities[s.faker.IntBetween(0, len(cities)-1)]
				Latitude := c.Lat + (float64(s.faker.IntBetween(-50, 50)) / 1000.0)
				Longitude := c.Lng + (float64(s.faker.IntBetween(-50, 50)) / 1000.0)
				branch := &Branch{
					CreatedAt:               time.Now().UTC(),
					CreatedByID:             user.ID,
					UpdatedAt:               time.Now().UTC(),
					UpdatedByID:             user.ID,
					OrganizationID:          organization.ID,
					Type:                    []string{"main", "satellite", "branch"}[k%3],
					Name:                    s.faker.Company().Name(),
					Email:                   s.faker.Internet().Email(),
					Address:                 s.faker.Address().Address(),
					Province:                s.faker.Address().State(),
					City:                    s.faker.Address().City(),
					Region:                  s.faker.Address().State(),
					Barangay:                s.faker.Address().StreetName(),
					PostalCode:              s.faker.Address().PostCode(),
					CurrencyID:              &currency.ID,
					ContactNumber:           handlers.Ptr(fmt.Sprintf("+6391%08d", s.faker.IntBetween(10000000, 99999999))),
					MediaID:                 &branchMedia.ID,
					Latitude:                &Latitude,
					Longitude:               &Longitude,
					TaxIdentificationNumber: handlers.Ptr(fmt.Sprintf("%09d", s.faker.IntBetween(100000000, 999999999))),
				}
				if err := s.BranchManager().Create(ctx, branch); err != nil {
					return err
				}
				branchSetting := &BranchSetting{
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
					BranchID:  branch.ID,

					WithdrawAllowUserInput: true,
					WithdrawPrefix:         s.faker.Lorem().Word(),
					WithdrawORStart:        1,
					WithdrawORCurrent:      1,
					WithdrawOREnd:          999999,
					WithdrawORIteration:    1,
					WithdrawORUnique:       true,
					WithdrawUseDateOR:      false,

					DepositAllowUserInput: true,
					DepositPrefix:         s.faker.Lorem().Word(),
					DepositORStart:        1,
					DepositORCurrent:      1,
					DepositOREnd:          999999,
					DepositORIteration:    1,
					DepositORUnique:       true,
					DepositUseDateOR:      false,

					LoanAllowUserInput: true,
					LoanPrefix:         s.faker.Lorem().Word(),
					LoanORStart:        1,
					LoanORCurrent:      1,
					LoanOREnd:          999999,
					LoanORIteration:    1,
					LoanORUnique:       true,
					LoanUseDateOR:      false,

					CheckVoucherAllowUserInput: true,
					CheckVoucherPrefix:         s.faker.Lorem().Word(),
					CheckVoucherORStart:        1,
					CheckVoucherORCurrent:      1,
					CheckVoucherOREnd:          999999,
					CheckVoucherORIteration:    1,
					CheckVoucherORUnique:       true,
					CheckVoucherUseDateOR:      false,

					DefaultMemberTypeID:       nil,
					LoanAppliedEqualToBalance: true,
					CurrencyID:                currency.ID,
				}

				if err := s.BranchSettingManager().Create(ctx, branchSetting); err != nil {
					return err
				}

				developerKey, err := s.provider.Service.Security.GenerateUUIDv5(ctx, fmt.Sprintf("%s-%s-%s", user.ID, organization.ID, branch.ID))
				if err != nil {
					return err
				}

				ownerOrganization := &UserOrganization{
					CreatedAt:                time.Now().UTC(),
					CreatedByID:              user.ID,
					UpdatedAt:                time.Now().UTC(),
					UpdatedByID:              user.ID,
					BranchID:                 &branch.ID,
					OrganizationID:           organization.ID,
					UserID:                   user.ID,
					UserType:                 UserOrganizationTypeOwner,
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
					Status:                   UserOrganizationStatusOffline,
					LastOnlineAt:             time.Now().UTC(),
				}

				if err := s.UserOrganizationManager().Create(ctx, ownerOrganization); err != nil {
					return err
				}

				tx, endTx := s.provider.Service.Database.StartTransaction(ctx)
				if err := s.OrganizationSeeder(ctx, tx, user.ID, organization.ID, branch.ID); err != nil {
					return endTx(err)
				}
				if err := endTx(nil); err != nil {
					return err
				}
				numInvites := int(multiplier) * 1
				for m := range numInvites {
					userType := UserOrganizationTypeMember
					if m%2 == 0 {
						userType = UserOrganizationTypeEmployee
					}
					invitationCode := &InvitationCode{
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
					if err := s.InvitationCodeManager().Create(ctx, invitationCode); err != nil {
						return err
					}
				}

			}
		}
	}
	return nil
}

func (s *Core) SeedEmployees(ctx context.Context, multiplier int32) error {

	organizations, err := s.OrganizationManager().List(ctx)
	if err != nil {
		return err
	}
	users, err := s.UserManager().List(ctx)
	if err != nil {
		return err
	}
	if len(users) == 0 || len(organizations) == 0 {
		return nil
	}

	for _, org := range organizations {
		branches, err := s.BranchManager().Find(ctx, &Branch{
			OrganizationID: org.ID,
		})
		if err != nil {
			continue
		}

		potentialEmployees := make([]*User, 0)
		for _, user := range users {
			if user.ID != org.CreatedByID {
				potentialEmployees = append(potentialEmployees, user)
			}
		}
		if len(potentialEmployees) == 0 {
			continue
		}

		employeeIndex := 0
		for _, branch := range branches {
			existingEmployees, err := s.Employees(ctx, org.ID, branch.ID)
			if err != nil {
				continue
			}

			numEmployeesToCreate := int(multiplier) * 1

			numEmployeesToCreate = min(numEmployeesToCreate, len(potentialEmployees))

			maxPerBranch := 3 * int(multiplier)
			if len(existingEmployees)+numEmployeesToCreate > maxPerBranch {
				numEmployeesToCreate = maxPerBranch - len(existingEmployees)
			}

			if numEmployeesToCreate <= 0 {
				continue
			}

			for i := 0; i < numEmployeesToCreate; i++ {
				selectedUser := potentialEmployees[employeeIndex%len(potentialEmployees)]
				employeeIndex++

				existingAssociation, err := s.UserOrganizationManager().Count(ctx, &UserOrganization{
					UserID:         selectedUser.ID,
					OrganizationID: org.ID,
					BranchID:       &branch.ID,
				})
				if err != nil || existingAssociation > 0 {
					continue
				}

				if !s.UserOrganizationEmployeeCanJoin(ctx, selectedUser.ID, org.ID, branch.ID) {
					continue
				}

				developerKey, err := s.provider.Service.Security.GenerateUUIDv5(ctx, fmt.Sprintf("emp-%s-%s-%s", selectedUser.ID, org.ID, branch.ID))
				if err != nil {
					return err
				}

				employeeOrg := &UserOrganization{
					CreatedAt:                time.Now().UTC(),
					CreatedByID:              org.CreatedByID, // Created by the organization owner
					UpdatedAt:                time.Now().UTC(),
					UpdatedByID:              org.CreatedByID,
					BranchID:                 &branch.ID,
					OrganizationID:           org.ID,
					UserID:                   selectedUser.ID,
					UserType:                 UserOrganizationTypeEmployee,
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
					Status:                   UserOrganizationStatusOffline,
					LastOnlineAt:             time.Now().UTC(),
				}

				if err := s.UserOrganizationManager().Create(ctx, employeeOrg); err != nil {
					continue
				}
			}
		}
	}

	return nil
}

func (s *Core) SeedUsers(ctx context.Context, multiplier int32) error {
	users, err := s.UserManager().List(ctx)
	if err != nil {
		return err
	}
	if len(users) > 0 {
		return nil
	}

	basePassword := "sample-hello-world-12345"
	hashedPassword, err := s.provider.Service.Security.HashPassword(ctx, basePassword)
	if err != nil {
		return err
	}

	baseNumUsers := 1
	numUsers := int(multiplier) * baseNumUsers

	for i := range numUsers {
		firstName := s.faker.Person().FirstName()
		middleName := s.faker.Person().LastName()[:1] // Simulate middle initial
		lastName := s.faker.Person().LastName()
		suffix := s.faker.Person().Suffix()
		fullName := fmt.Sprintf("%s %s %s %s", firstName, middleName, lastName, suffix)
		birthdate := time.Now().AddDate(-25-s.faker.IntBetween(0, 40), -s.faker.IntBetween(0, 11), -s.faker.IntBetween(0, 30))
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

		user := &User{
			MediaID:           &userSharedMedia.ID,
			Email:             email,
			Password:          hashedPassword,
			Birthdate:         &birthdate,
			Username:          s.faker.Internet().User(),
			FullName:          fullName,
			FirstName:         handlers.Ptr(firstName),
			MiddleName:        handlers.Ptr(middleName),
			LastName:          handlers.Ptr(lastName),
			Suffix:            handlers.Ptr(suffix),
			ContactNumber:     fmt.Sprintf("+6391%08d", s.faker.IntBetween(10000000, 99999999)),
			IsEmailVerified:   true,
			IsContactVerified: true,
			CreatedAt:         time.Now().UTC(),
			UpdatedAt:         time.Now().UTC(),
		}
		if err := s.UserManager().Create(ctx, user); err != nil {
			return err
		}

	}
	return nil
}

func (s *Core) SeedMemberProfiles(ctx context.Context, multiplier int32) error {
	profiles, err := s.MemberProfileManager().List(ctx)
	if err != nil {
		return err
	}
	if len(profiles) > 0 {
		return nil
	}

	organizations, err := s.OrganizationManager().List(ctx)
	if err != nil {
		return err
	}

	users, err := s.UserManager().List(ctx)
	if err != nil {
		return err
	}

	if len(organizations) == 0 || len(users) == 0 {
		return nil
	}

	for _, org := range organizations {

		branches, err := s.BranchManager().Find(ctx, &Branch{
			OrganizationID: org.ID,
		})
		if err != nil {
			continue
		}

		for _, branch := range branches {
			numMembers := int(multiplier) * 1
			numMembers = min(numMembers, len(users))

			for i := 0; i < numMembers; i++ {
				firstName := s.faker.Person().FirstName()
				middleName := s.faker.Person().LastName()[:1]
				lastName := s.faker.Person().LastName()
				fullName := fmt.Sprintf("%s %s %s", firstName, middleName, lastName)

				age := 25 + (i % 40)
				birthDate := time.Now().AddDate(-age, 0, 0)

				passbook := fmt.Sprintf("PB-%s-%04d", branch.Name[:min(3, len(branch.Name))], i+1)

				memberProfile := &MemberProfile{
					CreatedAt:             time.Now().UTC(),
					CreatedByID:           &org.CreatedByID,
					UpdatedAt:             time.Now().UTC(),
					UpdatedByID:           &org.CreatedByID,
					OrganizationID:        org.ID,
					BranchID:              branch.ID,
					UserID:                &users[i%len(users)].ID, // Rotate through available users
					FirstName:             firstName,
					MiddleName:            middleName,
					LastName:              lastName,
					FullName:              fullName,
					BirthDate:             &birthDate,
					Status:                MemberStatusPending,
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

				if err := s.MemberProfileManager().Create(ctx, memberProfile); err != nil {
					continue
				}

				memberAddress := &MemberAddress{
					CreatedAt:       time.Now().UTC(),
					UpdatedAt:       time.Now().UTC(),
					CreatedByID:     &org.CreatedByID,
					UpdatedByID:     &org.CreatedByID,
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

				if err := s.MemberAddressManager().Create(ctx, memberAddress); err != nil {
					return err
				}

			}
		}
	}
	return nil
}
