package horizon

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/Delta456/box-cli-maker/v2"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
	"go.uber.org/fx"
)

const (
	Red    = "\033[31m"
	Yellow = "\033[33m"
	Green  = "\033[32m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	Reset  = "\033[0m"
)

// NewTerminal initializes the terminal for the Horizon application
func NewTerminal(lc fx.Lifecycle, request *HorizonRequest, config *HorizonConfig) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if !config.CanDebug() {
				return eris.New("cannot debug")
			}
			time.Sleep(3 * time.Second)
			fmt.Println("≿━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━༺❀༻━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━≾")
			printASCIIArt()
			printConfigBoxes(config)
			requestCMD(request)
			fmt.Println("🟢 Horizon App is starting...")
			fmt.Println("≿━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━༺❀༻━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━≾")
			return nil
		},
	})
}

// requestCMD prints out the grouped routes for the Horizon application
func requestCMD(request *HorizonRequest) {
	routesList := request.service.Routes()

	// Group routes by Controller -> Method -> []*Route
	grouped := make(map[string]map[string][]*echo.Route)

	for _, rt := range routesList {
		controller := extractController(rt.Name)
		if _, ok := grouped[controller]; !ok {
			grouped[controller] = make(map[string][]*echo.Route)
		}
		grouped[controller][rt.Method] = append(grouped[controller][rt.Method], rt)
	}

	// Sort controller names
	controllers := make([]string, 0, len(grouped))
	for c := range grouped {
		controllers = append(controllers, c)
	}
	sort.Strings(controllers)

	// Print grouped routes
	for _, controller := range controllers {
		fmt.Printf("\n%s========== %s ==========%s\n", Cyan, controller, Reset)

		// Sort methods: prioritize GET, POST, PUT, DELETE
		methods := make([]string, 0, len(grouped[controller]))
		for method := range grouped[controller] {
			methods = append(methods, method)
		}
		sort.Slice(methods, func(i, j int) bool {
			order := map[string]int{"GET": 1, "POST": 2, "PUT": 3, "DELETE": 4}
			return order[methods[i]] < order[methods[j]]
		})

		// Print routes by method
		for _, method := range methods {
			routes := grouped[controller][method]

			// Sort routes within each method group by Path
			sort.Slice(routes, func(i, j int) bool {
				return routes[i].Path < routes[j].Path
			})

			// Print each route with color-coding based on HTTP method
			for _, rt := range routes {
				switch rt.Method {
				case "GET":
					fmt.Printf("\t%s▶ %s %s \t%s- %s%s\n", Green, rt.Method, rt.Path, Reset, rt.Name, Reset)
				case "POST":
					fmt.Printf("\t%s▶ %s %s \t%s- %s%s\n", Blue, rt.Method, rt.Path, Reset, rt.Name, Reset)
				case "PUT":
					fmt.Printf("\t%s▶ %s %s \t%s- %s%s\n", Yellow, rt.Method, rt.Path, Reset, rt.Name, Reset)
				case "DELETE":
					fmt.Printf("\t%s▶ %s %s \t%s- %s%s\n", Red, rt.Method, rt.Path, Reset, rt.Name, Reset)
				default:
					fmt.Printf("\t▶ %s %s \t- %s\n", rt.Method, rt.Path, rt.Name)
				}
			}
		}
	}
}

