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
  ______                ______              ______            __                  __     ____             __                     
 /_  __/______  _____  / ____/___ _   __   / ____/___  ____  / /__________ ______/ /_   / __ \___  ____  / /___  __  _____  _____
  / / / ___/ / / / _ \/ / __/ __ \ | / /  / /   / __ \/ __ \/ __/ ___/ __ / ___/ __/  / / / / _ \/ __ \/ / __ \/ / / / _ \/ ___/
 / / / /  / /_/ /  __/ /_/ / /_/ / |/ /  / /___/ /_/ / / / / /_/ /  / /_/ / /__/ /_   / /_/ /  __/ /_/ / / /_/ / /_/ /  __/ /    
/_/ /_/   \__,_/\___/\____/\____/|___/   \____/\____/_/ /_/\__/_/   \__,_/\___/\__/  /_____/\___/ .___/_/\____/\__, /\___/_/     
                                                                                               /_/            /____/             
%sVersion: %s%s%s
`, yellow, green, version, reset)

	// Print banner with current date and time
	fmt.Println(banner)
}
