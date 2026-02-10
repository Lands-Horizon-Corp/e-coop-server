package ui

import (
	"log"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func Separator(title ...string) {
	line := "━"
	left := "≿"
	right := "≾"
	center := "༺❀༻"
	repeat := 35
	full := strings.Repeat(line, repeat)
	sep := left + full + center + full + right
	if len(title) > 0 && title[0] != "" {
		t := title[0]
		sep = left + full[:15] + " " + t + " " + full[:15] + center + full[:15] + right
	}
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true)
	log.Println(style.Render(sep))
}

func Logo() {
	art := `
           ..............                            
        .,,,,,,,,,,,,,,,,,,,                             
    ,,,,,,,,,,,,,,,,,,,,,,,,,,                          
  ,,,,,,,,,,,,,,  .,,,,,,,,,,,,,                        
,,,,,,,,,,           ,,,,,,,,,,,                     
,,,,,,,          .,,,,,,,,,,,                          
@@,,,,,,          ,,,,,,,,,,,,                             
@@@,,,,.@@      ,,,,,,,,,,,,                                
@,,,,,,,@@    ,,,,,,,,,,,                                   
  ,,,,@@@       ,,,,,,                                      
    @@@@@@@                                          
    @@@@@@@@@@           @@@@@@@@                          
      @@@@@@@@@@@@@@  @@@@@@@@@@@@                          
        @@@@@@@@@@@@@@@@@@@@@@@@@@                          
            @@@@@@@@@@@@@@@@@@@@                             
                  @@@@@@@@
	`
	blueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true)
	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("28"))
	lines := strings.Split(art, "\n")
	for _, line := range lines {
		var styled strings.Builder
		for _, r := range line {
			switch r {
			case '@':
				styled.WriteString(blueStyle.Render(string(r)))
			case ',', '.':
				styled.WriteString(greenStyle.Render(string(r)))
			default:
				styled.WriteRune(r)
			}
		}
		log.Println(styled.String())
	}
}
