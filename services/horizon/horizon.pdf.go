package horizon

// type PDFGenerator struct{}

// type Unit string

// const (
// 	UnitMM Unit = "mm"
// 	UnitIN Unit = "in" // inches
// 	UnitPX Unit = "px" // pixels (needs DPI)
// 	UnitPT Unit = "pt" // points = 1/72 inch
// 	UnitCM Unit = "cm"
// )

// type PDFSize struct {
// 	Name        string
// 	DisplayName string
// 	Description string

// 	// One and only one of these will be > 0 depending on original unit
// 	WidthMM  float64
// 	HeightMM float64

// 	WidthIN  float64
// 	HeightIN float64

// 	WidthPX  float64
// 	HeightPX float64

// 	WidthPT  float64
// 	HeightPT float64

// 	WidthCM  float64
// 	HeightCM float64

// 	SourceUnit Unit // Which field above is the original source
// }

// // ————————————————————————————————————————
// // Auto-conversion methods (cached internally)
// // ————————————————————————————————————————

// func (s *PDFSize) Width(unit Unit, dpi ...float64) float64 {
// 	switch unit {
// 	case UnitMM:
// 		return s.toMM().WidthMM
// 	case UnitIN:
// 		return s.toMM().WidthMM / 25.4
// 	case UnitCM:
// 		return s.toMM().WidthMM / 10.0
// 	case UnitPT:
// 		return s.toMM().WidthMM / 25.4 * 72.0
// 	case UnitPX:
// 		d := 300.0
// 		if len(dpi) > 0 {
// 			d = dpi[0]
// 		}
// 		return s.toMM().WidthMM / 25.4 * d
// 	}
// 	return 0
// }

// func (s *PDFSize) Height(unit Unit, dpi ...float64) float64 {
// 	switch unit {
// 	case UnitMM:
// 		return s.toMM().HeightMM
// 	case UnitIN:
// 		return s.toMM().HeightMM / 25.4
// 	// ≈8.27 × 11.69
// 	case UnitCM:
// 		return s.toMM().HeightMM / 10.0
// 	case UnitPT:
// 		return s.toMM().HeightMM / 25.4 * 72.0 // ≈595.28 × 841.89 pt
// 	case UnitPX:
// 		d := 300.0
// 		if len(dpi) > 0 {
// 			d = dpi[0]
// 		}
// 		return s.toMM().HeightMM / 25.4 * d
// 	}
// 	return 0
// }

// // Always returns a copy with MM filled WidthMM/HeightMM
// func (s *PDFSize) toMM() PDFSize {
// 	if s.WidthMM > 0 {
// 		return *s
// 	}

// 	switch s.SourceUnit {
// 	case UnitIN:
// 		return PDFSize{WidthMM: s.WidthIN * 25.4, HeightMM: s.HeightIN * 25.4}
// 	case UnitCM:
// 		return PDFSize{WidthMM: s.WidthCM * 10, HeightMM: s.HeightCM * 10}
// 	case UnitPT:
// 		return PDFSize{WidthMM: s.WidthPT / 72.0 * 25.4, HeightMM: s.HeightPT / 72.0 * 25.4}
// 	case UnitPX:
// 		// Default fallback: 300 DPI if not specified elsewhere
// 		dpi := 300.0
// 		return PDFSize{WidthMM: s.WidthPX / dpi * 25.4, HeightMM: s.HeightPX / dpi * 25.4}
// 	}
// 	return *s
// }

// // Convenience shortcuts
// func (s *PDFSize) MM() (w, h float64)     { return s.Width(UnitMM), s.Height(UnitMM) }
// func (s *PDFSize) Inches() (w, h float64) { return s.Width(UnitIN), s.Height(UnitIN) }
// func (s *PDFSize) Points() (w, h float64) { return s.Width(UnitPT), s.Height(UnitPT) }
// func (s *PDFSize) PX(dpi float64) (w, h float64) {
// 	return s.Width(UnitPX, dpi), s.Height(UnitPX, dpi)
// }

// var AllPDFSizes = []PDFSize{
// 	// ================================================================
// 	// 1. ISO A-Series – Millimeters
// 	// ================================================================
// 	{Name: "A0_mm", DisplayName: "A0", Description: "ISO A0 Poster", WidthMM: 841, HeightMM: 1189, SourceUnit: UnitMM},
// 	{Name: "A1_mm", DisplayName: "A1", Description: "ISO A1 Poster", WidthMM: 594, HeightMM: 841, SourceUnit: UnitMM},
// 	{Name: "A2_mm", DisplayName: "A2", Description: "ISO A2 Poster", WidthMM: 420, HeightMM: 594, SourceUnit: UnitMM},
// 	{Name: "A3_mm", DisplayName: "A3", Description: "ISO A3 (Common for drawings)", WidthMM: 297, HeightMM: 420, SourceUnit: UnitMM},
// 	{Name: "A4_mm", DisplayName: "A4", Description: "A4 – Standard office paper", WidthMM: 210, HeightMM: 297, SourceUnit: UnitMM},
// 	{Name: "A5_mm", DisplayName: "A5", Description: "A5 – Half A4, notebooks", WidthMM: 148, HeightMM: 210, SourceUnit: UnitMM},
// 	{Name: "A6_mm", DisplayName: "A6", Description: "A6 – Postcards, flyers", WidthMM: 105, HeightMM: 148, SourceUnit: UnitMM},
// 	{Name: "A7_mm", DisplayName: "A7", Description: "A7 – Small notes", WidthMM: 74, HeightMM: 105, SourceUnit: UnitMM},
// 	{Name: "A8_mm", DisplayName: "A8", Description: "A8 – Tiny cards", WidthMM: 52, HeightMM: 74, SourceUnit: UnitMM},

