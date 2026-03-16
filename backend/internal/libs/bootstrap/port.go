package bootstrap

import "fmt"

func InitPortStringFromConfig(cfg Config, fallbackPort int) string {
	if cfg.Port != 0 {
		return fmt.Sprintf(":%v", cfg.Port)
	}
	return fmt.Sprintf(":%v", fallbackPort)
}
