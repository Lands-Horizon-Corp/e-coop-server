package core

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
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

func Seed(ctx context.Context, service *horizon.HorizonService, multiplier int32) error {
	if multiplier <= 0 {
		return nil
	}
	images, err := loadImagePaths()
	if err != nil {
		return eris.Wrap(err, "failed to load image paths for seeding")
	}
	if err := GlobalSeeder(ctx, service); err != nil {
		return err
	}
	if err := SeedUsers(ctx, service, images, multiplier); err != nil {
		return err
	}
	if err := SeedOrganization(ctx, service, images, multiplier); err != nil {
		return err
	}
	if err := SeedEmployees(ctx, service, multiplier); err != nil {
		return err
	}
	if err := SeedMemberProfiles(ctx, service, multiplier); err != nil {
		return err
	}
	return nil
}

func SeedOrganization(ctx context.Context, service *horizon.HorizonService, imagePaths []string, multiplier int32) error {
	orgs, err := OrganizationManager(service).List(ctx)
	if err != nil {
		return err
	}
	if len(orgs) > 0 {
		return nil
	}
	numOrgsPerUser := int(multiplier) * 1
	users, err := UserManager(service).List(ctx)
	if err != nil {
		return err
	}
	subscriptions, err := SubscriptionPlanManager(service).List(ctx)
	if err != nil {
		return err
	}
	categories, err := CategoryManager(service).List(ctx)
	if err != nil {
		return err
	}
	for _, user := range users {
		for j := range numOrgsPerUser {

			sub := subscriptions[j%len(subscriptions)]
			subscriptionEndDate := time.Now().Add(30 * 24 * time.Hour)
			orgMedia, err := createImageMedia(ctx, service, imagePaths, "Organization")
			if err != nil {
				return eris.Wrap(err, "failed to create organization media")
			}
			organization := &Organization{
				CreatedAt:                           time.Now().UTC(),
				CreatedByID:                         user.ID,
				UpdatedAt:                           time.Now().UTC(),
				UpdatedByID:                         user.ID,
				Name:                                service.Faker.Company().Name(),
				Address:                             helpers.Ptr(service.Faker.Address().Address()),
				Email:                               helpers.Ptr(service.Faker.Internet().Email()),
				ContactNumber:                       helpers.Ptr(fmt.Sprintf("+6391%08d", service.Faker.IntBetween(10000000, 99999999))),
				Description:                         helpers.Ptr(service.Faker.Lorem().Paragraph(3)),
				Color:                               helpers.Ptr(service.Faker.Color().Hex()),
				TermsAndConditions:                  helpers.Ptr(service.Faker.Lorem().Paragraph(5)),
				PrivacyPolicy:                       helpers.Ptr(service.Faker.Lorem().Paragraph(5)),
				CookiePolicy:                        helpers.Ptr(service.Faker.Lorem().Paragraph(5)),
				RefundPolicy:                        helpers.Ptr(service.Faker.Lorem().Paragraph(5)),
				UserAgreement:                       helpers.Ptr(service.Faker.Lorem().Paragraph(5)),
				IsPrivate:                           service.Faker.Bool(),
				MediaID:                             &orgMedia.ID,
				CoverMediaID:                        &orgMedia.ID,
				SubscriptionPlanMaxBranches:         sub.MaxBranches,
				SubscriptionPlanMaxEmployees:        sub.MaxEmployees,
				SubscriptionPlanMaxMembersPerBranch: sub.MaxMembersPerBranch,
				SubscriptionPlanID:                  &sub.ID,
				SubscriptionStartDate:               time.Now().UTC(),
				SubscriptionEndDate:                 subscriptionEndDate,
				InstagramLink:                       helpers.Ptr(service.Faker.Internet().URL()),
				FacebookLink:                        helpers.Ptr(service.Faker.Internet().URL()),
				YoutubeLink:                         helpers.Ptr(service.Faker.Internet().URL()),
				PersonalWebsiteLink:                 helpers.Ptr(service.Faker.Internet().URL()),
				XLink:                               helpers.Ptr(service.Faker.Internet().URL()),
			}

			if err := OrganizationManager(service).Create(ctx, organization); err != nil {
				return err
			}
			for _, category := range categories {
				if err := OrganizationCategoryManager(service).Create(ctx, &OrganizationCategory{
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
				branchMedia, err := createImageMedia(ctx, service, imagePaths, "Organization")
				if err != nil {
					return eris.Wrap(err, "failed to create organization media")
				}

				currency, err := CurrencyFindByAlpha2(ctx, service, country_code)
				if err != nil {
					return eris.Wrap(err, "failed to find currency for account seeding")
				}
				c := cities[service.Faker.IntBetween(0, len(cities)-1)]
				Latitude := c.Lat + (float64(service.Faker.IntBetween(-50, 50)) / 1000.0)
				Longitude := c.Lng + (float64(service.Faker.IntBetween(-50, 50)) / 1000.0)
				branch := &Branch{
					CreatedAt:               time.Now().UTC(),
					CreatedByID:             user.ID,
					UpdatedAt:               time.Now().UTC(),
					UpdatedByID:             user.ID,
					OrganizationID:          organization.ID,
					Type:                    []string{"main", "satellite", "branch"}[k%3],
					Name:                    service.Faker.Company().Name(),
					Email:                   service.Faker.Internet().Email(),
					Address:                 service.Faker.Address().Address(),
					Province:                service.Faker.Address().State(),
					City:                    service.Faker.Address().City(),
					Region:                  service.Faker.Address().State(),
					Barangay:                service.Faker.Address().StreetName(),
					PostalCode:              service.Faker.Address().PostCode(),
					CurrencyID:              &currency.ID,
					ContactNumber:           helpers.Ptr(fmt.Sprintf("+6391%08d", service.Faker.IntBetween(10000000, 99999999))),
					MediaID:                 &branchMedia.ID,
					Latitude:                &Latitude,
					Longitude:               &Longitude,
					TaxIdentificationNumber: helpers.Ptr(fmt.Sprintf("%09d", service.Faker.IntBetween(100000000, 999999999))),
				}
				if err := BranchManager(service).Create(ctx, branch); err != nil {
					return err
				}
				branchSetting := &BranchSetting{
					CreatedAt: time.Now().UTC(),
					UpdatedAt: time.Now().UTC(),
					BranchID:  branch.ID,

					WithdrawAllowUserInput: true,
					WithdrawPrefix:         service.Faker.Lorem().Word(),
					WithdrawORStart:        1,
					WithdrawORCurrent:      1,
					WithdrawOREnd:          999999,
					WithdrawORIteration:    1,
					WithdrawORUnique:       true,
					WithdrawUseDateOR:      false,

					DepositAllowUserInput: true,
					DepositPrefix:         service.Faker.Lorem().Word(),
					DepositORStart:        1,
					DepositORCurrent:      1,
					DepositOREnd:          999999,
					DepositORIteration:    1,
					DepositORUnique:       true,
					DepositUseDateOR:      false,

					LoanAllowUserInput: true,
					LoanPrefix:         service.Faker.Lorem().Word(),
					LoanORStart:        1,
					LoanORCurrent:      1,
					LoanOREnd:          999999,
					LoanORIteration:    1,
					LoanORUnique:       true,
					LoanUseDateOR:      false,

					CheckVoucherAllowUserInput: true,
					CheckVoucherPrefix:         service.Faker.Lorem().Word(),
					CheckVoucherORStart:        1,
					CheckVoucherORCurrent:      1,
					CheckVoucherOREnd:          999999,
					CheckVoucherORIteration:    1,
					CheckVoucherORUnique:       true,
					CheckVoucherUseDateOR:      false,

					DefaultMemberTypeID:       nil,
					DefaultMemberGenderID:     nil,
					LoanAppliedEqualToBalance: true,
					CurrencyID:                currency.ID,
				}

				if err := BranchSettingManager(service).Create(ctx, branchSetting); err != nil {
					return err
				}

				developerKey, err := service.Security.GenerateUUIDv5(fmt.Sprintf("%s-%s-%s", user.ID, organization.ID, branch.ID))
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
					Description:              service.Faker.Lorem().Sentence(5),
					ApplicationDescription:   service.Faker.Lorem().Sentence(3),
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

				if err := UserOrganizationManager(service).Create(ctx, ownerOrganization); err != nil {
					return err
				}

				tx, endTx := service.Database.StartTransaction(ctx)
				if err := OrganizationSeeder(ctx, service, tx, user.ID, organization.ID, branch.ID); err != nil {
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
						Description:    service.Faker.Lorem().Sentence(3),
					}
					if err := InvitationCodeManager(service).Create(ctx, invitationCode); err != nil {
						return err
					}
				}

			}
		}
	}
	return nil
}

func SeedEmployees(ctx context.Context, service *horizon.HorizonService, multiplier int32) error {

	organizations, err := OrganizationManager(service).List(ctx)
	if err != nil {
		return err
	}
	users, err := UserManager(service).List(ctx)
	if err != nil {
		return err
	}
	if len(users) == 0 || len(organizations) == 0 {
		return nil
	}

	for _, org := range organizations {
		branches, err := BranchManager(service).Find(ctx, &Branch{
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
			existingEmployees, err := Employees(ctx, service, org.ID, branch.ID)
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

				existingAssociation, err := UserOrganizationManager(service).Count(ctx, &UserOrganization{
					UserID:         selectedUser.ID,
					OrganizationID: org.ID,
					BranchID:       &branch.ID,
				})
				if err != nil || existingAssociation > 0 {
					continue
				}

				if !UserOrganizationEmployeeCanJoin(ctx, service, selectedUser.ID, org.ID, branch.ID) {
					continue
				}

				developerKey, err := service.Security.GenerateUUIDv5(fmt.Sprintf("emp-%s-%s-%s", selectedUser.ID, org.ID, branch.ID))
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
					Description:              service.Faker.Lorem().Sentence(5),
					ApplicationDescription:   service.Faker.Lorem().Sentence(3),
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

				if err := UserOrganizationManager(service).Create(ctx, employeeOrg); err != nil {
					continue
				}
			}
		}
	}

	return nil
}

func SeedUsers(ctx context.Context, service *horizon.HorizonService, imagePaths []string, multiplier int32) error {
	users, err := UserManager(service).List(ctx)
	if err != nil {
		return err
	}
	if len(users) > 0 {
		return nil
	}

	basePassword := "sample-hello-world-12345"
	hashedPassword, err := service.Security.HashPassword(basePassword)
	if err != nil {
		return err
	}

	baseNumUsers := 1
	numUsers := int(multiplier) * baseNumUsers

	for i := range numUsers {
		firstName := service.Faker.Person().FirstName()
		middleName := service.Faker.Person().LastName()[:1] // Simulate middle initial
		lastName := service.Faker.Person().LastName()
		suffix := service.Faker.Person().Suffix()
		fullName := fmt.Sprintf("%s %s %s %s", firstName, middleName, lastName, suffix)
		birthdate := time.Now().AddDate(-25-service.Faker.IntBetween(0, 40), -service.Faker.IntBetween(0, 11), -service.Faker.IntBetween(0, 30))
		userSharedMedia, err := createImageMedia(ctx, service, imagePaths, "User")
		if err != nil {
			return eris.Wrap(err, "failed to create user media")
		}
		var email string
		if i == 0 {
			email = "sample@example.com"
		} else {
			email = service.Faker.Internet().Email()
		}

		user := &User{
			MediaID:           &userSharedMedia.ID,
			Email:             email,
			Password:          hashedPassword,
			Birthdate:         &birthdate,
			Username:          service.Faker.Internet().User(),
			FullName:          fullName,
			FirstName:         helpers.Ptr(firstName),
			MiddleName:        helpers.Ptr(middleName),
			LastName:          helpers.Ptr(lastName),
			Suffix:            helpers.Ptr(suffix),
			ContactNumber:     fmt.Sprintf("+6391%08d", service.Faker.IntBetween(10000000, 99999999)),
			IsEmailVerified:   true,
			IsContactVerified: true,
			CreatedAt:         time.Now().UTC(),
			UpdatedAt:         time.Now().UTC(),
		}
		if err := UserManager(service).Create(ctx, user); err != nil {
			return err
		}

	}
	return nil
}

func SeedMemberProfiles(ctx context.Context, service *horizon.HorizonService, multiplier int32) error {
	profiles, err := MemberProfileManager(service).List(ctx)
	if err != nil {
		return err
	}
	if len(profiles) > 0 {
		return nil
	}

	organizations, err := OrganizationManager(service).List(ctx)
	if err != nil {
		return err
	}

	users, err := UserManager(service).List(ctx)
	if err != nil {
		return err
	}

	if len(organizations) == 0 || len(users) == 0 {
		return nil
	}

	for _, org := range organizations {

		branches, err := BranchManager(service).Find(ctx, &Branch{
			OrganizationID: org.ID,
		})
		if err != nil {
			continue
		}

		for _, branch := range branches {
			numMembers := int(multiplier) * 1
			numMembers = min(numMembers, len(users))

			for i := 0; i < numMembers; i++ {
				firstName := service.Faker.Person().FirstName()
				middleName := service.Faker.Person().LastName()[:1]
				lastName := service.Faker.Person().LastName()
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
					Description:           service.Faker.Lorem().Paragraph(2),
					Notes:                 service.Faker.Lorem().Paragraph(1),
					ContactNumber:         fmt.Sprintf("+6391%08d", service.Faker.IntBetween(10000000, 99999999)),
					OldReferenceID:        fmt.Sprintf("REF-%04d", i+1),
					Passbook:              passbook,
					Occupation:            []string{"Farmer", "Teacher", "Driver", "Vendor", "Employee", "Business Owner"}[i%6],
					BusinessAddress:       service.Faker.Address().Address(),
					BusinessContactNumber: fmt.Sprintf("+6391%08d", service.Faker.IntBetween(10000000, 99999999)),
					CivilStatus:           []string{"married", "single", "widowed", "divorced"}[i%4],
					IsClosed:              false,
					IsMutualFundMember:    i%2 == 0,
					IsMicroFinanceMember:  i%3 == 0,
				}

				if err := MemberProfileManager(service).Create(ctx, memberProfile); err != nil {
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
					Address:         service.Faker.Address().Address(),
					ProvinceState:   service.Faker.Address().State(),
					City:            service.Faker.Address().City(),
					Barangay:        service.Faker.Address().StreetName(),
					PostalCode:      service.Faker.Address().PostCode(),
					CountryCode:     "PH",
					Landmark:        service.Faker.Lorem().Sentence(2),
				}

				if err := MemberAddressManager(service).Create(ctx, memberAddress); err != nil {
					return err
				}

			}
		}
	}
	return nil
}