// 	// ================================================================
// 	// 2. ISO B & C Series + Envelopes
// 	// ================================================================
// 	{Name: "B5_mm", DisplayName: "B5", Description: "ISO B5 Book/Notebook", WidthMM: 176, HeightMM: 250, SourceUnit: UnitMM},
// 	{Name: "DL_mm", DisplayName: "DL Envelope", Description: "DL Envelope (1/3 A4)", WidthMM: 110, HeightMM: 220, SourceUnit: UnitMM},
// 	{Name: "C5_mm", DisplayName: "C5 Envelope", Description: "C5 Envelope", WidthMM: 162, HeightMM: 229, SourceUnit: UnitMM},
// 	{Name: "C4_mm", DisplayName: "C4 Envelope", Description: "C4 Envelope (fits A4)", WidthMM: 229, HeightMM: 324, SourceUnit: UnitMM},

// 	// ================================================================
// 	// 3. North American Paper Sizes
// 	// ================================================================
// 	{Name: "Letter_mm", DisplayName: "US Letter", Description: "8.5 × 11 in", WidthMM: 215.9, HeightMM: 279.4, SourceUnit: UnitMM},
// 	{Name: "Legal_mm", DisplayName: "US Legal", Description: "8.5 × 14 in", WidthMM: 215.9, HeightMM: 355.6, SourceUnit: UnitMM},
// 	{Name: "Tabloid_mm", DisplayName: "Tabloid / Ledger", Description: "11 × 17 in", WidthMM: 279.4, HeightMM: 431.8, SourceUnit: UnitMM},
// 	{Name: "HalfLetter_mm", DisplayName: "Half Letter", Description: "5.5 × 8.5 in", WidthMM: 139.7, HeightMM: 215.9, SourceUnit: UnitMM},

// 	// ================================================================
// 	// 4. Receipts & Thermal Printers
// 	// ================================================================
// 	{Name: "Receipt_80mm", DisplayName: "Receipt 80mm", Description: "Standard thermal receipt", WidthMM: 80, HeightMM: 200, SourceUnit: UnitMM},
// 	{Name: "Receipt_80mm_Long", DisplayName: "Receipt 80mm Long", Description: "Long receipt/invoice", WidthMM: 80, HeightMM: 600, SourceUnit: UnitMM},
// 	{Name: "Receipt_58mm", DisplayName: "Receipt 58mm", Description: "Small mobile printer", WidthMM: 58, HeightMM: 200, SourceUnit: UnitMM},
// 	{Name: "Receipt_80mm_203dpi", DisplayName: "Receipt 80mm @203dpi", Description: "Native thermal (203 dpi)", WidthPX: 640, HeightPX: 1600, SourceUnit: UnitPX},
// 	{Name: "Receipt_80mm_300dpi", DisplayName: "Receipt 80mm @300dpi", Description: "High-res thermal", WidthPX: 945, HeightPX: 2362, SourceUnit: UnitPX},
// 	{Name: "Receipt_58mm_203dpi", DisplayName: "Receipt 58mm @203dpi", Description: "Small printer native", WidthPX: 464, HeightPX: 1200, SourceUnit: UnitPX},

// 	// ================================================================
// 	// 5. Print-ready Invoices & Reports (pixels)
// 	// ================================================================
// 	{Name: "A4_300dpi", DisplayName: "A4 @300dpi", Description: "Print-ready A4", WidthPX: 2480, HeightPX: 3508, SourceUnit: UnitPX},
// 	{Name: "A4_600dpi", DisplayName: "A4 @600dpi", Description: "High-quality A4", WidthPX: 4961, HeightPX: 7016, SourceUnit: UnitPX},
// 	{Name: "Letter_300dpi", DisplayName: "Letter @300dpi", Description: "Print-ready US Letter", WidthPX: 2550, HeightPX: 3300, SourceUnit: UnitPX},
// 	{Name: "Legal_300dpi", DisplayName: "Legal @300dpi", Description: "Print-ready US Legal", WidthPX: 2550, HeightPX: 4200, SourceUnit: UnitPX},

