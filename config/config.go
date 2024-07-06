package config

type Config struct {
	App struct {
		Name    string `mapstructure:"name"`
		Version string `mapstructure:"version"`
	} `mapstructure:"app"`
	Ethereum struct {
		Rpc    string `mapstructure:"rpc"`
		Delays struct {
			Wallet struct {
				Min int `mapstructure:"min"`
				Max int `mapstructure:"max"`
			} `mapstructure:"wallet"`
			Block struct {
				Min int `mapstructure:"min"`
				Max int `mapstructure:"max"`
			} `mapstructure:"block"`
		} `mapstructure:"delays"`
		Workflow struct {
			GweiLimit              int `mapstructure:"gweiLimit"`
			WorkAmountRangePercent struct {
				Min int `mapstructure:"min"`
				Max int `mapstructure:"max"`
			} `mapstructure:"workAmountRangePercent"`
		} `mapstructure:"workflow"`
	} `mapstructure:"ethereum"`
}
