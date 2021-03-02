package cmd

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var cfgFile string
var debug bool
var info = true

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cal-calc",
	Short: "Calculate billable usage based on Google Calendar events",
	Long: `This utility calculates billable usage based on your Google Calendar events.

It's best used when organizing your timesheets within Google Calendar.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		calculate(
			viper.GetBool("debug"),
			viper.GetFloat64("targetUtilization"),
		)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().Bool("debug", false, "enable debug messaging")
	rootCmd.PersistentFlags().Float64("targetUtilization", 0.7, "utilization percentage you'd like to target")
	cobra.CheckErr(viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug")))
	cobra.CheckErr(viper.BindPFlag("targetUtilization", rootCmd.PersistentFlags().Lookup("targetUtilization")))

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// couldn't for the life of me figure out the right way to do this with viper
		cfgFile = "./config.yaml"
		viper.SetConfigFile(cfgFile)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
