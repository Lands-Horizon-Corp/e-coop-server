package event

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Lands-Horizon-Corp/e-coop-server/server/model/core"
	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

// LoanStatsDataPoint represents statistics for a specific time period
type LoanStatsDataPoint struct {
	Date               string    `json:"date"`                // Date in YYYY-MM-DD format
	Timestamp          time.Time `json:"timestamp"`           // Full timestamp
	LoansReleased      int       `json:"loans_released"`      // New loans released in this period
	LoansFullyPaid     int       `json:"loans_fully_paid"`    // Loans completed in this period
	ActiveLoans        int       `json:"active_loans"`        // Total active loans at this point
	TotalMembers       int       `json:"total_members"`       // Members with loans at this point
	NewMembers         int       `json:"new_members"`         // New members who got loans in this period
	AmountDisbursed    float64   `json:"amount_disbursed"`    // Total amount disbursed in this period
	AmountCollected    float64   `json:"amount_collected"`    // Total payments collected in this period
	PrincipalCollected float64   `json:"principal_collected"` // Principal portion of collections
	InterestCollected  float64   `json:"interest_collected"`  // Interest portion of collections
	FeesCollected      float64   `json:"fees_collected"`      // Fees portion of collections
	TotalArrears       float64   `json:"total_arrears"`       // Total arrears at this point
	TotalOverdue       float64   `json:"total_overdue"`       // Total overdue amount at this point
	OverdueLoans       int       `json:"overdue_loans"`       // Number of overdue loans at this point
	AverageTicketSize  float64   `json:"average_ticket_size"` // Average loan amount in this period
	CollectionRate     float64   `json:"collection_rate"`     // Payment collection rate (%)
	DefaultRate        float64   `json:"default_rate"`        // Default/overdue rate (%)
	PortfolioValue     float64   `json:"portfolio_value"`     // Total outstanding principal
	NetIncome          float64   `json:"net_income"`          // Interest + Fees - Expenses
	ProfitMargin       float64   `json:"profit_margin"`       // (Net Income / Amount Disbursed) * 100
}

// LoanStatsOverTimeResponse represents loan statistics over a time period
type LoanStatsOverTimeResponse struct {
	DataPoints     []LoanStatsDataPoint `json:"data_points"`
	Period         string               `json:"period"` // "daily", "weekly", "monthly", "yearly"
	StartDate      time.Time            `json:"start_date"`
	EndDate        time.Time            `json:"end_date"`
	OrganizationID uuid.UUID            `json:"organization_id"`
	BranchID       uuid.UUID            `json:"branch_id"`

	// Aggregated Summary
	TotalLoansReleased      int     `json:"total_loans_released"`
	TotalLoansFullyPaid     int     `json:"total_loans_fully_paid"`
	TotalAmountDisbursed    float64 `json:"total_amount_disbursed"`
	TotalAmountCollected    float64 `json:"total_amount_collected"`
	TotalPrincipalCollected float64 `json:"total_principal_collected"`
	TotalInterestCollected  float64 `json:"total_interest_collected"`
	TotalFeesCollected      float64 `json:"total_fees_collected"`
	TotalNetIncome          float64 `json:"total_net_income"`
	AverageProfitMargin     float64 `json:"average_profit_margin"`
	AverageCollectionRate   float64 `json:"average_collection_rate"`

	GeneratedAt time.Time `json:"generated_at"`
}

// MemberLoanSummary represents loan summary for a single member
type MemberLoanSummary struct {
	MemberProfileID   uuid.UUID  `json:"member_profile_id"`
	TotalLoans        int        `json:"total_loans"`
	TotalArrears      float64    `json:"total_arrears"`
	TotalPrincipal    float64    `json:"total_principal"`
	TotalPaid         float64    `json:"total_paid"`
	TotalRemaining    float64    `json:"total_remaining"`
	TotalDue          float64    `json:"total_due"`
	ActiveLoans       int        `json:"active_loans"`
	FullyPaidLoans    int        `json:"fully_paid_loans"`
	OverdueLoans      int        `json:"overdue_loans"`
	LastPaymentDate   *time.Time `json:"last_payment_date,omitempty"`
	LastPaymentAmount float64    `json:"last_payment_amount"`
}

