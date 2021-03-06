package environment

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"k8s.io/client-go/tools/clientcmd"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/log"
)

// AirshipCTLSettings is a container for all of the settings needed by airshipctl
type AirshipCTLSettings struct {
	// Debug is used for verbose output
	Debug             bool
	airshipConfigPath string
	kubeConfigPath    string
	config            *config.Config
}

// InitFlags adds the default settings flags to cmd
func (a *AirshipCTLSettings) InitFlags(cmd *cobra.Command) {
	flags := cmd.PersistentFlags()
	flags.BoolVar(
		&a.Debug,
		"debug",
		false,
		"enable verbose output")

	defaultAirshipConfigDir := filepath.Join(HomeEnvVar, config.AirshipConfigDir)

	defaultAirshipConfigPath := filepath.Join(defaultAirshipConfigDir, config.AirshipConfig)
	flags.StringVar(
		&a.airshipConfigPath,
		config.FlagConfigFilePath,
		"",
		`Path to file for airshipctl configuration. (default "`+defaultAirshipConfigPath+`")`)

	defaultKubeConfigPath := filepath.Join(defaultAirshipConfigDir, config.AirshipKubeConfig)
	flags.StringVar(
		&a.kubeConfigPath,
		clientcmd.RecommendedConfigPathFlag,
		"",
		`Path to kubeconfig associated with airshipctl configuration. (default "`+defaultKubeConfigPath+`")`)
}

func (a *AirshipCTLSettings) Config() *config.Config {
	return a.config
}

func (a *AirshipCTLSettings) SetConfig(conf *config.Config) {
	a.config = conf
}

func (a *AirshipCTLSettings) AirshipConfigPath() string {
	return a.airshipConfigPath
}

func (a *AirshipCTLSettings) SetAirshipConfigPath(acp string) {
	a.airshipConfigPath = acp
}

func (a *AirshipCTLSettings) KubeConfigPath() string {
	return a.kubeConfigPath
}

func (a *AirshipCTLSettings) SetKubeConfigPath(kcp string) {
	a.kubeConfigPath = kcp
}

// InitConfig - Initializes and loads Config it exists.
func (a *AirshipCTLSettings) InitConfig() {
	a.SetConfig(config.NewConfig())

	a.initAirshipConfigPath()
	a.initKubeConfigPath()

	err := a.Config().LoadConfig(a.AirshipConfigPath(), a.KubeConfigPath())
	if err != nil {
		// Should stop airshipctl
		log.Fatal(err)
	}
}

func (a *AirshipCTLSettings) initAirshipConfigPath() {
	// The airshipConfigPath may already have been received as a command line argument
	if a.airshipConfigPath != "" {
		return
	}

	// Otherwise, we can check if we got the path via ENVIRONMENT variable
	a.airshipConfigPath = os.Getenv(config.AirshipConfigEnv)
	if a.airshipConfigPath != "" {
		return
	}

	// Otherwise, we'll try putting it in the home directory
	homeDir := userHomeDir()
	a.airshipConfigPath = filepath.Join(homeDir, config.AirshipConfigDir, config.AirshipConfig)
}

func (a *AirshipCTLSettings) initKubeConfigPath() {
	// NOTE(howell): This function will set the kubeConfigPath to the
	// default location under the airship directory unless the user
	// *explicitly* specifies a different location, either by setting the
	// ENVIRONMENT variable or by passing a command line argument.
	// This avoids messing up the user's kubeconfig if they didn't
	// explicitly want airshipctl to use it.

	// The kubeConfigPath may already have been received as a command line argument
	if a.kubeConfigPath != "" {
		return
	}

	// Otherwise, we can check if we got the path via ENVIRONMENT variable
	a.kubeConfigPath = os.Getenv(config.AirshipKubeConfigEnv)
	if a.kubeConfigPath != "" {
		return
	}

	// Otherwise, we'll try putting it in the home directory
	homeDir := userHomeDir()
	a.kubeConfigPath = filepath.Join(homeDir, config.AirshipConfigDir, config.AirshipKubeConfig)
}

// userHomeDir is a utility function that wraps os.UserHomeDir and returns no
// errors. If the user has no home directory, the returned value will be the
// empty string
func userHomeDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = ""
	}
	return homeDir
}
