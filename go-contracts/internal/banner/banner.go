package banner

import (
	"fmt"
)

var green = "\033[32m"
var reset = "\033[0m"
var yellow = "\033[33m"

// printBanner prints a banner when the application starts
func PrintBanner(version string) {
	banner := fmt.Sprintf(`
  ______                  ______                                            
 / _____)                / _____)           _                     _         
| /  ___  ___     ___   | /      ___  ____ | |_   ____ ____  ____| |_   ___ 
| | (___)/ _ \   (___)  | |     / _ \|  _ \|  _) / ___) _  |/ ___)  _) /___)
| \____/| |_| |         | \____| |_| | | | | |__| |  ( ( | ( (___| |__|___ |
 \_____/ \___/           \______)___/|_| |_|\___)_|   \_||_|\____)\___|___/ 
                                                                          %sVersion: %s%s%s
`, yellow, green, version, reset)

	// Print banner with current date and time
	fmt.Println(banner)
}
