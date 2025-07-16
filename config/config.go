package config

type Config struct {
	DB struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	WechatPay struct {
		AppID     string `json:"app_id"`
		MchID     string `json:"mch_id"`
		APIKey    string `json:"api_key"`
		NotifyURL string `json:"notify_url"`
	}
	Server struct {
		Port string `json:"port"`
	}
}

// Load 加载配置（实际生产环境应从环境变量或配置文件读取）
func Load() *Config {
	cfg := &Config{}
	// 配置数据库信息（根据您提供的信息）
	cfg.DB.Host = "localhost"
	cfg.DB.Port = "3306"
	cfg.DB.User = "root"
	cfg.DB.Password = "root"
	cfg.DB.Name = "photo_wall"
	
	// 添加微信支付配置
	cfg.WechatPay.AppID = "wx607862131d2a8cfd"
	cfg.WechatPay.MchID = "1677234195"
	cfg.WechatPay.APIKey = "dx06787019930806dx06787019930806"
	cfg.WechatPay.NotifyURL = "/api/payment/notify"
	
	cfg.Server.Port = "8080"
	return cfg
}