package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
	"github.com/go-faker/faker/v4"
)

func SeedFakeOrganization(ctx context.Context, service *horizon.HorizonService) error {
	domain := faker.DomainName()
	firstName := faker.FirstName()
	lastName := faker.LastName()
	orgName := fmt.Sprintf("%s %s Cooperative", faker.Word(), faker.Word())
	config := types.OrganizationSeedConfig{
		AdminEmail:           fmt.Sprintf("admin@%s", domain),
		AdminPassword:        "admin123",
		AdminBirthdate:       time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC),
		AdminUsername:        faker.Username(),
		AdminFullName:        fmt.Sprintf("%s %s", firstName, lastName),
		AdminFirstName:       firstName,
		AdminMiddleName:      helpers.Ptr(faker.FirstName()),
		AdminLastName:        lastName,
		AdminContactNumber:   faker.Phonenumber(),
		AdminLogoPath:        "seeder/images/default-admin.png",
		OrgName:              orgName,
		OrgAddress:           helpers.Ptr(faker.GetRealAddress().Address),
		OrgEmail:             helpers.Ptr(fmt.Sprintf("contact@%s", domain)),
		OrgContactNumber:     helpers.Ptr(faker.Phonenumber()),
		OrgDescription:       helpers.Ptr(faker.Paragraph()),
		OrgColor:             helpers.Ptr("#" + faker.UUIDDigit()[:6]),
		OrgTerms:             helpers.Ptr("Fake Terms: " + faker.Sentence()),
		OrgPrivacy:           helpers.Ptr("Fake Privacy: " + faker.Sentence()),
		OrgCookie:            helpers.Ptr("Fake Cookie Policy"),
		OrgRefund:            helpers.Ptr("Fake Refund Policy"),
		OrgUserAgreement:     helpers.Ptr("Fake User Agreement"),
		OrgIsPrivate:         false,
		OrgLogoPath:          "seeder/images/default-logo.png",
		OrgProfilePath:       "seeder/images/default-profile.png",
		OrgInstagram:         helpers.Ptr("https://instagram.com/" + faker.Username()),
		OrgFacebook:          helpers.Ptr("https://facebook.com/" + faker.Username()),
		OrgYoutube:           helpers.Ptr("https://youtube.com/@" + faker.Username()),
		OrgPersonalWebsite:   helpers.Ptr("https://" + domain),
		BranchesRandom:       helpers.Ptr(5),
		SeminarsRandom:       helpers.Ptr(3),
		CurrencyAlpha2:       "PH",
		SubscriptionDays:     30,
		InvitationMaxUse:     500,
		InvitationExpiration: 30 * 24 * time.Hour,
	}

	return SeedOrganization(ctx, service, config)
}
