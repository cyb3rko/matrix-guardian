package util

import "strings"

const commandPrefix = "!gd"

func IsGuardianCommand(command string) bool {
	return command == commandPrefix || strings.HasPrefix(strings.TrimSpace(command), commandPrefix+" ")
}

func ParseCommands(input string) (string, []string) {
	input = strings.TrimPrefix(strings.TrimSpace(input), commandPrefix)
	commands := strings.Split(strings.TrimSpace(input), " ")
	if len(commands) > 1 {
		return strings.ToLower(commands[0]), commands[1:]
	} else if len(commands) == 1 {
		return strings.ToLower(commands[0]), nil
	} else {
		return "", nil
	}
}
