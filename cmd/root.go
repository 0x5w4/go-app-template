package cmd

import (
	"fmt"
	"goapptemp/cmd/app"
	"goapptemp/config"
	"goapptemp/pkg/logger"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "goapptemp",
	Short: "goapptemp service",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of goapptemp service",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println("goapptemp Service v0.1")
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the goapptemp service",
	Run: func(cmd *cobra.Command, _ []string) {
		configFile, err := cmd.Flags().GetString("config")
		if err != nil {
			fmt.Println("Failed to get config flag:", err)
			os.Exit(1)
		}

		debug, err := cmd.Flags().GetBool("debug")
		if err != nil {
			fmt.Println("Failed to get debug flag:", err)
			os.Exit(1)
		}

		env, err := cmd.Flags().GetString("env")
		if err != nil {
			fmt.Println("Failed to get env flag:", err)
			os.Exit(1)
		}

		logger := logger.NewZerologLogger(debug)

		config, err := config.LoadConfig(configFile)
		if err != nil {
			fmt.Println("Failed to load config:", err)
			os.Exit(1)
		}

		config.App.Environment = env
		config.App.Debug = debug

		app, err := app.NewApp(config, logger)
		if err != nil {
			fmt.Println("Failed to create app:", err)
			os.Exit(1)
		}

		if err := app.Run(); err != nil {
			fmt.Println("Failed to run app:", err)
			os.Exit(1)
		}
	},
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run database migrations",
	Run: func(cmd *cobra.Command, _ []string) {
		configFile, err := cmd.Flags().GetString("config")
		if err != nil {
			fmt.Println("Failed to get config flag:", err)
			os.Exit(1)
		}

		reset, err := cmd.Flags().GetBool("reset")
		if err != nil {
			fmt.Println("Failed to get reset flag:", err)
			os.Exit(1)
		}

		logger := logger.NewZerologLogger(true)

		config, err := config.LoadConfig(configFile)
		if err != nil {
			fmt.Println("Failed to load config:", err)
			os.Exit(1)
		}

		app, err := app.NewApp(config, logger)
		if err != nil {
			fmt.Println("Failed to create app:", err)
			os.Exit(1)
		}

		if err := app.Migrate(reset); err != nil {
			fmt.Println("Failed to migrate database:", err)
			os.Exit(1)
		}
	},
}

func runCmdPreRunE(cmd *cobra.Command, _ []string) error {
	env, err := cmd.Flags().GetString("env")
	if err != nil {
		return fmt.Errorf("failed to get env flag: %w", err)
	}

	validEnvs := []string{"local", "development", "staging", "sandbox", "production"}
	if !slices.Contains(validEnvs, env) {
		return fmt.Errorf("invalid environment %s. valid environments are: %v", env, strings.Join(validEnvs, ", "))
	}

	return nil
}

func init() {
	runCmd.Flags().StringP("config", "c", ".env", "Specify the config file (optional)")

	if err := viper.BindPFlag("config", runCmd.Flags().Lookup("config")); err != nil {
		fmt.Println("Failed to bind config flag:", err)
		os.Exit(1)
	}

	runCmd.Flags().BoolP("debug", "d", false, "Is debug (optional)")

	if err := viper.BindPFlag("debug", runCmd.Flags().Lookup("debug")); err != nil {
		fmt.Println("Failed to bind debug flag:", err)
		os.Exit(1)
	}

	runCmd.Flags().StringP("env", "e", "development", "Specify the environment (required)")

	if err := viper.BindPFlag("env", runCmd.Flags().Lookup("env")); err != nil {
		fmt.Println("Failed to bind env flag:", err)
		os.Exit(1)
	}

	migrateCmd.Flags().StringP("config", "c", ".env", "Specify the config file (optional)")

	if err := viper.BindPFlag("config", migrateCmd.Flags().Lookup("config")); err != nil {
		fmt.Println("Failed to bind config flag:", err)
		os.Exit(1)
	}

	migrateCmd.Flags().BoolP("reset", "r", false, "Reset the database (optional)")

	if err := viper.BindPFlag("reset", migrateCmd.Flags().Lookup("reset")); err != nil {
		fmt.Println("Failed to bind reset flag:", err)
		os.Exit(1)
	}

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(migrateCmd)

	runCmd.PreRunE = runCmdPreRunE
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Failed to execute command:", err)
		os.Exit(1)
	}
}
