import re
import sys
import pyperclip

result = '''
Created: func(data *FinancialStatementAccountsGrouping) []string {
	return []string{
		"financial_statement_accounts_grouping.create",
		fmt.Sprintf("financial_statement_accounts_grouping.create.%s", data.ID),
		fmt.Sprintf("financial_statement_accounts_grouping.create.branch.%s", data.BranchID),
		fmt.Sprintf("financial_statement_accounts_grouping.create.organization.%s", data.OrganizationID),
	}
},
Updated: func(data *FinancialStatementAccountsGrouping) []string {
	return []string{
		"financial_statement_accounts_grouping.update",
		fmt.Sprintf("financial_statement_accounts_grouping.update.%s", data.ID),
		fmt.Sprintf("financial_statement_accounts_grouping.update.branch.%s", data.BranchID),
		fmt.Sprintf("financial_statement_accounts_grouping.update.organization.%s", data.OrganizationID),
	}
},
Deleted: func(data *FinancialStatementAccountsGrouping) []string {
	return []string{
		"financial_statement_accounts_grouping.delete",
		fmt.Sprintf("financial_statement_accounts_grouping.delete.%s", data.ID),
		fmt.Sprintf("financial_statement_accounts_grouping.delete.branch.%s", data.BranchID),
		fmt.Sprintf("financial_statement_accounts_grouping.delete.organization.%s", data.OrganizationID),
	}
},
'''

def to_snake(s: str) -> str:
    s = re.sub(r'([A-Z]+)([A-Z][a-z])', r'\1_\2', s)
    s = re.sub(r'([a-z\d])([A-Z])', r'\1_\2', s)
    return s.lower()

def copy_output_to_clipboard(output: str):
    pyperclip.copy(output)
    print(output)

change = "FinancialStatementAccountsGrouping"   

def changer(to: str):
    global result
    result = result.replace(change, to).replace(to_snake(change), to_snake(to))
    copy_output_to_clipboard(result)



if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python main.py <NewName>")
        sys.exit(1)

    new_name = sys.argv[1]
    changer(new_name)
















