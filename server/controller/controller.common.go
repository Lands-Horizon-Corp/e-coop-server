package v1

import (
	"net/http"

	"github.com/Lands-Horizon-Corp/e-coop-server/services/handlers"
	"github.com/labstack/echo/v4"
)

type GovernmentIDResponse struct {
	Name string `json:"name"`

	HasExpiryDate bool `json:"has_expiry_date"`

	FieldName string `json:"field_name"`
	HasNumber bool   `json:"has_number"`
	Regex     string `json:"regex,omitempty"`
}

func (c *Controller) CommonController() {
	req := c.provider.Service.Request

	req.RegisterRoute(handlers.Route{
		Route:        "/api/v1/government-ids/:country_code",
		Method:       "GET",
		ResponseType: GovernmentIDResponse{},
		Note:         "Retrieves a list of all government IDs available in the system.",
	}, func(ctx echo.Context) error {
		countryCode := ctx.Param("country_code")
		if countryCode == "" {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid country_code"})
		}
		result := []GovernmentIDResponse{}
		switch countryCode {
		case "USD": // United States
			result = []GovernmentIDResponse{
				{
					Name:          "Social Security Number (SSN)",
					HasExpiryDate: false,
					FieldName:     "SSN",
					HasNumber:     true,
					Regex:         `^\d{3}-\d{2}-\d{4}$`,
				},
				{
					Name:          "U.S. Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,9}$`,
				},
				{
					Name:          "U.S. Passport Card",
					HasExpiryDate: true,
					FieldName:     "Passport Card Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,9}$`,
				},
				{
					Name:          "Driver's License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Za-z0-9]{4,20}$`, // varies by state
				},
				{
					Name:          "State Identification Card",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^[A-Za-z0-9]{4,20}$`,
				},
				{
					Name:          "Permanent Resident Card (Green Card)",
					HasExpiryDate: true,
					FieldName:     "USCIS Number",
					HasNumber:     true,
					Regex:         `^([A-Z]{3}\d{10}|[0-9]{9})$`, // old & new format
				},
				{
					Name:          "Employment Authorization Document (EAD)",
					HasExpiryDate: true,
					FieldName:     "USCIS Number",
					HasNumber:     true,
					Regex:         `^([A-Z]{3}\d{10}|[0-9]{9})$`,
				},
				{
					Name:          "U.S. Military ID",
					HasExpiryDate: true,
					FieldName:     "DoD ID Number",
					HasNumber:     true,
					Regex:         `^\d{10}$`,
				},
				{
					Name:          "Trusted Traveler Program ID (Global Entry / TSA PreCheck / NEXUS)",
					HasExpiryDate: true,
					FieldName:     "PASS ID",
					HasNumber:     true,
					Regex:         `^\d{9}$`,
				},
				{
					Name:          "Tribal Identification Card",
					HasExpiryDate: true,
					FieldName:     "Tribal ID Number",
					HasNumber:     true,
					Regex:         `^[A-Za-z0-9]{4,20}$`,
				},
				{
					Name:          "U.S. Citizenship Certificate",
					HasExpiryDate: false,
					FieldName:     "Certificate Number",
					HasNumber:     true,
					Regex:         `^(\d{8}|[A-Z]{1}\d{7})$`,
				},
				{
					Name:          "U.S. Naturalization Certificate",
					HasExpiryDate: false,
					FieldName:     "Certificate Number",
					HasNumber:     true,
					Regex:         `^(\d{8}|[A-Z]{1}\d{7})$`,
				},
			}
		case "EUR": // European Union (Germany as representative)
			result = []GovernmentIDResponse{
				{
					Name:          "National Identity Card (Personalausweis)",
					HasExpiryDate: true,
					FieldName:     "ID Card Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{9}$`,
				},
				{
					Name:          "European Union ID Card",
					HasExpiryDate: true,
					FieldName:     "EU ID Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{8,12}$`,
				},
				{
					Name:          "German Passport (Reisepass)",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[CFGHJKLMNPRTVWXYZ][0-9]{8}$`,
				},
				{
					Name:          "European Union Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{8,10}$`,
				},
				{
					Name:          "Driver's License (Führerschein)",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{5,18}$`,
				},
				{
					Name:          "Residence Permit (Aufenthaltstitel)",
					HasExpiryDate: true,
					FieldName:     "Residence Permit Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{10,12}$`,
				},
				{
					Name:          "Social Insurance Number (Sozialversicherungsnummer)",
					HasExpiryDate: false,
					FieldName:     "Insurance Number",
					HasNumber:     true,
					Regex:         `^[0-9]{12}$`,
				},
				{
					Name:          "Tax Identification Number (Steuer-ID)",
					HasExpiryDate: false,
					FieldName:     "Tax ID",
					HasNumber:     true,
					Regex:         `^[0-9]{11}$`,
				},
				{
					Name:          "Health Insurance Card (EHIC / European Health Insurance Card)",
					HasExpiryDate: true,
					FieldName:     "EHIC Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,20}$`,
				},
				{
					Name:          "European Blue Card (Arbeitskarte)",
					HasExpiryDate: true,
					FieldName:     "Card Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{1}[0-9]{7,10}$`,
				},
			}
		case "JPY": // Japan
			result = []GovernmentIDResponse{
				{
					Name:          "My Number Card (Individual Number Card)",
					HasExpiryDate: true,
					FieldName:     "My Number",
					HasNumber:     true,
					Regex:         `^\d{12}$`, // 12-digit number
				},
				{
					Name:          "Japanese Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}[0-9]{7}$`, // e.g., TR1234567
				},
				{
					Name:          "Residence Card (在留カード)",
					HasExpiryDate: true,
					FieldName:     "Residence Card Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}[0-9]{6}[A-Z0-9]{6}$`, // 12 characters with letters & digits
				},
				{
					Name:          "Special Permanent Resident Certificate (特別永住者証明書)",
					HasExpiryDate: true,
					FieldName:     "Certificate Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}[0-9]{8}$`,
				},
				{
					Name:          "Driver's License (運転免許証)",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^\d{12}$`, // 12 numeric digits
				},
				{
					Name:          "Health Insurance Card (健康保険証)",
					HasExpiryDate: false,
					FieldName:     "Insurance Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9\-]{8,20}$`,
				},
				{
					Name:          "Basic Resident Registration Card (住基カード)",
					HasExpiryDate: true,
					FieldName:     "Card Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{10,12}$`,
				},
				{
					Name:          "Japanese Pension Book (年金手帳)",
					HasExpiryDate: false,
					FieldName:     "Pension Number",
					HasNumber:     true,
					Regex:         `^\d{4}-\d{6}$`,
				},
				{
					Name:          "Employee ID (for Corporate Use)",
					HasExpiryDate: false,
					FieldName:     "Employee Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{4,12}$`,
				},
			}
		case "GBP": // United Kingdom
			result = []GovernmentIDResponse{
				{
					Name:          "UK Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^\d{9}$`, // exactly 9 digits
				},
				{
					Name:          "UK Driving Licence",
					HasExpiryDate: true,
					FieldName:     "Driving Licence Number",
					HasNumber:     true,
					Regex:         `^[A-Z9]{5}\d{6}[A-Z]{2}\d{3}$`, // e.g., SMITH701215J99AA
				},
				{
					Name:          "National Insurance Number (NINO)",
					HasExpiryDate: false,
					FieldName:     "NINO",
					HasNumber:     true,
					Regex:         `^[A-CEGHJ-PR-TW-Z]{2}\d{6}[A-D]$`, // AB123456C
				},
				{
					Name:          "Biometric Residence Permit (BRP)",
					HasExpiryDate: true,
					FieldName:     "BRP Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{7}$`, // e.g., RA1234567
				},
				{
					Name:          "Voter Registration Number",
					HasExpiryDate: false,
					FieldName:     "Elector Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{3}\d{8}$`,
				},
				{
					Name:          "NHS Number",
					HasExpiryDate: false,
					FieldName:     "NHS Number",
					HasNumber:     true,
					Regex:         `^\d{3}\s?\d{3}\s?\d{4}$`, // 10 digits, formatted or plain
				},
				{
					Name:          "Residence Permit (Home Office Reference)",
					HasExpiryDate: true,
					FieldName:     "Home Office Reference",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{7}$`,
				},
				{
					Name:          "Firearms Certificate",
					HasExpiryDate: true,
					FieldName:     "Certificate Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,10}$`,
				},
				{
					Name:          "Seaman's Discharge Book",
					HasExpiryDate: true,
					FieldName:     "Discharge Book Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{1}\d{7}$`,
				},
				{
					Name:          "UK CitizenCard (Photo ID)",
					HasExpiryDate: true,
					FieldName:     "Card Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{8,12}$`,
				},
			}
		case "AUD": // Australia
			result = []GovernmentIDResponse{
				{
					Name:          "Australian Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]\d{7}$`, // Example: N1234567
				},
				{
					Name:          "Driver's Licence",
					HasExpiryDate: true,
					FieldName:     "Licence Number",
					HasNumber:     true,
					Regex:         `^[0-9]{5,10}$`, // Varies by state (strict validation is state-dependent)
				},
				{
					Name:          "Medicare Card",
					HasExpiryDate: true,
					FieldName:     "Medicare Number",
					HasNumber:     true,
					Regex:         `^\d{4}\s?\d{5}\s?\d{1}$`, // 10 digits, may contain spaces
				},
				{
					Name:          "Tax File Number (TFN)",
					HasExpiryDate: false,
					FieldName:     "TFN",
					HasNumber:     true,
					Regex:         `^\d{9}$`, // 9 digits exactly
				},
				{
					Name:          "Australian Business Number (ABN)",
					HasExpiryDate: false,
					FieldName:     "ABN",
					HasNumber:     true,
					Regex:         `^\d{11}$`, // 11 digits
				},
				{
					Name:          "Australian Citizenship Certificate",
					HasExpiryDate: false,
					FieldName:     "Certificate Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{1}[0-9]{7}$`, // Example: C1234567
				},
				{
					Name:          "ImmiCard (Immigration ID Card)",
					HasExpiryDate: true,
					FieldName:     "ImmiCard Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{3}[0-9]{6}$`, // Example: EAA123456
				},
				{
					Name:          "Australian Proof of Age Card",
					HasExpiryDate: true,
					FieldName:     "Card Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Electoral Roll Number",
					HasExpiryDate: false,
					FieldName:     "Elector Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{8,12}$`,
				},
				{
					Name:          "Firearms Licence",
					HasExpiryDate: true,
					FieldName:     "Licence Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,10}$`,
				},
			}
		case "CAD": // Canada
			result = []GovernmentIDResponse{
				{
					Name:          "Canadian Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6}$`, // e.g., AB123456
				},
				{
					Name:          "Driver's License",
					HasExpiryDate: true,
					FieldName:     "Licence Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{5,15}$`, // varies by province (general format)
				},
				{
					Name:          "Social Insurance Number (SIN)",
					HasExpiryDate: false,
					FieldName:     "SIN",
					HasNumber:     true,
					Regex:         `^\d{3}-\d{3}-\d{3}$`, // or plain 9 digits: ^\d{9}$
				},
				{
					Name:          "Permanent Resident Card",
					HasExpiryDate: true,
					FieldName:     "PR Card Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{8}$`, // e.g., RA12345678
				},
				{
					Name:          "Canadian Citizenship Certificate",
					HasExpiryDate: false,
					FieldName:     "Certificate Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{1}\d{7}$`, // e.g., A1234567
				},
				{
					Name:          "Health Card (Provincial)",
					HasExpiryDate: true,
					FieldName:     "Health Card Number",
					HasNumber:     true,
					Regex:         `^\d{4}-\d{3}-\d{3}-\d{2}$`, // Ontario example format
				},
				{
					Name:          "Canadian Armed Forces ID",
					HasExpiryDate: true,
					FieldName:     "Service Number",
					HasNumber:     true,
					Regex:         `^[A-Z]\d{7}$`, // e.g., M1234567
				},
				{
					Name:          "Secure Certificate of Indian Status (SCIS)",
					HasExpiryDate: true,
					FieldName:     "SCIS Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6}$`,
				},
				{
					Name:          "Firearms Possession and Acquisition Licence (PAL)",
					HasExpiryDate: true,
					FieldName:     "PAL Number",
					HasNumber:     true,
					Regex:         `^\d{8}$`, // 8 digits
				},
				{
					Name:          "NEXUS Card",
					HasExpiryDate: true,
					FieldName:     "NEXUS Number",
					HasNumber:     true,
					Regex:         `^[0-9]{9}$`,
				},
			}
		case "CHF": // Switzerland
			result = []GovernmentIDResponse{
				{Name: "Swiss Passport", HasExpiryDate: true, FieldName: "Passport Number", HasNumber: true, Regex: `^[A-Z]\d{7}$`},                           // Example: X1234567
				{Name: "Swiss National ID Card", HasExpiryDate: true, FieldName: "ID Number", HasNumber: true, Regex: `^[A-Z0-9]{9}$`},                        // 9 alphanumeric
				{Name: "Swiss Residence Permit", HasExpiryDate: true, FieldName: "Permit Number", HasNumber: true, Regex: `^[A-Z]{1}\d{9}$`},                  // S123456789
				{Name: "AVS/AHV Social Security Number", HasExpiryDate: false, FieldName: "AHV Number", HasNumber: true, Regex: `^756\.\d{4}\.\d{4}\.\d{2}$`}, // 756.1234.5678.97
				{Name: "Swiss Driving License", HasExpiryDate: true, FieldName: "License Number", HasNumber: true, Regex: `^[A-Z0-9]{10}$`},                   // 10 alphanumeric
			}
		case "CNY": // China
			result = []GovernmentIDResponse{
				{Name: "Resident Identity Card", HasExpiryDate: true, FieldName: "ID Number", HasNumber: true, Regex: `^\d{17}[\dX]$`}, // 18 digits with possible X
				{Name: "Chinese Passport", HasExpiryDate: true, FieldName: "Passport Number", HasNumber: true, Regex: `^[GDE]\d{8}$`},  // Example: G12345678
				{Name: "Household Registration Booklet (Hukou)", HasExpiryDate: false, FieldName: "Hukou Number", HasNumber: true, Regex: `^[A-Z0-9]{5,20}$`},
				{Name: "Driving License", HasExpiryDate: true, FieldName: "License Number", HasNumber: true, Regex: `^[A-Z0-9]{8,16}$`},
				{Name: "Foreigner's Permanent Residence ID", HasExpiryDate: true, FieldName: "Residence ID Number", HasNumber: true, Regex: `^[A-Z]{2}\d{8}$`},
			}

		case "SEK": // Sweden
			result = []GovernmentIDResponse{
				{Name: "Swedish Passport", HasExpiryDate: true, FieldName: "Passport Number", HasNumber: true, Regex: `^[A-Z0-9]{8,10}$`},
				{Name: "Swedish National ID Card", HasExpiryDate: true, FieldName: "ID Number", HasNumber: true, Regex: `^[A-Z0-9]{10,12}$`},
				{Name: "Personnummer (Personal Identity Number)", HasExpiryDate: false, FieldName: "Personnummer", HasNumber: true, Regex: `^\d{6}[-+]?\d{4}$`}, // e.g., 850709-9805
				{Name: "Swedish Driving License", HasExpiryDate: true, FieldName: "License Number", HasNumber: true, Regex: `^[A-Z0-9]{6,15}$`},
				{Name: "Residence Permit Card", HasExpiryDate: true, FieldName: "Permit Number", HasNumber: true, Regex: `^[A-Z]{2}\d{6,8}$`},
			}
		case "NZD": // New Zealand
			result = []GovernmentIDResponse{
				{
					Name:          "New Zealand Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]\d{7}$`, // Example: A1234567
				},
				{
					Name:          "New Zealand Driver's License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[0-9]{5,15}$`, // Varies by region
				},
				{
					Name:          "Kiwi Access / 18+ Card",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Inland Revenue Department (IRD) Number",
					HasExpiryDate: false,
					FieldName:     "IRD Number",
					HasNumber:     true,
					Regex:         `^\d{8,9}$`, // 8-9 digits
				},
				{
					Name:          "Residence Permit",
					HasExpiryDate: true,
					FieldName:     "Permit Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{1}\d{7,9}$`,
				},
			}
		case "PHP": // Philippines
			result = []GovernmentIDResponse{
				{
					Name:          "Philippine Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]\d{8}$`, // Example: P12345678
				},
				{
					Name:          "Philippine Driver's License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{1,2}\d{6,8}$`, // Varies by issuance
				},
				{
					Name:          "Philippine National ID (PhilSys ID)",
					HasExpiryDate: false,
					FieldName:     "PhilSys Number",
					HasNumber:     true,
					Regex:         `^\d{12}$`, // 12-digit PhilSys number
				},
				{
					Name:          "Social Security System (SSS) Number",
					HasExpiryDate: false,
					FieldName:     "SSS Number",
					HasNumber:     true,
					Regex:         `^\d{2}-\d{7}-\d$`, // Example: 34-1234567-8
				},
				{
					Name:          "Government Service Insurance System (GSIS) ID",
					HasExpiryDate: false,
					FieldName:     "GSIS Number",
					HasNumber:     true,
					Regex:         `^\d{9,11}$`, // 9–11 digits
				},
				{
					Name:          "PhilHealth ID",
					HasExpiryDate: false,
					FieldName:     "PhilHealth Number",
					HasNumber:     true,
					Regex:         `^\d{12}$`, // 12 digits
				},
				{
					Name:          "Tax Identification Number (TIN)",
					HasExpiryDate: false,
					FieldName:     "TIN",
					HasNumber:     true,
					Regex:         `^\d{3}-\d{3}-\d{3}-\d$`, // 12 digits with hyphens
				},
				{
					Name:          "Voter's ID / Commission on Elections (COMELEC)",
					HasExpiryDate: true,
					FieldName:     "Voter ID Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,15}$`,
				},
				{
					Name:          "Postal ID",
					HasExpiryDate: true,
					FieldName:     "Postal ID Number",
					HasNumber:     true,
					Regex:         `^\d{9,12}$`,
				},
				{
					Name:          "Barangay Clearance / Certificate",
					HasExpiryDate: true,
					FieldName:     "Certificate Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Senior Citizen ID",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "PWD ID (Persons with Disability)",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Seaman's Book / Seafarer’s ID",
					HasExpiryDate: true,
					FieldName:     "Seaman Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,10}$`,
				},
				{
					Name:          "Unified Multi-Purpose ID (UMID)",
					HasExpiryDate: true,
					FieldName:     "UMID Number",
					HasNumber:     true,
					Regex:         `^\d{12}$`, // 12-digit UMID
				},
				{
					Name:          "PRC License (Professional Regulation Commission)",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2,5}\d{4,6}$`,
				},
				{
					Name:          "Police Clearance",
					HasExpiryDate: true,
					FieldName:     "Clearance Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "School/Student ID",
					HasExpiryDate: true,
					FieldName:     "Student ID Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{4,12}$`,
				},
				{
					Name:          "Company / Employee ID",
					HasExpiryDate: false,
					FieldName:     "Employee ID",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{4,12}$`,
				},
			}
		case "INR": // India
			result = []GovernmentIDResponse{
				{
					Name:          "Aadhaar Card",
					HasExpiryDate: false,
					FieldName:     "Aadhaar Number",
					HasNumber:     true,
					Regex:         `^\d{12}$`, // 12 digits
				},
				{
					Name:          "PAN Card",
					HasExpiryDate: false,
					FieldName:     "PAN Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{5}\d{4}[A-Z]{1}$`, // Example: ABCDE1234F
				},
				{
					Name:          "Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]\d{7}$`, // Example: A1234567
				},
				{
					Name:          "Voter ID (EPIC)",
					HasExpiryDate: false,
					FieldName:     "Voter ID Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{3}\d{7}$`, // Example: ABC1234567
				},
				{
					Name:          "Driving License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{13}$`, // Example: MH1234567890123
				},
				{
					Name:          "Ration Card",
					HasExpiryDate: false,
					FieldName:     "Card Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "PAN-linked Bank Account",
					HasExpiryDate: false,
					FieldName:     "Account Number",
					HasNumber:     true,
					Regex:         `^\d{9,18}$`, // Typical bank account number
				},
			}
		case "KRW": // South Korea
			result = []GovernmentIDResponse{
				{
					Name:          "Resident Registration Number",
					HasExpiryDate: false,
					FieldName:     "RRN",
					HasNumber:     true,
					Regex:         `^\d{6}-\d{7}$`, // Example: 900101-1234567
				},
				{
					Name:          "South Korean Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]\d{8}$`, // Example: M12345678
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[0-9]{12,16}$`,
				},
				{
					Name:          "Alien Registration Card (for foreigners)",
					HasExpiryDate: true,
					FieldName:     "ARC Number",
					HasNumber:     true,
					Regex:         `^\d{9,10}$`,
				},
				{
					Name:          "Health Insurance Card",
					HasExpiryDate: true,
					FieldName:     "Card Number",
					HasNumber:     true,
					Regex:         `^[0-9]{10,12}$`,
				},
			}

		case "THB": // Thailand
			result = []GovernmentIDResponse{
				{
					Name:          "Thai National ID Card",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{13}$`, // 13 digits
				},
				{
					Name:          "Thai Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]\d{8}$`, // Example: A12345678
				},
				{
					Name:          "Thai Driver's License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^\d{12,15}$`,
				},
				{
					Name:          "Thai Social Security Card",
					HasExpiryDate: true,
					FieldName:     "SSO Number",
					HasNumber:     true,
					Regex:         `^\d{10}$`, // 10 digits
				},
				{
					Name:          "Thai Work Permit (for foreigners)",
					HasExpiryDate: true,
					FieldName:     "Permit Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6,8}$`,
				},
			}
		case "SGD": // Singapore
			result = []GovernmentIDResponse{
				{
					Name:          "Singapore NRIC (National Registration Identity Card)",
					HasExpiryDate: true,
					FieldName:     "NRIC Number",
					HasNumber:     true,
					Regex:         `^[STFG]\d{7}[A-Z]$`, // Example: S1234567D
				},
				{
					Name:          "Singapore Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{7}$`, // Example: MA1234567
				},
				{
					Name:          "Singapore FIN (Foreign Identification Number)",
					HasExpiryDate: true,
					FieldName:     "FIN Number",
					HasNumber:     true,
					Regex:         `^[FTG]\d{7}[A-Z]$`, // Example: F1234567A
				},
				{
					Name:          "Singapore Driver's License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[SFTG]\d{7}[A-Z]$`, // 8-character alphanumeric
				},
				{
					Name:          "Employment Pass / Work Permit",
					HasExpiryDate: true,
					FieldName:     "Pass Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{1,2}\d{7,8}$`,
				},
			}
		case "HKD": // Hong Kong
			result = []GovernmentIDResponse{
				{
					Name:          "Hong Kong Identity Card (HKID)",
					HasExpiryDate: true,
					FieldName:     "HKID Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{1,2}\d{6}\([0-9A]\)$`, // Example: A123456(7)
				},
				{
					Name:          "Hong Kong Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]\d{7}$`, // Example: P1234567
				},
				{
					Name:          "Hong Kong Driving License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^\d{6,8}$`,
				},
				{
					Name:          "Hong Kong Permanent Identity Card",
					HasExpiryDate: true,
					FieldName:     "Permanent ID Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{1,2}\d{6}\([0-9A]\)$`,
				},
				{
					Name:          "Hong Kong Student ID",
					HasExpiryDate: true,
					FieldName:     "Student ID Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{5,12}$`,
				},
			}
		case "MYR": // Malaysia
			result = []GovernmentIDResponse{
				{
					Name:          "MyKad (National Identity Card)",
					HasExpiryDate: true,
					FieldName:     "MyKad Number",
					HasNumber:     true,
					Regex:         `^\d{6}-\d{2}-\d{4}$`, // Example: 800101-01-1234
				},
				{
					Name:          "Malaysian Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{1,2}\d{7}$`, // Example: A1234567
				},
				{
					Name:          "Driving License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{1,2}\d{6,8}$`,
				},
				{
					Name:          "Police / Security Clearance ID",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Student / University ID",
					HasExpiryDate: true,
					FieldName:     "Student ID Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{5,12}$`,
				},
			}
		case "IDR": // Indonesia
			result = []GovernmentIDResponse{
				{
					Name:          "KTP (Kartu Tanda Penduduk / National ID Card)",
					HasExpiryDate: true,
					FieldName:     "KTP Number",
					HasNumber:     true,
					Regex:         `^\d{16}$`, // 16 digits
				},
				{
					Name:          "Indonesian Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]\d{8}$`, // Example: A12345678
				},
				{
					Name:          "Driver’s License (SIM)",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{1}\d{11}$`, // Example: B12345678901
				},
				{
					Name:          "NPWP (Tax Identification Number)",
					HasExpiryDate: false,
					FieldName:     "NPWP Number",
					HasNumber:     true,
					Regex:         `^\d{15}$`, // 15 digits
				},
				{
					Name:          "Family Card (KK / Kartu Keluarga)",
					HasExpiryDate: false,
					FieldName:     "KK Number",
					HasNumber:     true,
					Regex:         `^\d{16}$`,
				},
			}
		case "VND": // Vietnam
			result = []GovernmentIDResponse{
				{
					Name:          "Vietnamese Citizen Identity Card (CCCD / CMND)",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{9,12}$`, // 9 or 12 digits depending on issuance year
				},
				{
					Name:          "Vietnamese Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]\d{7,8}$`, // Example: B1234567
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{1}\d{10,12}$`,
				},
				{
					Name:          "Tax Identification Number (TIN)",
					HasExpiryDate: false,
					FieldName:     "TIN Number",
					HasNumber:     true,
					Regex:         `^\d{10,13}$`,
				},
				{
					Name:          "Residence Booklet (Hộ khẩu)",
					HasExpiryDate: false,
					FieldName:     "Residence Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "TWD": // Taiwan
			result = []GovernmentIDResponse{
				{
					Name:          "National Identification Card",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^[A-Z][12]\d{8}$`, // Example: A123456789
				},
				{
					Name:          "Taiwan Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{7}$`, // Example: AB1234567
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{1,2}\d{6,8}$`,
				},
				{
					Name:          "Alien Resident Certificate (ARC / APRC)",
					HasExpiryDate: true,
					FieldName:     "ARC Number",
					HasNumber:     true,
					Regex:         `^[A-Z]\d{9}$`,
				},
				{
					Name:          "Health Insurance Card (NHI)",
					HasExpiryDate: true,
					FieldName:     "Card Number",
					HasNumber:     true,
					Regex:         `^\d{8,10}$`,
				},
			}
		case "BND": // Brunei
			result = []GovernmentIDResponse{
				{
					Name:          "Brunei National Identity Card (NRIC / Kad Pengenalan)",
					HasExpiryDate: true,
					FieldName:     "NRIC Number",
					HasNumber:     true,
					Regex:         `^\d{6}-\d{2}-\d{4}$`, // Example: 990101-01-1234
				},
				{
					Name:          "Brunei Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{1,2}\d{7,8}$`,
				},
				{
					Name:          "Brunei Driving License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Brunei Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
				{
					Name:          "Brunei Residence Permit (for foreigners)",
					HasExpiryDate: true,
					FieldName:     "Permit Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{1,2}\d{6,8}$`,
				},
			}
		case "SAR": // Saudi Arabia
			result = []GovernmentIDResponse{
				{
					Name:          "Saudi National ID (Iqama / Civil ID for citizens)",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{10}$`, // 10 digits
				},
				{
					Name:          "Saudi Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]\d{7}$`, // Example: A1234567
				},
				{
					Name:          "Saudi Driving License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^\d{7,10}$`,
				},
				{
					Name:          "Saudi Iqama (Residence Permit for expatriates)",
					HasExpiryDate: true,
					FieldName:     "Iqama Number",
					HasNumber:     true,
					Regex:         `^\d{10}$`,
				},
				{
					Name:          "Saudi Health Card / Insurance Card",
					HasExpiryDate: true,
					FieldName:     "Card Number",
					HasNumber:     true,
					Regex:         `^\d{8,12}$`,
				},
			}

		case "AED": // United Arab Emirates
			result = []GovernmentIDResponse{
				{
					Name:          "Emirates ID (National ID Card)",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{15}$`, // 15 digits
				},
				{
					Name:          "UAE Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{7}$`, // Example: A1234567
				},
				{
					Name:          "UAE Driving License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{7,12}$`,
				},
				{
					Name:          "UAE Resident Visa / Permit",
					HasExpiryDate: true,
					FieldName:     "Permit Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{8,12}$`,
				},
				{
					Name:          "Health / Insurance Card",
					HasExpiryDate: true,
					FieldName:     "Card Number",
					HasNumber:     true,
					Regex:         `^[0-9]{8,12}$`,
				},
			}
		case "ILS": // Israel
			result = []GovernmentIDResponse{
				{
					Name:          "Israeli National ID (Teudat Zehut)",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{9}$`, // 9 digits
				},
				{
					Name:          "Israeli Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]\d{7}$`, // Example: A1234567
				},
				{
					Name:          "Driver's License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^\d{7,8}$`,
				},
				{
					Name:          "Military ID / Service ID",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{7,9}$`,
				},
				{
					Name:          "Health Insurance Card (Kupot Holim)",
					HasExpiryDate: true,
					FieldName:     "Card Number",
					HasNumber:     true,
					Regex:         `^\d{8,10}$`,
				},
			}
		case "ZAR": // South Africa
			result = []GovernmentIDResponse{
				{
					Name:          "South African ID Card / Smart ID",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{13}$`, // 13 digits
				},
				{
					Name:          "South African Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{7}$`, // Example: AB1234567
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{5,12}$`,
				},
				{
					Name:          "Social Security / Tax Number (SARS)",
					HasExpiryDate: false,
					FieldName:     "Tax Number",
					HasNumber:     true,
					Regex:         `^\d{10,12}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "EGP": // Egypt
			result = []GovernmentIDResponse{
				{
					Name:          "Egyptian National ID Card",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{14}$`, // 14 digits
				},
				{
					Name:          "Egyptian Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]\d{7}$`, // Example: A1234567
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^\d{6,10}$`,
				},
				{
					Name:          "Military / Service ID",
					HasExpiryDate: true,
					FieldName:     "Service Number",
					HasNumber:     true,
					Regex:         `^\d{7,10}$`,
				},
				{
					Name:          "Tax / National Insurance Number",
					HasExpiryDate: false,
					FieldName:     "TIN",
					HasNumber:     true,
					Regex:         `^\d{10,12}$`,
				},
			}
		case "TRY": // Turkey
			result = []GovernmentIDResponse{
				{
					Name:          "Turkish National ID Card (T.C. Kimlik Kartı)",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{11}$`, // 11 digits
				},
				{
					Name:          "Turkish Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]\d{8}$`, // Example: A12345678
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^\d{8,12}$`,
				},
				{
					Name:          "Tax Identification Number",
					HasExpiryDate: false,
					FieldName:     "TIN",
					HasNumber:     true,
					Regex:         `^\d{10}$`,
				},
				{
					Name:          "Residence Permit (for foreigners)",
					HasExpiryDate: true,
					FieldName:     "Permit Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
			}
		case "XOF": // West African CFA Franc
			result = []GovernmentIDResponse{
				{
					Name:          "National Identity Card",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{8,12}$`, // Varies by country
				},
				{
					Name:          "Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{1,2}\d{6,8}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax / Fiscal ID Number",
					HasExpiryDate: false,
					FieldName:     "TIN",
					HasNumber:     true,
					Regex:         `^\d{8,12}$`,
				},
				{
					Name:          "Voter Registration Card",
					HasExpiryDate: true,
					FieldName:     "Card Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
			}
		case "XAF": // Central African CFA Franc
			result = []GovernmentIDResponse{
				{
					Name:          "National Identity Card",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{8,12}$`, // Varies by country
				},
				{
					Name:          "Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{1,2}\d{6,8}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax / Fiscal ID Number",
					HasExpiryDate: false,
					FieldName:     "TIN",
					HasNumber:     true,
					Regex:         `^\d{8,12}$`,
				},
				{
					Name:          "Voter Registration Card",
					HasExpiryDate: true,
					FieldName:     "Card Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
			}
		case "MUR": // Mauritius
			result = []GovernmentIDResponse{
				{
					Name:          "Mauritian National ID Card",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{8,12}$`,
				},
				{
					Name:          "Mauritian Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]\d{7,8}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax / National Insurance Number",
					HasExpiryDate: false,
					FieldName:     "TIN / NIN",
					HasNumber:     true,
					Regex:         `^\d{8,12}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "MVR": // Maldives
			result = []GovernmentIDResponse{
				{
					Name:          "Maldivian National ID Card",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{7,12}$`,
				},
				{
					Name:          "Maldivian Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]\d{7}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax / Social Security Number",
					HasExpiryDate: false,
					FieldName:     "TIN / SSN",
					HasNumber:     true,
					Regex:         `^\d{8,12}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "NOK": // Norway
			result = []GovernmentIDResponse{
				{
					Name:          "Norwegian National ID Number (Fødselsnummer)",
					HasExpiryDate: false,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{11}$`, // 11 digits
				},
				{
					Name:          "Norwegian Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6,7}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Residence Permit (for foreigners)",
					HasExpiryDate: true,
					FieldName:     "Permit Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (TIN)",
					HasExpiryDate: false,
					FieldName:     "TIN",
					HasNumber:     true,
					Regex:         `^\d{9}$`,
				},
			}
		case "DKK": // Denmark
			result = []GovernmentIDResponse{
				{
					Name:          "Danish Personal Identification Number (CPR number)",
					HasExpiryDate: false,
					FieldName:     "CPR Number",
					HasNumber:     true,
					Regex:         `^\d{10}$`, // 10 digits
				},
				{
					Name:          "Danish Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6}$`, // Example: AB123456
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Danish Health Insurance Card",
					HasExpiryDate: true,
					FieldName:     "Card Number",
					HasNumber:     true,
					Regex:         `^\d{10,12}$`,
				},
				{
					Name:          "Residence / Work Permit",
					HasExpiryDate: true,
					FieldName:     "Permit Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
			}
		case "PLN": // Poland
			result = []GovernmentIDResponse{
				{
					Name:          "Polish National ID Card (Dowód Osobisty)",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{3}\d{6}$`, // Example: ABC123456
				},
				{
					Name:          "Polish Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{7}$`, // Example: AB1234567
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "PESEL (National Identification Number)",
					HasExpiryDate: false,
					FieldName:     "PESEL Number",
					HasNumber:     true,
					Regex:         `^\d{11}$`, // 11 digits
				},
				{
					Name:          "Residence Permit / Alien Card",
					HasExpiryDate: true,
					FieldName:     "Permit Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
			}
		case "CZK": // Czech Republic
			result = []GovernmentIDResponse{
				{
					Name:          "Czech National ID Card (Občanský průkaz)",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{8,10}$`, // Usually 8 digits, sometimes 10
				},
				{
					Name:          "Czech Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]\d{7}$`, // Example: A1234567
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Residence / Alien Card",
					HasExpiryDate: true,
					FieldName:     "Permit Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Health Insurance Card",
					HasExpiryDate: true,
					FieldName:     "Card Number",
					HasNumber:     true,
					Regex:         `^\d{9,12}$`,
				},
			}
		case "HUF": // Hungary
			result = []GovernmentIDResponse{
				{
					Name:          "Hungarian Personal ID (Személyi Igazolvány)",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6}$`, // Example: AB123456
				},
				{
					Name:          "Hungarian Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6,7}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (Adószám)",
					HasExpiryDate: false,
					FieldName:     "TIN",
					HasNumber:     true,
					Regex:         `^\d{10}$`,
				},
				{
					Name:          "Residence Permit (for foreigners)",
					HasExpiryDate: true,
					FieldName:     "Permit Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
			}
		case "RUB": // Russia
			result = []GovernmentIDResponse{
				{
					Name:          "Russian Internal Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^\d{2}\s\d{2}\s\d{6}$`, // Example: 12 34 567890
				},
				{
					Name:          "Russian International Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^\d{9}$`, // 9 digits
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^\d{9}$`,
				},
				{
					Name:          "SNILS (Social Security Number)",
					HasExpiryDate: false,
					FieldName:     "SNILS Number",
					HasNumber:     true,
					Regex:         `^\d{3}-\d{3}-\d{3} \d{2}$`, // Example: 123-456-789 00
				},
				{
					Name:          "INN (Taxpayer Identification Number)",
					HasExpiryDate: false,
					FieldName:     "INN",
					HasNumber:     true,
					Regex:         `^\d{10,12}$`,
				},
			}
		case "EUR-HR": // Croatia (Euro)
			result = []GovernmentIDResponse{
				{
					Name:          "Croatian National ID Card (Osobna iskaznica)",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6}$`, // Example: AB123456
				},
				{
					Name:          "Croatian Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6,7}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "OIB (Personal Identification Number / Tax Number)",
					HasExpiryDate: false,
					FieldName:     "OIB",
					HasNumber:     true,
					Regex:         `^\d{11}$`, // 11 digits
				},
				{
					Name:          "Residence Permit (for foreigners)",
					HasExpiryDate: true,
					FieldName:     "Permit Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
			}
		case "BRL": // Brazil
			result = []GovernmentIDResponse{
				{
					Name:          "Brazilian National ID Card (RG)",
					HasExpiryDate: true,
					FieldName:     "RG Number",
					HasNumber:     true,
					Regex:         `^\d{2}\.\d{3}\.\d{3}-\d{1}$`, // Example: 12.345.678-9
				},
				{
					Name:          "Brazilian CPF (Taxpayer Number)",
					HasExpiryDate: false,
					FieldName:     "CPF Number",
					HasNumber:     true,
					Regex:         `^\d{3}\.\d{3}\.\d{3}-\d{2}$`, // Example: 123.456.789-09
				},
				{
					Name:          "Brazilian Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6}$`,
				},
				{
					Name:          "Driver’s License (CNH)",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{9,12}$`,
				},
				{
					Name:          "Voter Registration Card (Título de Eleitor)",
					HasExpiryDate: false,
					FieldName:     "Card Number",
					HasNumber:     true,
					Regex:         `^\d{12}$`,
				},
			}
		case "MXN": // Mexico
			result = []GovernmentIDResponse{
				{
					Name:          "Mexican National ID Card (CURP)",
					HasExpiryDate: false,
					FieldName:     "CURP",
					HasNumber:     true,
					Regex:         `^[A-Z]{4}\d{6}[A-Z]{6}\d{2}$`, // Example: GARC800101HDFABC09
				},
				{
					Name:          "Mexican Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6,7}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (RFC)",
					HasExpiryDate: false,
					FieldName:     "RFC",
					HasNumber:     true,
					Regex:         `^[A-Z]{3,4}\d{6}[A-Z0-9]{3}$`,
				},
				{
					Name:          "Voter Registration Card",
					HasExpiryDate: true,
					FieldName:     "Voter ID Number",
					HasNumber:     true,
					Regex:         `^\d{18}$`,
				},
			}
		case "ARS": // Argentina
			result = []GovernmentIDResponse{
				{
					Name:          "Argentine National ID (DNI)",
					HasExpiryDate: true,
					FieldName:     "DNI Number",
					HasNumber:     true,
					Regex:         `^\d{7,8}$`, // 7-8 digits
				},
				{
					Name:          "Argentine Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]\d{7}$`, // Example: A1234567
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (CUIT / CUIL)",
					HasExpiryDate: false,
					FieldName:     "CUIT / CUIL",
					HasNumber:     true,
					Regex:         `^\d{11}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "CLP": // Chile
			result = []GovernmentIDResponse{
				{
					Name:          "Chilean National ID (RUN / RUT)",
					HasExpiryDate: false,
					FieldName:     "RUT Number",
					HasNumber:     true,
					Regex:         `^\d{7,8}-[\dKk]$`, // Example: 12345678-9 or 12345678-K
				},
				{
					Name:          "Chilean Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (RUT for businesses)",
					HasExpiryDate: false,
					FieldName:     "Tax Number",
					HasNumber:     true,
					Regex:         `^\d{9,12}-[\dKk]$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "PEN": // Peru
			result = []GovernmentIDResponse{
				{
					Name:          "Peruvian National ID (DNI)",
					HasExpiryDate: true,
					FieldName:     "DNI Number",
					HasNumber:     true,
					Regex:         `^\d{8}$`, // 8 digits
				},
				{
					Name:          "Peruvian Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6,7}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (RUC)",
					HasExpiryDate: false,
					FieldName:     "RUC Number",
					HasNumber:     true,
					Regex:         `^\d{11}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "UYU": // Uruguay
			result = []GovernmentIDResponse{
				{
					Name:          "Uruguayan National ID (Cédula de Identidad)",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{7,8}$`, // 7-8 digits
				},
				{
					Name:          "Uruguayan Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6,7}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (RUT)",
					HasExpiryDate: false,
					FieldName:     "RUT Number",
					HasNumber:     true,
					Regex:         `^\d{11}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "DOP": // Dominican Republic
			result = []GovernmentIDResponse{
				{
					Name:          "Dominican National ID (Cédula de Identidad)",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{11}$`, // 11 digits
				},
				{
					Name:          "Dominican Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6,7}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (RNC)",
					HasExpiryDate: false,
					FieldName:     "RNC Number",
					HasNumber:     true,
					Regex:         `^\d{9}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "PYG": // Paraguay
			result = []GovernmentIDResponse{
				{
					Name:          "Paraguayan National ID (Cédula de Identidad)",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{6,8}$`,
				},
				{
					Name:          "Paraguayan Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6,7}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (RUC)",
					HasExpiryDate: false,
					FieldName:     "RUC Number",
					HasNumber:     true,
					Regex:         `^\d{10,12}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "BOB": // Bolivia
			result = []GovernmentIDResponse{
				{
					Name:          "Bolivian National ID (Cédula de Identidad)",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{7,10}$`, // 7-10 digits
				},
				{
					Name:          "Bolivian Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6,7}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (NIT)",
					HasExpiryDate: false,
					FieldName:     "NIT Number",
					HasNumber:     true,
					Regex:         `^\d{11}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "VES": // Venezuela
			result = []GovernmentIDResponse{
				{
					Name:          "Venezuelan National ID (Cédula de Identidad)",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{7,8}$`,
				},
				{
					Name:          "Venezuelan Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6,7}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (RIF)",
					HasExpiryDate: false,
					FieldName:     "RIF Number",
					HasNumber:     true,
					Regex:         `^[JGVEP]-\d{8}-\d$`, // Example: V-12345678-9
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "PKR": // Pakistan
			result = []GovernmentIDResponse{
				{
					Name:          "Pakistani National Identity Card (CNIC)",
					HasExpiryDate: true,
					FieldName:     "CNIC Number",
					HasNumber:     true,
					Regex:         `^\d{5}-\d{7}-\d$`, // Example: 12345-1234567-1
				},
				{
					Name:          "Pakistani Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{1,2}\d{7}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (NTN)",
					HasExpiryDate: false,
					FieldName:     "NTN",
					HasNumber:     true,
					Regex:         `^\d{7,9}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "BDT": // Bangladesh
			result = []GovernmentIDResponse{
				{
					Name:          "Bangladeshi National ID (NID)",
					HasExpiryDate: true,
					FieldName:     "NID Number",
					HasNumber:     true,
					Regex:         `^\d{10,17}$`, // 10-17 digits depending on issuance
				},
				{
					Name:          "Bangladeshi Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{7}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (TIN)",
					HasExpiryDate: false,
					FieldName:     "TIN Number",
					HasNumber:     true,
					Regex:         `^\d{9}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "LKR": // Sri Lanka
			result = []GovernmentIDResponse{
				{
					Name:          "Sri Lankan National ID (NIC)",
					HasExpiryDate: true,
					FieldName:     "NIC Number",
					HasNumber:     true,
					Regex:         `^\d{9}[VvXx]$|^\d{12}$`, // Old: 123456789V, New: 12 digits
				},
				{
					Name:          "Sri Lankan Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (TIN)",
					HasExpiryDate: false,
					FieldName:     "TIN Number",
					HasNumber:     true,
					Regex:         `^\d{9,10}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "NPR": // Nepal
			result = []GovernmentIDResponse{
				{
					Name:          "Nepalese Citizenship Card",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{10,12}$`,
				},
				{
					Name:          "Nepalese Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (PAN / TIN)",
					HasExpiryDate: false,
					FieldName:     "PAN / TIN",
					HasNumber:     true,
					Regex:         `^\d{9,10}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "MMK": // Myanmar
			result = []GovernmentIDResponse{
				{
					Name:          "Myanmar National Registration Card (NRC)",
					HasExpiryDate: true,
					FieldName:     "NRC Number",
					HasNumber:     true,
					Regex:         `^\d{1,2}/[A-Z]{1,3}\(\w+\)\d{6}$`, // Example: 12/LAK(N)123456
				},
				{
					Name:          "Myanmar Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (TIN)",
					HasExpiryDate: false,
					FieldName:     "TIN Number",
					HasNumber:     true,
					Regex:         `^\d{9,12}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "KHR": // Cambodia
			result = []GovernmentIDResponse{
				{
					Name:          "Cambodian National ID (NID / CID)",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{9,12}$`,
				},
				{
					Name:          "Cambodian Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (TIN)",
					HasExpiryDate: false,
					FieldName:     "TIN Number",
					HasNumber:     true,
					Regex:         `^\d{9,12}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "LAK": // Laos
			result = []GovernmentIDResponse{
				{
					Name:          "Laos National ID Card",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{9,12}$`,
				},
				{
					Name:          "Laotian Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (TIN)",
					HasExpiryDate: false,
					FieldName:     "TIN Number",
					HasNumber:     true,
					Regex:         `^\d{9,12}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}

		case "NGN": // Nigeria
			result = []GovernmentIDResponse{
				{
					Name:          "National Identity Number (NIN)",
					HasExpiryDate: false,
					FieldName:     "NIN Number",
					HasNumber:     true,
					Regex:         `^\d{11}$`,
				},
				{
					Name:          "Nigerian Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6,7}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (TIN)",
					HasExpiryDate: false,
					FieldName:     "TIN Number",
					HasNumber:     true,
					Regex:         `^\d{10}$`,
				},
				{
					Name:          "Voter Registration Card",
					HasExpiryDate: false,
					FieldName:     "Voter ID Number",
					HasNumber:     true,
					Regex:         `^\d{10,12}$`,
				},
			}

		case "KES": // Kenya
			result = []GovernmentIDResponse{
				{
					Name:          "Kenyan National ID Card",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{7,8}$`,
				},
				{
					Name:          "Kenyan Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6,7}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Personal Identification Number (PIN / KRA)",
					HasExpiryDate: false,
					FieldName:     "PIN",
					HasNumber:     true,
					Regex:         `^[A-Z]{1}\d{9}[A-Z]{1}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}

		case "GHS": // Ghana
			result = []GovernmentIDResponse{
				{
					Name:          "Ghanaian National ID Card",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{9,12}$`,
				},
				{
					Name:          "Ghanaian Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6,7}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (TIN)",
					HasExpiryDate: false,
					FieldName:     "TIN Number",
					HasNumber:     true,
					Regex:         `^\d{9}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}

		case "MAD": // Morocco
			result = []GovernmentIDResponse{
				{
					Name:          "Moroccan National ID Card (CIN)",
					HasExpiryDate: true,
					FieldName:     "CIN Number",
					HasNumber:     true,
					Regex:         `^\d{8}$`, // 8 digits
				},
				{
					Name:          "Moroccan Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6,7}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (IF / Identifiant Fiscal)",
					HasExpiryDate: false,
					FieldName:     "IF Number",
					HasNumber:     true,
					Regex:         `^\d{9}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}

		case "TND": // Tunisia
			result = []GovernmentIDResponse{
				{
					Name:          "Tunisian National Identity Card (CIN)",
					HasExpiryDate: true,
					FieldName:     "CIN Number",
					HasNumber:     true,
					Regex:         `^\d{8}$`, // 8 digits
				},
				{
					Name:          "Tunisian Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6,7}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (IF / Identifiant Fiscal)",
					HasExpiryDate: false,
					FieldName:     "IF Number",
					HasNumber:     true,
					Regex:         `^\d{9}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}

		case "ETB": // Ethiopia
			result = []GovernmentIDResponse{
				{
					Name:          "Ethiopian National ID Card",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{10,12}$`,
				},
				{
					Name:          "Ethiopian Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6,7}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (TIN)",
					HasExpiryDate: false,
					FieldName:     "TIN Number",
					HasNumber:     true,
					Regex:         `^\d{9,12}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}

		case "DZD": // Algeria
			result = []GovernmentIDResponse{
				{
					Name:          "Algerian National ID Card (CIN)",
					HasExpiryDate: true,
					FieldName:     "CIN Number",
					HasNumber:     true,
					Regex:         `^\d{8}$`, // 8 digits
				},
				{
					Name:          "Algerian Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (NIF)",
					HasExpiryDate: false,
					FieldName:     "NIF Number",
					HasNumber:     true,
					Regex:         `^\d{10}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}

		case "UAH": // Ukraine
			result = []GovernmentIDResponse{
				{
					Name:          "Ukrainian Passport (Internal ID Card)",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{9}$`, // 9 digits
				},
				{
					Name:          "International Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (TIN / INN)",
					HasExpiryDate: false,
					FieldName:     "TIN Number",
					HasNumber:     true,
					Regex:         `^\d{10}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}

		case "RON": // Romania
			result = []GovernmentIDResponse{
				{
					Name:          "Romanian National ID Card (CNP)",
					HasExpiryDate: true,
					FieldName:     "CNP Number",
					HasNumber:     true,
					Regex:         `^\d{13}$`, // 13 digits
				},
				{
					Name:          "Romanian Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (CIF / TIN)",
					HasExpiryDate: false,
					FieldName:     "TIN Number",
					HasNumber:     true,
					Regex:         `^\d{7}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "BGN": // Bulgaria
			result = []GovernmentIDResponse{
				{
					Name:          "Bulgarian Personal Number (EGN)",
					HasExpiryDate: true,
					FieldName:     "EGN Number",
					HasNumber:     true,
					Regex:         `^\d{10}$`, // 10 digits
				},
				{
					Name:          "Bulgarian Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (BG TIN / BULSTAT)",
					HasExpiryDate: false,
					FieldName:     "TIN Number",
					HasNumber:     true,
					Regex:         `^\d{9}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "RSD": // Serbia
			result = []GovernmentIDResponse{
				{
					Name:          "Serbian National ID Card",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{9,13}$`,
				},
				{
					Name:          "Serbian Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6,7}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (PIB)",
					HasExpiryDate: false,
					FieldName:     "TIN Number",
					HasNumber:     true,
					Regex:         `^\d{9}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}

		case "ISK": // Iceland
			result = []GovernmentIDResponse{
				{
					Name:          "Icelandic National ID (Kennitala)",
					HasExpiryDate: false,
					FieldName:     "Kennitala",
					HasNumber:     true,
					Regex:         `^\d{6}-\d{4}$`, // Format: DDMMYY-XXXX
				},
				{
					Name:          "Icelandic Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (TIN)",
					HasExpiryDate: false,
					FieldName:     "TIN Number",
					HasNumber:     true,
					Regex:         `^\d{10}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}

		case "BYN": // Belarus
			result = []GovernmentIDResponse{
				{
					Name:          "Belarusian National ID Card (Internal Passport)",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{9,14}$`,
				},
				{
					Name:          "Belarusian Passport (International)",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6,7}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (INN)",
					HasExpiryDate: false,
					FieldName:     "TIN Number",
					HasNumber:     true,
					Regex:         `^\d{12}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "FJD": // Fiji
			result = []GovernmentIDResponse{
				{
					Name:          "Fijian National ID (Voter ID / NID)",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{7,12}$`,
				},
				{
					Name:          "Fijian Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (TIN)",
					HasExpiryDate: false,
					FieldName:     "TIN Number",
					HasNumber:     true,
					Regex:         `^\d{8,10}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "PGK": // Papua New Guinea
			result = []GovernmentIDResponse{
				{
					Name:          "Papua New Guinea National ID",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{7,12}$`,
				},
				{
					Name:          "Papua New Guinea Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6,7}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (TIN)",
					HasExpiryDate: false,
					FieldName:     "TIN Number",
					HasNumber:     true,
					Regex:         `^\d{8,12}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "JMD": // Jamaica
			result = []GovernmentIDResponse{
				{
					Name:          "Jamaican National ID Card",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{7,12}$`,
				},
				{
					Name:          "Jamaican Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Taxpayer Registration Number (TRN)",
					HasExpiryDate: false,
					FieldName:     "TRN Number",
					HasNumber:     true,
					Regex:         `^\d{9}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "CRC": // Costa Rica
			result = []GovernmentIDResponse{
				{
					Name:          "Costa Rican National ID Card (Cédula de Identidad)",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{9}$`, // 9 digits
				},
				{
					Name:          "Costa Rican Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6,7}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (NIF / Cédula Tributaria)",
					HasExpiryDate: false,
					FieldName:     "TIN Number",
					HasNumber:     true,
					Regex:         `^\d{9}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "GTQ": // Guatemala
			result = []GovernmentIDResponse{
				{
					Name:          "Guatemalan Personal Identification Number (DPI)",
					HasExpiryDate: true,
					FieldName:     "DPI Number",
					HasNumber:     true,
					Regex:         `^\d{13}$`, // 13 digits
				},
				{
					Name:          "Guatemalan Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6,7}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (NIT)",
					HasExpiryDate: false,
					FieldName:     "NIT Number",
					HasNumber:     true,
					Regex:         `^\d{8,9}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "XDR": // Special Drawing Rights (IMF)
			result = []GovernmentIDResponse{
				{
					Name:          "No national ID applicable",
					HasExpiryDate: false,
					FieldName:     "N/A",
					HasNumber:     false,
					Regex:         "",
				},
			}

		case "KWD": // Kuwait
			result = []GovernmentIDResponse{
				{
					Name:          "Kuwaiti Civil ID",
					HasExpiryDate: true,
					FieldName:     "Civil ID Number",
					HasNumber:     true,
					Regex:         `^\d{12}$`, // 12 digits
				},
				{
					Name:          "Kuwaiti Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (TIN)",
					HasExpiryDate: false,
					FieldName:     "TIN Number",
					HasNumber:     true,
					Regex:         `^\d{8,12}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}

		case "QAR": // Qatar
			result = []GovernmentIDResponse{
				{
					Name:          "Qatari National ID",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{11}$`, // 11 digits
				},
				{
					Name:          "Qatari Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (TIN)",
					HasExpiryDate: false,
					FieldName:     "TIN Number",
					HasNumber:     true,
					Regex:         `^\d{8,12}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}

		case "OMR": // Oman
			result = []GovernmentIDResponse{
				{
					Name:          "Omani National ID",
					HasExpiryDate: true,
					FieldName:     "Civil ID Number",
					HasNumber:     true,
					Regex:         `^\d{10,12}$`,
				},
				{
					Name:          "Omani Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (TIN)",
					HasExpiryDate: false,
					FieldName:     "TIN Number",
					HasNumber:     true,
					Regex:         `^\d{8,12}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "BHD": // Bahrain
			result = []GovernmentIDResponse{
				{
					Name:          "Bahraini National ID",
					HasExpiryDate: true,
					FieldName:     "Civil ID Number",
					HasNumber:     true,
					Regex:         `^\d{9}$`, // 9 digits
				},
				{
					Name:          "Bahraini Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (TIN)",
					HasExpiryDate: false,
					FieldName:     "TIN Number",
					HasNumber:     true,
					Regex:         `^\d{8,12}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "JOD": // Jordan
			result = []GovernmentIDResponse{
				{
					Name:          "Jordanian National ID",
					HasExpiryDate: true,
					FieldName:     "Civil ID Number",
					HasNumber:     true,
					Regex:         `^\d{10}$`, // 10 digits
				},
				{
					Name:          "Jordanian Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (TIN)",
					HasExpiryDate: false,
					FieldName:     "TIN Number",
					HasNumber:     true,
					Regex:         `^\d{8,12}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		case "KZT": // Kazakhstan
			result = []GovernmentIDResponse{
				{
					Name:          "Kazakh National ID (Internal Passport)",
					HasExpiryDate: true,
					FieldName:     "ID Number",
					HasNumber:     true,
					Regex:         `^\d{9,12}$`,
				},
				{
					Name:          "Kazakh Passport",
					HasExpiryDate: true,
					FieldName:     "Passport Number",
					HasNumber:     true,
					Regex:         `^[A-Z]{2}\d{6,7}$`,
				},
				{
					Name:          "Driver’s License",
					HasExpiryDate: true,
					FieldName:     "License Number",
					HasNumber:     true,
					Regex:         `^[A-Z0-9]{6,12}$`,
				},
				{
					Name:          "Tax Identification Number (TIN)",
					HasExpiryDate: false,
					FieldName:     "TIN Number",
					HasNumber:     true,
					Regex:         `^\d{12}$`,
				},
				{
					Name:          "Birth Certificate",
					HasExpiryDate: false,
					FieldName:     "Registration Number",
					HasNumber:     true,
					Regex:         `^\d{6,12}$`,
				},
			}
		}
		return ctx.JSON(http.StatusNoContent, result)
	})
}