// printConfigBoxes displays configuration information for the Horizon app
func printConfigBoxes(config *HorizonConfig) {
	box := box.Box{
		TopRight:    "*",
		TopLeft:     "*",
		BottomRight: "*",
		BottomLeft:  "*",
		Horizontal:  "-", Vertical: "┃",
		Config: box.Config{
			Px:       5,
			Py:       0,
			Type:     "",
			TitlePos: "Inside", // ensure "Left", "Center", or "Right"
			Color:    "Cyan",
		},
	}

	// App Info
	appConfig := fmt.Sprintf(
		"Environment     : %s\n"+
			"App Name        : %s\n"+
			"Client URL      : %s\n"+
			"App Port        : http://localhost:%d\n"+
			"Metrics Port    : http://localhost:%d",
		config.AppEnvironment,
		config.AppName,
		config.AppClientURL,
		config.AppPort,
		config.AppMetricsPort,
	)

	// PostgreSQL DSN and Admin info
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		config.PostgresUser,
		config.PostgresPassword,
		config.PostgresHost,
		config.PostgresPort,
		config.PostgresDB,
	)

	pgadminURL := fmt.Sprintf("http://%s:%d", config.PgAdminHost, config.PgAdminPort)

	postgresInfo := fmt.Sprintf(
		"DSN               : %s\n"+
			"Postgres User     : %s\n"+
			"Postgres DB       : %s\n"+
			"Postgres Host     : %s:%d\n\n"+
			"PgAdmin Email     : %s\n"+
			"PgAdmin Password  : %s\n"+
			"PgAdmin URL       : %s or http://localhost:%d",
		dsn,
		config.PostgresUser,
		config.PostgresDB,
		config.PostgresHost,
		config.PostgresPort,
		config.PgAdminDefaultEmail,
		config.PgAdminDefaultPassword,
		pgadminURL,
		config.PgAdminPort,
	)

	// Redis DSN and Info
	redisDSN := fmt.Sprintf("redis://%s:%s@%s:%d", config.RedisUsername, config.RedisPassword, config.RedisHost, config.RedisPort)

	redisInfo := fmt.Sprintf(
		"Redis DSN         : %s\n"+
			"Redis Host        : %s\n"+
			"Redis Port        : %d\n"+
			"Redis Username    : %s\n"+
			"Redis Password    : %s\n"+
			"Redis Insight URL : http://%s:%d or http://localhost:%d",
		redisDSN,
		config.RedisHost,
		config.RedisPort,
		config.RedisUsername,
		config.RedisPassword,
		config.RedisInsightHost,
		config.RedisInsightPort,
		config.RedisInsightPort,
	)

	// SMTP DSN and Info
	smtpDSN := fmt.Sprintf("smtp://%s:%s@%s:%d", config.SMTPUsername, config.SMTPPassword, config.SMTPHost, config.SMTPPort)

	smtpInfo := fmt.Sprintf(
		"SMTP DSN          : %s\n"+
			"SMTP Host         : %s\n"+
			"SMTP Port         : %d\n"+
			"SMTP Username     : %s\n"+
			"SMTP Password     : %s\n"+
			"SMTP From Address : %s\n"+
			"MailPit UI URL    : http://%s:%d or http://localhost:%d",
		smtpDSN,
		config.SMTPHost,
		config.SMTPPort,
		config.SMTPUsername,
		config.SMTPPassword,
		config.SMTPFrom,
		config.MailPitUIHost,
		config.MailPitUIPort,
		config.MailPitUIPort,
	)

	// Corrected Storage DSN format
	storageDSN := fmt.Sprintf("storage://%s:%s@%s:%d/%s",
		config.StorageAccessKey, config.StorageSecretKey,
		config.StorageHost, config.StorageApiPort,
		config.StorageBucket)
	storageInfo := fmt.Sprintf(
		"Storage DSN        : %s\n"+
			"Storage Driver     : %s\n"+
			"Storage Access Key : %s\n"+
			"Storage Secret Key : %s\n"+
			"Storage Endpoint   : %s\n"+
			"Storage Region     : %s\n"+
			"Storage Bucket     : %s\n"+
			"Storage API Port   : %d\n"+
			"Storage Console Port: http://localhost:%d",
		storageDSN,
		config.StorageDriver,
		config.StorageAccessKey,
		config.StorageSecretKey,
		config.StorageHost,
		config.StorageRegion,
		config.StorageBucket,
		config.StorageApiPort,
		config.StorageConsolePort,
	)

	// NATS DSN and Info
	natsDSN := fmt.Sprintf("nats://%s:%d", config.NATSHost, config.NATSClientPort)

	natsInfo := fmt.Sprintf(
		"NATS DSN          : %s\n"+
			"NATS Host         : %s\n"+
			"NATS Client Port  : %d\n"+
			"NATS Monitor Port : http://localhost:%d\n"+
			"NATS WebSocket    : ws://localhost:%d",
		natsDSN,
		config.NATSHost,
		config.NATSClientPort,
		config.NATSMonitorPort,
		config.NATSClientWSPort,
	)

	// Print the config boxes
	box.Print("💻 APP CONFIG", appConfig)
	box.Print("💾 DATABASE CONFIG", postgresInfo)
	box.Print("💿 REDIS CONFIG", redisInfo)
	box.Print("✉️ SMTP CONFIG", smtpInfo)
	box.Print("📦 STORAGE CONFIG", storageInfo)
	box.Print("💬 NATS CONFIG", natsInfo)
}

// runCmd executes a command in the terminal
func runCmd(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// ClearTerminal clears the terminal screen
func ClearTerminal() {
	switch runtime.GOOS {
	case "darwin", "linux":
		runCmd("clear")
	case "windows":
		runCmd("cmd", "/c", "cls")
	default:
		runCmd("clear")
	}
}

// printASCIIArt prints ASCII art in the terminal
func printASCIIArt() {
	asciiArt := `
	           ..............                            
            .,,,,,,,,,,,,,,,,,,,                             
        ,,,,,,,,,,,,,,,,,,,,,,,,,,                          
      ,,,,,,,,,,,,,,  .,,,,,,,,,,,,,                        
    ,,,,,,,,,,           ,,,,,,,,,,,                     
    ,,,,,,,          .,,,,,,,,,,,                          
  @@,,,,,,          ,,,,,,,,,,,,                             
@@@,,,,.@@      .,,,,,,,,,,,                                
@,,,,,,,@@    ,,,,,,,,,,,                                   
  ,,,,@@@       ,,,,,,                                      
    @@@@@@@                                          
    @@@@@@@@@@           @@@@@@@@                          
      @@@@@@@@@@@@@@  @@@@@@@@@@@@                          
        @@@@@@@@@@@@@@@@@@@@@@@@@@                          
            @@@@@@@@@@@@@@@@@@@@                             
                  @@@@@@@@
	`

	lines := strings.Split(asciiArt, "\n")

	for _, line := range lines {
		coloredLine := ""
		for _, char := range line {
			switch char {
			case '@':
				coloredLine += Blue + "@" + Reset
			case ',', '.':
				coloredLine += Green + string(char) + Reset
			default:
				coloredLine += string(char)
			}
		}
		fmt.Println(coloredLine)
	}
}
