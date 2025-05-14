
def main():
    txt = """
memberGenderHistoryG := service.Group("/member-gender-history")
	{
		memberGenderHistoryG.GET("", c.MemberGenderHistoryList)
		memberGenderHistoryG.GET("/:member_gender_history_id", c.MemberGenderHistoryGetByID)
		memberGenderHistoryG.DELETE("/:member_gender_history_id", c.MemberGenderHistoryDelete)
		memberGenderHistoryG.GET("/branch/:branch_id", c.MemberGenderHistoryListByBranch)
		memberGenderHistoryG.GET("/organization/:organization_id", c.MemberGenderHistoryListByOrganization)
		memberGenderHistoryG.GET("/organization/:organization_id/branch/:branch_id", c.MemberGenderHistoryListByOrganizationBranch)
	}
"""
    replacements = [
        ("member_classification", "member_gender"),
        ("member-classificationr", "member-gender"),
        ("memberClassification", "memberGender"),
        ("MemberClassification", "MemberGender"),
        ("MemberClassification", "MemberGender"),
    ]




    for (to_change, from_change) in replacements:
        txt = txt.replace(from_change, to_change)

    print(txt)


if __name__ == "__main__":
    main()
