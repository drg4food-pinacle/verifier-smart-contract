package types

type Config struct {
	GethNodeUrl             string `mapstructure:"GETH_NODE_URL" validate:"required,url"`
	GethNodeKeystore        string `mapstructure:"GETH_NODE_KEYSTORE" validate:"required,file_exists"`
	GethNodePassword        string `mapstructure:"GETH_NODE_PASSWORD" validate:"required"`
	AccountsDir             string `mapstructure:"ACCOUNTS_DIR" validate:"required"`
	AccountsNumber          int    `mapstructure:"ACCOUNTS_NUMBER" validate:"required,min=1"`
	AddressesDir            string `mapstructure:"CONTRACTS_ADDRESSES_DIR" validate:"required"`
	LoggerMode              string `mapstructure:"LOGGER_MODE" validate:"required,oneof=production development"`
	DisableBanner           bool   `mapstructure:"DISABLE_BANNER"`
	Version                 string `mapstructure:"VERSION" validate:"required,version"`
	WasmFilename            string `mapstructure:"ZK_WASM_FILENAME" validate:"required,file_exists"`
	ZkeyFilename            string `mapstructure:"ZK_ZKEY_FILENAME" validate:"required,file_exists"`
	VerificationKeyFilename string `mapstructure:"ZK_VERIFICATION_KEY_FILENAME" validate:"required,file_exists"`
}

func (Config) CustomErrorMessages() map[string]string {
	return map[string]string{
		"Config.Config.GethNodeUrl.required":                "Geth node URL is required",
		"Config.Config.GethNodeUrl.url":                     "Geth node URL must be a valid URL",
		"Config.Config.GethNodeKeystore.required":           "Geth node keystore path is required",
		"Config.Config.GethNodeKeystore.file_exists":        "Geth node keystore file must exist",
		"Config.Config.GethNodePassword.required":           "Geth node password is required",
		"Config.Config.AccountsDir.required":                "Accounts directory path is required",
		"Config.Config.AccountsNumber.required":             "Accounts number is required",
		"Config.Config.AccountsNumber.min":                  "Accounts number must be greater than 0",
		"Config.Config.AddressesDir.required":               "Addresses directory path is required",
		"Config.Config.LoggerMode.required":                 "Logger mode is required",
		"Config.Config.LoggerMode.oneof":                    "Logger mode must be either 'production' or 'development'",
		"Config.Config.Version.required":                    "Version is required",
		"Config.Config.Version.version":                     "Version must be a valid version string (e.g., v1.0.0)",
		"Config.Config.WasmFilename.required":               "ZK wasm filename is required",
		"Config.Config.WasmFilename.file_exists":            "ZK wasm file must exist",
		"Config.Config.ZkeyFilename.required":               "ZKey filename is required",
		"Config.Config.ZkeyFilename.file_exists":            "ZKey file must exist",
		"Config.Config.VerificationKeyFilename.required":    "Verification key filename is required",
		"Config.Config.VerificationKeyFilename.file_exists": "Verification key file must exist",
	}
}