// AllMembersLoanSummaryResponse represents loan summaries for all members
type AllMembersLoanSummaryResponse struct {
	MemberSummaries     []MemberLoanSummary `json:"member_summaries"`
	TotalMembers        int                 `json:"total_members"`
	TotalLoans          int                 `json:"total_loans"`
	TotalArrears        float64             `json:"total_arrears"`
	TotalPrincipal      float64             `json:"total_principal"`
	TotalPaid           float64             `json:"total_paid"`
	TotalRemaining      float64             `json:"total_remaining"`
	TotalDue            float64             `json:"total_due"`
	TotalActiveLoans    int                 `json:"total_active_loans"`
	TotalFullyPaidLoans int                 `json:"total_fully_paid_loans"`
	TotalOverdueLoans   int                 `json:"total_overdue_loans"`
	MembersWithLoans    int                 `json:"members_with_loans"`
	MembersWithOverdue  int                 `json:"members_with_overdue"`
	MembersFullyPaid    int                 `json:"members_fully_paid"`
	OrganizationID      uuid.UUID           `json:"organization_id"`
	BranchID            uuid.UUID           `json:"branch_id"`
	GeneratedAt         time.Time           `json:"generated_at"`
}

func (e *Event) AllMembersLoanSummary(
	context context.Context,
	userOrg *core.UserOrganization,
) (*AllMembersLoanSummaryResponse, error) {
	// ===============================================================================================
	// STEP 1: VALIDATE INPUT PARAMETERS
	// ===============================================================================================
	if userOrg == nil {
		return nil, eris.New("user organization is required")
	}
	if userOrg.BranchID == nil {
		return nil, eris.New("branch ID is required")
	}

	// ===============================================================================================
	// STEP 1.5: CHECK CACHE
	// ===============================================================================================
	// Generate cache key based on organization, branch, and current date (daily cache)
	currentDate := userOrg.UserOrgTime().Format("2006-01-02")
	cacheKey := fmt.Sprintf("loan_summary:all_members:%s:%s:%s",
		userOrg.OrganizationID.String(),
		userOrg.BranchID.String(),
		currentDate,
	)

	// Try to get from cache
	if e.provider.Service.Cache != nil {
		cachedData, err := e.provider.Service.Cache.Get(context, cacheKey)
		if err == nil && cachedData != nil {
			var cachedResponse AllMembersLoanSummaryResponse
			if err := json.Unmarshal(cachedData, &cachedResponse); err == nil {
				return &cachedResponse, nil
			}
		}
	}

	// ===============================================================================================
	// STEP 2: FETCH ALL LOAN TRANSACTIONS FOR THE BRANCH
	// ===============================================================================================
	allLoanTransactions, err := e.core.LoanTransactionManager.Find(context, &core.LoanTransaction{
		OrganizationID: userOrg.OrganizationID,
		BranchID:       *userOrg.BranchID,
	})
	if err != nil {
		return nil, eris.Wrap(err, "failed to retrieve loan transactions")
	}

	// ===============================================================================================
	// STEP 3: GROUP LOAN TRANSACTIONS BY MEMBER PROFILE
	// ===============================================================================================
	memberLoanMap := make(map[uuid.UUID][]*core.LoanTransaction)
	for _, loan := range allLoanTransactions {
		if loan.MemberProfileID != nil {
			memberLoanMap[*loan.MemberProfileID] = append(memberLoanMap[*loan.MemberProfileID], loan)
		}
	}

	// ===============================================================================================
	// STEP 4: INITIALIZE AGGREGATION VARIABLES
	// ===============================================================================================
	memberSummaries := []MemberLoanSummary{}
	totalMembers := len(memberLoanMap)
	totalLoans := 0
	totalArrears := 0.0
	totalPrincipal := 0.0
	totalPaid := 0.0
	totalRemaining := 0.0
	totalDue := 0.0
	totalActiveLoans := 0
	totalFullyPaidLoans := 0
	totalOverdueLoans := 0
	membersWithLoans := 0
	membersWithOverdue := 0
	membersFullyPaid := 0

	// ===============================================================================================
	// STEP 5: PROCESS EACH MEMBER'S LOANS
	// ===============================================================================================
	for memberProfileID, loans := range memberLoanMap {
		// -------------------------------------------------------------------------------------------
		// 5.1: Initialize Member-Level Aggregates
		// -------------------------------------------------------------------------------------------
		memberTotalArrears := 0.0
		memberTotalPrincipal := 0.0
		memberTotalPaid := 0.0
		memberTotalRemaining := 0.0
		memberTotalDue := 0.0
		memberActiveLoans := 0
		memberFullyPaidLoans := 0
		memberOverdueLoans := 0
		var memberLastPaymentDate *time.Time
		memberLastPaymentAmount := 0.0

		// -------------------------------------------------------------------------------------------
		// 5.2: Process Each Loan Transaction for This Member
		// -------------------------------------------------------------------------------------------
		for _, loan := range loans {
			// Get detailed loan summary
			loanSummary, err := e.LoanSummary(context, &loan.ID, userOrg)
			if err != nil {
				// Log error but continue processing other loans
				continue
			}

			// Get payment summary for overdue calculations
			loanPaymentSummary, err := e.LoanPaymenSummary(context, &loan.ID, userOrg)
			if err != nil {
				// If payment summary fails, continue with basic summary
				loanPaymentSummary = nil
			}

			// Aggregate member-level metrics
			memberTotalArrears = e.provider.Service.Decimal.Add(memberTotalArrears, loanSummary.Arrears)
			memberTotalPrincipal = e.provider.Service.Decimal.Add(memberTotalPrincipal, loanSummary.TotalPrincipal)
			memberTotalPaid = e.provider.Service.Decimal.Add(memberTotalPaid, loanSummary.TotalPrincipalPaid)
			memberTotalRemaining = e.provider.Service.Decimal.Add(memberTotalRemaining, loanSummary.TotalRemainingPrincipal)

			// Use payment summary for accurate due amount
			if loanPaymentSummary != nil {
				memberTotalDue = e.provider.Service.Decimal.Add(memberTotalDue, loanPaymentSummary.Summary.TotalDueAmount)
			}

			// Track loan status
			if loanSummary.TotalRemainingPrincipal > 0.01 {
				memberActiveLoans++
			} else {
				memberFullyPaidLoans++
			}

			// Track overdue status using payment summary (more accurate)
			if loanPaymentSummary != nil && loanPaymentSummary.Summary.TotalOverduePayments > 0 {
				memberOverdueLoans++
			} else if loanPaymentSummary == nil && loanSummary.Arrears > 0.01 {
				// Fallback to arrears if payment summary unavailable
				memberOverdueLoans++
			}

			// Track latest payment date from payment summary
			if loanPaymentSummary != nil && loanPaymentSummary.Summary.LastPaymentDate != "" {
				parsedDate, err := time.Parse("2006-01-02", loanPaymentSummary.Summary.LastPaymentDate)
				if err == nil {
					if memberLastPaymentDate == nil || parsedDate.After(*memberLastPaymentDate) {
						memberLastPaymentDate = &parsedDate
						memberLastPaymentAmount = loanPaymentSummary.Summary.LastPaymentAmount
					}
				}
			} else if loanSummary.LastPayment != nil {
				// Fallback to loan summary last payment
				if memberLastPaymentDate == nil || loanSummary.LastPayment.After(*memberLastPaymentDate) {
					memberLastPaymentDate = loanSummary.LastPayment
					memberLastPaymentAmount = loanSummary.TotalPrincipalPaid
				}
			}
		}

		// -------------------------------------------------------------------------------------------
		// 5.3: Fetch Member Profile Details
		// -------------------------------------------------------------------------------------------

		// -------------------------------------------------------------------------------------------
		// 5.4: Build Member Summary
		// -------------------------------------------------------------------------------------------
		memberSummary := MemberLoanSummary{
			MemberProfileID:   memberProfileID,
			TotalLoans:        len(loans),
			TotalArrears:      memberTotalArrears,
			TotalPrincipal:    memberTotalPrincipal,
			TotalPaid:         memberTotalPaid,
			TotalRemaining:    memberTotalRemaining,
			TotalDue:          memberTotalDue,
			ActiveLoans:       memberActiveLoans,
			FullyPaidLoans:    memberFullyPaidLoans,
			OverdueLoans:      memberOverdueLoans,
			LastPaymentDate:   memberLastPaymentDate,
			LastPaymentAmount: memberLastPaymentAmount,
		}
		memberSummaries = append(memberSummaries, memberSummary)

		// -------------------------------------------------------------------------------------------
		// 5.5: Aggregate Organization-Level Totals
		// -------------------------------------------------------------------------------------------
		totalLoans += len(loans)
		totalArrears = e.provider.Service.Decimal.Add(totalArrears, memberTotalArrears)
		totalPrincipal = e.provider.Service.Decimal.Add(totalPrincipal, memberTotalPrincipal)
		totalPaid = e.provider.Service.Decimal.Add(totalPaid, memberTotalPaid)
		totalRemaining = e.provider.Service.Decimal.Add(totalRemaining, memberTotalRemaining)
		totalDue = e.provider.Service.Decimal.Add(totalDue, memberTotalDue)
		totalActiveLoans += memberActiveLoans
		totalFullyPaidLoans += memberFullyPaidLoans
		totalOverdueLoans += memberOverdueLoans

		if len(loans) > 0 {
			membersWithLoans++
		}
		if memberOverdueLoans > 0 {
			membersWithOverdue++
		}
		if memberActiveLoans == 0 && len(loans) > 0 {
			membersFullyPaid++
		}
	}

	// ===============================================================================================
	// STEP 6: BUILD RESPONSE
	// ===============================================================================================
	response := &AllMembersLoanSummaryResponse{
		MemberSummaries:     memberSummaries,
		TotalMembers:        totalMembers,
		TotalLoans:          totalLoans,
		TotalArrears:        totalArrears,
		TotalPrincipal:      totalPrincipal,
		TotalPaid:           totalPaid,
		TotalRemaining:      totalRemaining,
		TotalDue:            totalDue,
		TotalActiveLoans:    totalActiveLoans,
		TotalFullyPaidLoans: totalFullyPaidLoans,
		TotalOverdueLoans:   totalOverdueLoans,
		MembersWithLoans:    membersWithLoans,
		MembersWithOverdue:  membersWithOverdue,
		MembersFullyPaid:    membersFullyPaid,
		OrganizationID:      userOrg.OrganizationID,
		BranchID:            *userOrg.BranchID,
		GeneratedAt:         userOrg.UserOrgTime(),
	}

	// ===============================================================================================
	// STEP 7: CACHE THE RESPONSE
	// ===============================================================================================
	if e.provider.Service.Cache != nil {
		responseData, err := json.Marshal(response)
		if err == nil {
			// Cache for 24 hours
			_ = e.provider.Service.Cache.Set(context, cacheKey, responseData, 24*time.Hour)
		}
	}

	return response, nil
}

