package types

// Config holds the configuration settings for the project
type Config struct {
	ContractsBaseDir           string `mapstructure:"CONTRACTS_BASE_DIR" validate:"required,dir"`
	ContractsDir               string `mapstructure:"CONTRACTS_DIR" validate:"required"`
	ContractsNames             string `mapstructure:"CONTRACTS_NAME" validate:"required"`
	ContractsAllowMultiple     bool   `mapstructure:"CONTRACTS_ALLOW_MULTIPLE"`
	ContractsGroupDelimeter    string `mapstructure:"CONTRACTS_GROUP_DELIMETER" validate:"required_if=ContractsAllowMultiple true"`
	ContractsSubGroupDelimeter string `mapstructure:"CONTRACTS_SUBGROUP_DELIMETER" validate:"required_if=ContractsAllowMultiple true"`
	ContractsRemappings        string `mapstructure:"CONTRACTS_REMAPPINGS"`
	SolcVersion                string `mapstructure:"SOLC_VERSION" validate:"required,version,solc_version"`
	SolcOptimizer              bool   `mapstructure:"SOLC_OPTIMIZER"`
	SolcOptimizerRuns          int    `mapstructure:"SOLC_OPTIMIZER_RUNS" validate:"required_if=SolcOptimizer true"`
	SolcOutputDir              string `mapstructure:"SOLC_OUTPUT_DIR" validate:"required"`
	SolcViaIR                  bool   `mapstructure:"SOLC_VIA_IR"`
	AbigenVersion              string `mapstructure:"ABIGEN_VERSION" validate:"required,version"`
	AbigenOutputDir            string `mapstructure:"ABIGEN_OUTPUT_DIR" validate:"required"`
	LoggerMode                 string `mapstructure:"LOGGER_MODE" validate:"required,oneof=production development"`
	DisableBanner              bool   `mapstructure:"DISABLE_BANNER"`
	Version                    string `mapstructure:"VERSION" validate:"required,version"`
	OutputSelection            string `mapstructure:"SOLC_OUTPUT_SELECTION"`
	CleanMode                  bool   `mapstructure:"CLEAN_MODE"`
}

func (c Config) CustomErrorMessages() map[string]string {
	return map[string]string{
		// Contracts
		"Config.Config.ContractsBaseDir.required":              "Contracts base directory is required",
		"Config.Config.ContractsBaseDir.dir":                   "Contracts base directory must be a valid directory path",
		"Config.Config.ContractsDir.required":                  "Contracts directory is required",
		"Config.Config.ContractsNames.required":                "Contracts name is required",
		"Config.Config.ContractsGroupDelimeter.required_if":    "Contracts group delimiter is required when multiple contracts are allowed",
		"Config.Config.ContractsSubGroupDelimeter.required_if": "Contracts subgroup delimiter is required when multiple contracts are allowed",
		// Solc
		"Config.Config.SolcVersion.required":          "Solidity compiler version is required",
		"Config.Config.SolcVersion.version":           "Solidity compiler version must follow semantic versioning",
		"Config.Config.SolcVersion.solc_version":      "Solidity compiler version must be a valid supported solc version",
		"Config.Config.SolcOptimizerRuns.required_if": "Optimizer runs count is required when Solidity optimizer is enabled",
		"Config.Config.SolcOutputDir.required":        "Solidity output directory is required",
		// Abigen
		"Config.Config.AbigenVersion.required":   "Abigen version is required",
		"Config.Config.AbigenVersion.version":    "Abigen version must follow semantic versioning",
		"Config.Config.AbigenOutputDir.required": "Abigen output directory is required",
		// Logger
		"Config.Config.LoggerMode.required": "Logger mode is required",
		"Config.Config.LoggerMode.oneof":    "Logger mode must be either 'production' or 'development'",
		// Version
		"Config.Config.Version.required": "Application version is required",
		"Config.Config.Version.version":  "Application version must follow semantic versioning",
	}
}