// 	// ================================================================
// 	// 6. Shipping Labels
// 	// ================================================================
// 	{Name: "Shipping_4x6in_mm", DisplayName: "4×6″ Shipping Label", Description: "Zebra/Rollo/DYMO", WidthMM: 101.6, HeightMM: 152.4, SourceUnit: UnitMM},
// 	{Name: "Shipping_4x6in_300dpi", DisplayName: "4×6″ @300dpi", Description: "Most common shipping label", WidthPX: 1200, HeightPX: 1800, SourceUnit: UnitPX},
// 	{Name: "Shipping_4x6in_203dpi", DisplayName: "4×6″ @203dpi", Description: "Zebra thermal native", WidthPX: 812, HeightPX: 1218, SourceUnit: UnitPX},

// 	// ================================================================
// 	// 7. Business Cards
// 	// ================================================================
// 	{Name: "BizCard_US_mm", DisplayName: "Business Card US", Description: "3.5 × 2 in", WidthMM: 88.9, HeightMM: 50.8, SourceUnit: UnitMM},
// 	{Name: "BizCard_US_300dpi", DisplayName: "Business Card US @300dpi", Description: "Print-ready", WidthPX: 1050, HeightPX: 600, SourceUnit: UnitPX},
// 	{Name: "BizCard_EU_mm", DisplayName: "Business Card EU", Description: "85 × 55 mm", WidthMM: 85, HeightMM: 55, SourceUnit: UnitMM},
// 	{Name: "BizCard_EU_300dpi", DisplayName: "Business Card EU @300dpi", Description: "Print-ready", WidthPX: 1004, HeightPX: 650, SourceUnit: UnitPX},
// 	{Name: "BizCard_Square", DisplayName: "Square Business Card", Description: "Modern square", WidthMM: 65, HeightMM: 65, SourceUnit: UnitMM},

// 	// ================================================================
// 	// 8. Posters & Banners
// 	// ================================================================
// 	{Name: "Poster_24x36in", DisplayName: "24×36″ Poster", Description: "Standard large poster", WidthMM: 609.6, HeightMM: 914.4, SourceUnit: UnitMM},
// 	{Name: "Poster_27x40in", DisplayName: "27×40″ Movie Poster", Description: "Cinema one-sheet", WidthMM: 685.8, HeightMM: 1016, SourceUnit: UnitMM},
// 	{Name: "Banner_850x2000", DisplayName: "Rollup 85×200cm", Description: "Trade show banner", WidthMM: 850, HeightMM: 2000, SourceUnit: UnitMM},

// 	// ================================================================
// 	// 9. Books & Photos
// 	// ================================================================
// 	{Name: "Trade_Paperback_6x9", DisplayName: "6×9 Trade Paperback", Description: "Most popular novel size", WidthMM: 152.4, HeightMM: 228.6, SourceUnit: UnitMM},
// 	{Name: "Photo_4x6in", DisplayName: "4×6 Photo", Description: "Standard photo print", WidthMM: 101.6, HeightMM: 152.4, SourceUnit: UnitMM},
// 	{Name: "Passport_US", DisplayName: "US Passport Photo", Description: "2×2 inch", WidthMM: 50.8, HeightMM: 50.8, SourceUnit: UnitMM},

// 	// ================================================================
// 	// 10. Screen Resolutions
// 	// ================================================================
// 	{Name: "FullHD", DisplayName: "Full HD 1080p", Description: "1920×1080", WidthPX: 1920, HeightPX: 1080, SourceUnit: UnitPX},
// 	{Name: "4K_UHD", DisplayName: "4K UHD", Description: "3840×2160", WidthPX: 3840, HeightPX: 2160, SourceUnit: UnitPX},
// 	{Name: "iPhone15ProMax", DisplayName: "iPhone 15 Pro Max", Description: "Latest iPhone", WidthPX: 1290, HeightPX: 2796, SourceUnit: UnitPX},

// 	// ================================================================
// 	// 11. PDF Native Sizes (Points)
// 	// ================================================================
// 	{Name: "A4_pt", DisplayName: "A4 (Points)", Description: "PDF canvas in points", WidthPT: 595.28, HeightPT: 841.89, SourceUnit: UnitPT},
// 	{Name: "Letter_pt", DisplayName: "US Letter (Points)", Description: "PDF US Letter", WidthPT: 612, HeightPT: 792, SourceUnit: UnitPT},
// 	{Name: "Legal_pt", DisplayName: "US Legal (Points)", Description: "PDF US Legal", WidthPT: 612, HeightPT: 1008, SourceUnit: UnitPT},
// }

// func NewPDFGenerator() *PDFGenerator {
// 	return &PDFGenerator{}
// }

// // func (p *PDFGenerator) Generate(data interface{}) ([]byte, error) {
// // 	ret
// // }

// func (p *PDFGenerator) Name(name string) string {
// 	return name
// }
