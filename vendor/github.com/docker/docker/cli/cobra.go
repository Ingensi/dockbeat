package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// SetupRootCommand sets default usage, help, and error handling for the
// root command.
func SetupRootCommand(rootCmd *cobra.Command) {
	rootCmd.SetUsageTemplate(usageTemplate)
	rootCmd.SetHelpTemplate(helpTemplate)
	rootCmd.SetFlagErrorFunc(FlagErrorFunc)

	rootCmd.PersistentFlags().BoolP("help", "h", false, "Print usage")
	rootCmd.PersistentFlags().MarkShorthandDeprecated("help", "please use --help")
}

// FlagErrorFunc prints an error messages which matches the format of the
// docker/docker/cli error messages
func FlagErrorFunc(cmd *cobra.Command, err error) error {
	if err == nil {
		return err
	}

	usage := ""
	if cmd.HasSubCommands() {
		usage = "\n\n" + cmd.UsageString()
	}
	return StatusError{
		Status:     fmt.Sprintf("%s\nSee '%s --help'.%s", err, cmd.CommandPath(), usage),
		StatusCode: 125,
	}
}

var usageTemplate = `Usage:	{{if not .HasSubCommands}}{{.UseLine}}{{end}}{{if .HasSubCommands}}{{ .CommandPath}} COMMAND{{end}}

{{ .Short | trim }}{{if gt .Aliases 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{ .Example }}{{end}}{{if .HasFlags}}

Options:
{{.Flags.FlagUsages | trimRightSpace}}{{end}}{{ if .HasAvailableSubCommands}}

Commands:{{range .Commands}}{{if .IsAvailableCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{ if .HasSubCommands }}

Run '{{.CommandPath}} COMMAND --help' for more information on a command.{{end}}
`

var helpTemplate = `
{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`