// LoanStatsOverTime retrieves loan statistics over a specified time period for graphing
func (e *Event) LoanStatsOverTime(
	context context.Context,
	userOrg *core.UserOrganization,
	startDate time.Time,
	endDate time.Time,
	period string, // "daily", "weekly", "monthly", "yearly"
) (*LoanStatsOverTimeResponse, error) {
	// ===============================================================================================
	// STEP 1: VALIDATE INPUT PARAMETERS
	// ===============================================================================================
	if userOrg == nil {
		return nil, eris.New("user organization is required")
	}
	if userOrg.BranchID == nil {
		return nil, eris.New("branch ID is required")
	}
	if startDate.After(endDate) {
		return nil, eris.New("start date must be before end date")
	}

	validPeriods := map[string]bool{"daily": true, "weekly": true, "monthly": true, "yearly": true}
	if !validPeriods[period] {
		return nil, eris.New("period must be one of: daily, weekly, monthly, yearly")
	}

	// ===============================================================================================
	// STEP 1.5: CHECK CACHE
	// ===============================================================================================
	cacheKey := fmt.Sprintf("loan_stats:overtime:%s:%s:%s:%s:%s:%s",
		userOrg.OrganizationID.String(),
		userOrg.BranchID.String(),
		period,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
		userOrg.UserOrgTime().Format("2006-01-02"),
	)

	if e.provider.Service.Cache != nil {
		cachedData, err := e.provider.Service.Cache.Get(context, cacheKey)
		if err == nil && cachedData != nil {
			var cachedResponse LoanStatsOverTimeResponse
			if err := json.Unmarshal(cachedData, &cachedResponse); err == nil {
				return &cachedResponse, nil
			}
		}
	}

	// ===============================================================================================
	// STEP 2: GENERATE TIME BUCKETS
	// ===============================================================================================
	timeBuckets := generateTimeBuckets(startDate, endDate, period)
	dataPoints := make([]LoanStatsDataPoint, len(timeBuckets))

	// ===============================================================================================
	// STEP 3: FETCH ALL LOAN TRANSACTIONS IN DATE RANGE
	// ===============================================================================================
	allLoanTransactions, err := e.core.LoanTransactionManager.Find(context, &core.LoanTransaction{
		OrganizationID: userOrg.OrganizationID,
		BranchID:       *userOrg.BranchID,
	})
	if err != nil {
		return nil, eris.Wrap(err, "failed to retrieve loan transactions")
	}

	// ===============================================================================================
	// STEP 4: FETCH ALL GENERAL LEDGER ENTRIES IN DATE RANGE
	// ===============================================================================================
	allGeneralLedgers, err := e.core.GeneralLedgerManager.Find(context, &core.GeneralLedger{
		OrganizationID: userOrg.OrganizationID,
		BranchID:       *userOrg.BranchID,
	})
	if err != nil {
		return nil, eris.Wrap(err, "failed to retrieve general ledger entries")
	}

	// Track members seen over time
	membersSeen := make(map[uuid.UUID]time.Time)

	// Aggregated totals
	totalLoansReleased := 0
	totalLoansFullyPaid := 0
	totalAmountDisbursed := 0.0
	totalAmountCollected := 0.0
	totalPrincipalCollected := 0.0
	totalInterestCollected := 0.0
	totalFeesCollected := 0.0
	totalNetIncome := 0.0
	sumProfitMargin := 0.0
	sumCollectionRate := 0.0
	validProfitMarginCount := 0
	validCollectionRateCount := 0

	// ===============================================================================================
	// STEP 5: PROCESS EACH TIME BUCKET
	// ===============================================================================================
	for i, bucket := range timeBuckets {
		bucketStart := bucket
		var bucketEnd time.Time

		// Calculate bucket end based on period
		switch period {
		case "daily":
			bucketEnd = bucketStart.AddDate(0, 0, 1)
		case "weekly":
			bucketEnd = bucketStart.AddDate(0, 0, 7)
		case "monthly":
			bucketEnd = bucketStart.AddDate(0, 1, 0)
		case "yearly":
			bucketEnd = bucketStart.AddDate(1, 0, 0)
		}

		// Initialize data point
		dataPoint := LoanStatsDataPoint{
			Date:      bucketStart.Format("2006-01-02"),
			Timestamp: bucketStart,
		}

		loansReleasedInPeriod := 0
		loansFullyPaidInPeriod := 0
		activeLoansAtPoint := 0
		overdueLoansAtPoint := 0
		amountDisbursedInPeriod := 0.0
		amountCollectedInPeriod := 0.0
		principalCollectedInPeriod := 0.0
		interestCollectedInPeriod := 0.0
		feesCollectedInPeriod := 0.0
		totalArrearsAtPoint := 0.0
		totalOverdueAtPoint := 0.0
		portfolioValueAtPoint := 0.0
		newMembersInPeriod := 0
		totalDisbursedForTicketAvg := 0.0

		// -------------------------------------------------------------------------------------------
		// 5.1: Process Loans Released in This Period
		// -------------------------------------------------------------------------------------------
		for _, loan := range allLoanTransactions {
			if loan.ReleasedDate != nil &&
				!loan.ReleasedDate.Before(bucketStart) &&
				loan.ReleasedDate.Before(bucketEnd) {

				loansReleasedInPeriod++
				amountDisbursedInPeriod = e.provider.Service.Decimal.Add(
					amountDisbursedInPeriod,
					loan.Applied1,
				)
				totalDisbursedForTicketAvg = e.provider.Service.Decimal.Add(
					totalDisbursedForTicketAvg,
					loan.Applied1,
				)

				// Track new members
				if loan.MemberProfileID != nil {
					if firstSeen, exists := membersSeen[*loan.MemberProfileID]; !exists || loan.ReleasedDate.Before(firstSeen) {
						membersSeen[*loan.MemberProfileID] = *loan.ReleasedDate
						if !loan.ReleasedDate.Before(bucketStart) && loan.ReleasedDate.Before(bucketEnd) {
							newMembersInPeriod++
						}
					}
				}
			}

			// -------------------------------------------------------------------------------------------
			// 5.2: Check Active/Overdue Status at Bucket End
			// -------------------------------------------------------------------------------------------
			if loan.ReleasedDate != nil && !loan.ReleasedDate.After(bucketEnd) {
				// Get loan summary at this point in time
				loanSummary, err := e.LoanSummary(context, &loan.ID, userOrg)
				if err == nil {
					// Check if loan was active at bucket end
					if loanSummary.TotalRemainingPrincipal > 0.01 {
						activeLoansAtPoint++
						portfolioValueAtPoint = e.provider.Service.Decimal.Add(
							portfolioValueAtPoint,
							loanSummary.TotalRemainingPrincipal,
						)
					} else {
						// Check if loan was completed in this period
						if loanSummary.TotalRemainingPrincipal <= 0.01 && loanSummary.LastPayment != nil {
							if !loanSummary.LastPayment.Before(bucketStart) && loanSummary.LastPayment.Before(bucketEnd) {
								loansFullyPaidInPeriod++
							}
						}
					}

					// Arrears at this point
					totalArrearsAtPoint = e.provider.Service.Decimal.Add(
						totalArrearsAtPoint,
						loanSummary.Arrears,
					)

					// Get payment summary for overdue
					loanPaymentSummary, err := e.LoanPaymenSummary(context, &loan.ID, userOrg)
					if err == nil && loanPaymentSummary.Summary.TotalOverduePayments > 0 {
						overdueLoansAtPoint++
						totalOverdueAtPoint = e.provider.Service.Decimal.Add(
							totalOverdueAtPoint,
							loanPaymentSummary.Summary.TotalDueAmount,
						)
					}
				}
			}
		}

		// -------------------------------------------------------------------------------------------
		// 5.3: Process Payments/Collections in This Period
		// -------------------------------------------------------------------------------------------
		for _, gl := range allGeneralLedgers {
			if gl.Source == core.GeneralLedgerSourcePayment &&
				!gl.EntryDate.Before(bucketStart) &&
				gl.EntryDate.Before(bucketEnd) &&
				gl.AccountID != nil {

				// Get account to determine if this is a loan-related account
				account, err := e.core.AccountManager.GetByID(context, *gl.AccountID)
				if err == nil {
					// Only process credits to loan-related accounts (not cash/bank debits)
					isLoanAccount := false

					switch account.Type {
					case core.AccountTypeLoan:
						principalCollectedInPeriod = e.provider.Service.Decimal.Add(
							principalCollectedInPeriod,
							gl.Credit,
						)
						isLoanAccount = true
					case core.AccountTypeInterest:
						interestCollectedInPeriod = e.provider.Service.Decimal.Add(
							interestCollectedInPeriod,
							gl.Credit,
						)
						isLoanAccount = true
					case core.AccountTypeFines, core.AccountTypeSVFLedger:
						feesCollectedInPeriod = e.provider.Service.Decimal.Add(
							feesCollectedInPeriod,
							gl.Credit,
						)
						isLoanAccount = true
					}

					// Only add to total collected if this is a loan-related account credit
					if isLoanAccount && gl.Credit > 0 {
						amountCollectedInPeriod = e.provider.Service.Decimal.Add(
							amountCollectedInPeriod,
							gl.Credit,
						)
					}
				}
			}
		}

		// -------------------------------------------------------------------------------------------
		// 5.4: Calculate Derived Metrics
		// -------------------------------------------------------------------------------------------
		// Average ticket size
		averageTicketSize := 0.0
		if loansReleasedInPeriod > 0 {
			averageTicketSize = totalDisbursedForTicketAvg / float64(loansReleasedInPeriod)
		}

		// Collection rate
		collectionRate := 0.0
		if amountDisbursedInPeriod > 0 {
			collectionRate = (amountCollectedInPeriod / amountDisbursedInPeriod) * 100
		}

		// Default rate
		defaultRate := 0.0
		if activeLoansAtPoint > 0 {
			defaultRate = (float64(overdueLoansAtPoint) / float64(activeLoansAtPoint)) * 100
		}

		// Net income (simplified: interest + fees)
		netIncome := e.provider.Service.Decimal.Add(interestCollectedInPeriod, feesCollectedInPeriod)

		// Profit margin
		profitMargin := 0.0
		if amountDisbursedInPeriod > 0 {
			profitMargin = (netIncome / amountDisbursedInPeriod) * 100
		}

		// Members with loans at this point
		membersAtPoint := 0
		for _, seenDate := range membersSeen {
			if !seenDate.After(bucketEnd) {
				membersAtPoint++
			}
		}

		// -------------------------------------------------------------------------------------------
		// 5.5: Build Data Point
		// -------------------------------------------------------------------------------------------
		dataPoint.LoansReleased = loansReleasedInPeriod
		dataPoint.LoansFullyPaid = loansFullyPaidInPeriod
		dataPoint.ActiveLoans = activeLoansAtPoint
		dataPoint.TotalMembers = membersAtPoint
		dataPoint.NewMembers = newMembersInPeriod
		dataPoint.AmountDisbursed = amountDisbursedInPeriod
		dataPoint.AmountCollected = amountCollectedInPeriod
		dataPoint.PrincipalCollected = principalCollectedInPeriod
		dataPoint.InterestCollected = interestCollectedInPeriod
		dataPoint.FeesCollected = feesCollectedInPeriod
		dataPoint.TotalArrears = totalArrearsAtPoint
		dataPoint.TotalOverdue = totalOverdueAtPoint
		dataPoint.OverdueLoans = overdueLoansAtPoint
		dataPoint.AverageTicketSize = averageTicketSize
		dataPoint.CollectionRate = collectionRate
		dataPoint.DefaultRate = defaultRate
		dataPoint.PortfolioValue = portfolioValueAtPoint
		dataPoint.NetIncome = netIncome
		dataPoint.ProfitMargin = profitMargin

		dataPoints[i] = dataPoint

		// -------------------------------------------------------------------------------------------
		// 5.6: Aggregate Totals
		// -------------------------------------------------------------------------------------------
		totalLoansReleased += loansReleasedInPeriod
		totalLoansFullyPaid += loansFullyPaidInPeriod
		totalAmountDisbursed = e.provider.Service.Decimal.Add(totalAmountDisbursed, amountDisbursedInPeriod)
		totalAmountCollected = e.provider.Service.Decimal.Add(totalAmountCollected, amountCollectedInPeriod)
		totalPrincipalCollected = e.provider.Service.Decimal.Add(totalPrincipalCollected, principalCollectedInPeriod)
		totalInterestCollected = e.provider.Service.Decimal.Add(totalInterestCollected, interestCollectedInPeriod)
		totalFeesCollected = e.provider.Service.Decimal.Add(totalFeesCollected, feesCollectedInPeriod)
		totalNetIncome = e.provider.Service.Decimal.Add(totalNetIncome, netIncome)

		if profitMargin > 0 {
			sumProfitMargin += profitMargin
			validProfitMarginCount++
		}
		if collectionRate > 0 {
			sumCollectionRate += collectionRate
			validCollectionRateCount++
		}
	}

	// ===============================================================================================
	// STEP 6: CALCULATE AVERAGES
	// ===============================================================================================
	averageProfitMargin := 0.0
	if validProfitMarginCount > 0 {
		averageProfitMargin = sumProfitMargin / float64(validProfitMarginCount)
	}

	averageCollectionRate := 0.0
	if validCollectionRateCount > 0 {
		averageCollectionRate = sumCollectionRate / float64(validCollectionRateCount)
	}

	// ===============================================================================================
	// STEP 7: BUILD RESPONSE
	// ===============================================================================================
	response := &LoanStatsOverTimeResponse{
		DataPoints:              dataPoints,
		Period:                  period,
		StartDate:               startDate,
		EndDate:                 endDate,
		OrganizationID:          userOrg.OrganizationID,
		BranchID:                *userOrg.BranchID,
		TotalLoansReleased:      totalLoansReleased,
		TotalLoansFullyPaid:     totalLoansFullyPaid,
		TotalAmountDisbursed:    totalAmountDisbursed,
		TotalAmountCollected:    totalAmountCollected,
		TotalPrincipalCollected: totalPrincipalCollected,
		TotalInterestCollected:  totalInterestCollected,
		TotalFeesCollected:      totalFeesCollected,
		TotalNetIncome:          totalNetIncome,
		AverageProfitMargin:     averageProfitMargin,
		AverageCollectionRate:   averageCollectionRate,
		GeneratedAt:             userOrg.UserOrgTime(),
	}

	// ===============================================================================================
	// STEP 8: CACHE THE RESPONSE
	// ===============================================================================================
	if e.provider.Service.Cache != nil {
		responseData, err := json.Marshal(response)
		if err == nil {
			// Cache for 24 hours
			_ = e.provider.Service.Cache.Set(context, cacheKey, responseData, 24*time.Hour)
		}
	}

	return response, nil
}

// generateTimeBuckets creates time buckets for the specified period
func generateTimeBuckets(start, end time.Time, period string) []time.Time {
	buckets := []time.Time{}
	current := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())

	for current.Before(end) || current.Equal(end) {
		buckets = append(buckets, current)

		switch period {
		case "daily":
			current = current.AddDate(0, 0, 1)
		case "weekly":
			current = current.AddDate(0, 0, 7)
		case "monthly":
			current = current.AddDate(0, 1, 0)
		case "yearly":
			current = current.AddDate(1, 0, 0)
		}
	}

	return buckets
}
