
def main():
    txt = """
memberGroupG := service.Group("/member-group")
	{
		memberGroupG.GET("", c.MemberGroupList)
		memberGroupG.GET("/:member_group_id", c.MemberGroupGetByID)
		memberGroupG.POST("", c.MemberGroupCreate)
		memberGroupG.PUT("/:member_group_id", c.MemberGroupUpdate)
		memberGroupG.DELETE("/:member_group_id", c.MemberGroupDelete)
		memberGroupG.GET("/branch/:branch_id", c.MemberGroupListByBranch)
		memberGroupG.GET("/organization/:organization_id", c.MemberGroupListByOrganization)
		memberGroupG.GET("/organization/:organization_id/branch/:branch_id", c.MemberGroupListByOrganizationBranch)
	}


"""
    replacements = [
        ("member_group", "member_verification"),
        ("member-group", "member-verification"),
        ("memberGroup", "memberVerification"),
        ("MemberGroup", "MemberVerification"),
    ]
    for (from_change, to_change) in replacements:
        txt = txt.replace(from_change, to_change)

    print(txt)


if __name__ == "__main__":
    main()
