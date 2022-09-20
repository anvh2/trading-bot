package bot

type ExchangeConfig struct {
	Name             string            `yaml:"name"`              // Represents the exchange name.
	PublicKey        string            `yaml:"public_key"`        // Represents the public key used to connect to Exchange API.
	SecretKey        string            `yaml:"secret_key"`        // Represents the secret key used to connect to Exchange API.
	DepositAddresses map[string]string `yaml:"deposit_addresses"` // Represents the bindings between coins and deposit address on the exchange.
}

type BotConfig struct {
	ExchangeConfig *ExchangeConfig `yaml:"exchange_config"` // Represents the current exchange configuration.
}
