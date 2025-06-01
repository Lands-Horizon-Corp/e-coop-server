import re
import pyperclip

result = '''

func (m *Model) MemberGenderCurrentBranch(context context.Context, orgId uuid.UUID, branchId uuid.UUID) ([]*MemberGender, error) {
	return m.MemberGenderManager.Find(context, &MemberGender{
		OrganizationID: orgId,
		BranchID:       branchId,
	})
}

'''

def to_snake(s: str) -> str:
    s = re.sub(r'([A-Z]+)([A-Z][a-z])', r'\1_\2', s)
    s = re.sub(r'([a-z\d])([A-Z])', r'\1_\2', s)
    return s.lower()

def copy_output_to_clipboard(output: str):
    try:
        pyperclip.copy(output)
        print("✅ Copied to clipboard:")
    except pyperclip.PyperclipException as e:
        print("⚠️ Failed to copy to clipboard:", e)
        print("💡 Make sure xclip or xsel is installed and disk has space.")
    finally:
        print("----- Output -----")
        print(output)

change = "MemberGender"

def changer(to: str):
    global result
    result = result.replace(change, to).replace(to_snake(change), to_snake(to))
    copy_output_to_clipboard(result)

if __name__ == "__main__":
    try:
        new_name = pyperclip.paste().strip()
        if not new_name:
            raise ValueError("Clipboard is empty.")
        changer(new_name)
    except Exception as e:
        print(f"❌ Error: {e}")
