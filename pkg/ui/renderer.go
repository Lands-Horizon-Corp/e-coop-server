package ui

type Section struct {
	Title string
	Rows  []Row
}

func RenderSection(t Theme, s Section) string {
	lines := []string{}
	if s.Title != "" {
		lines = append(lines, RenderTitle(t, s.Title))
	}
	for _, r := range s.Rows {
		lines = append(lines, RenderRow(t, r))
	}
	return RenderBox(t, lines...)
}
