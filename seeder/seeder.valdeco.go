package seeder

import (
	"context"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/helpers"
	"github.com/Lands-Horizon-Corp/e-coop-server/horizon"
	"github.com/Lands-Horizon-Corp/e-coop-server/src/types"
)

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
			{
				MediaPath:   "seeder/images/valdeco/financial-literacy-2025.jpg",
				Name:        "ＢＵＫＡＳ ＮＡ ＇ＴＯ， ＭＧＡ ＫＡＳＡＰＩ！ Financial Literacy Seminar",
				Description: "💡 Financial Literacy is the first step to financial freedom! Ready to take control of your money and build a brighter future? 💸 Join us on November 21, 2025, at 1:00 PM and learn how smart money management can unlock doors to stability, growth, and lasting success. 🌱💰\n📍 Venue: 2nd Floor Valdeco Bldg, Greenleaf Market, Malinta, Valenzuela City.\nWhether you’re just starting or want to sharpen your skills, this event is for YOU! Don’t miss your chance to build a strong financial foundation.\n📱 For inquiries, contact: 📞 0925-511-5774 & look for Ms. Rochelle C. Planas or Ms. Kimberly T. Lalo.\n📲 Stay connected, Kasapi! Click LIKE & FOLLOW para lagi kang una sa balita at benepisyo! 💡\n#Valdeco #GrowWithUs #FinancialLiteracy",
			},
			{
				MediaPath:   "seeder/images/valdeco/basic-photography-videography-2025.jpg",
				Name:        "Basic Photography & Videography Seminar",
				Description: "📸✨ Capture. Create. Inspire. ✨🎬\n\nBilang bahagi ng Cooperative Month at VALDECO 45th Anniversary, inaanyayahan ang lahat — mga kasapi at non-member — sa isang makulay at makabuluhang Basic Photography and Videography Seminar! 🎥📷\n\n🎙 Resource Speakers:\nMr. Carlzon Niño Lumbang\nMr. Jesmil Flores Dela Cruz\n\n📅 Date: October 29, 2025 (Wednesday)\n🕐 Time: 1:00 PM\n📍 Venue: 2nd Floor, Tangke St., VALDECO Greenleaf Market, Malinta, Valenzuela City\n\n⚠️ Limited slots only — maximum of 100 participants\n📲 Register Here Now: https://docs.google.com/.../viewform\n\n🎯 Matutong magkwento gamit ang larawan at video — dahil bawat kuha ay may kwentong dapat ibahagi. 💚✨\n\n📱 For inquiries, contact: 📞 0925-511-5774 & look for Ms. Rochelle C. Planas or Ms. Kimberly T. Lalo.\n📲 Stay connected, Kasapi! Click LIKE & FOLLOW para lagi kang una sa balita at benepisyo! 💡\n#VALDECO #GrowWithUs #CooperativeMonth2025 #VALDECO45thAnniversary #BasicPhotography&Videography",
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
				Contact:                "+639255115772",
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
				Contact:    "+639222341493",
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
