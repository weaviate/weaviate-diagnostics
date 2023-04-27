package diagnostics

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	pprof_wrapper "github.com/weaviate/weaviate-diagnostics/diagnostics/pprof"
)

var rootCmd = &cobra.Command{
	Use:   "weaviate-diagnostics",
	Short: "Weaviate Diagnostics",
	Long:  `A tool to help diagnose issues with Weaviate`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("running the root command, see help or -h for available commands\n")
	},
}

var diagnosticsCmd = &cobra.Command{
	Use:   "diagnostics",
	Short: "Run Weaviate Diagnostics",
	Long:  `A tool to help diagnose issues with Weaviate`,
	Run: func(cmd *cobra.Command, args []string) {
		GenerateReport()
	},
}

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Generate a CPU profile",
	Long:  `A wrapper around golang's pprof tool`,
	Run: func(cmd *cobra.Command, args []string) {
		os.Args = []string{"pprof", "-png", "-output", globalConfig.ProfileOutputFile, globalConfig.ProfileUrl}
		pprof_wrapper.GenerateProfile()
	},
}

func initCommand() {
	diagnosticsCmd.PersistentFlags().StringVarP(&globalConfig.OutputFile,
		"output", "o", "weaviate-report.html", "File to write the report to")
	// todo make these configurable
	diagnosticsCmd.PersistentFlags().StringVarP(&globalConfig.Url,
		"url", "u", "http://localhost:8080", "URL of the Weaviate instance")

	diagnosticsCmd.PersistentFlags().StringVarP(&globalConfig.MetricsUrl,
		"metricsUrl", "m", "http://localhost:2112/metrics", "full URL plus path of the Weaviate metrics endpoint")

	diagnosticsCmd.PersistentFlags().StringVarP(&globalConfig.ProfileUrl,
		"profileUrl", "p", "http://localhost:6060/debug/pprof/profile?seconds=5", "URL of the Weaviate pprof endpoint")

	diagnosticsCmd.PersistentFlags().StringVarP(&globalConfig.ApiKey,
		"apiKey", "a", "", "API key authentication")

	diagnosticsCmd.PersistentFlags().StringVarP(&globalConfig.User,
		"user", "n", "", "Username for OIDC authentication")

	diagnosticsCmd.PersistentFlags().StringVarP(&globalConfig.Pass,
		"pass", "w", "", "Password for OIDC authentication (defaults to prompt)")

	profileCmd.PersistentFlags().StringVarP(&globalConfig.ProfileUrl,
		"profileUrl", "p", "http://localhost:6060/debug/pprof/profile?seconds=5", "URL of the Weaviate pprof endpoint")

	profileCmd.PersistentFlags().StringVarP(&globalConfig.ProfileOutputFile,
		"output", "o", "http://localhost:8080", "Where to write the profile to")

	rootCmd.AddCommand(diagnosticsCmd)
	rootCmd.AddCommand(profileCmd)
}

func Execute() {
	initCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
