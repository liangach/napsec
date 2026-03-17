package detector

type RuleCategory string

const (
	CategoryCredential RuleCategory = "凭证"
	CategoryPrivateKey RuleCategory = "私钥"
	CategoryPersonal   RuleCategory = "个人信息"
	CategoryFinancial  RuleCategory = "财务信息"
	CategoryConfig     RuleCategory = "配置文件"
)

// Rule 检测规则
type Rule struct {
	Name        string
	Category    RuleCategory
	Pattern     string   // 正则表达式
	Extensions  []string // 适用文件扩展名（为空则适用于所有文件）
	MinSeverity int      // 严重程度
	Description string
}

// DefaultRules 内置规则集
var DefaultRules = []Rule{
	// ========== 凭证类 ==========

	// AWS Key - 更宽松的模式
	{
		Name:        "AWS Access Key",
		Category:    CategoryCredential,
		Pattern:     `AKIA[0-9A-Z]{16}`,
		MinSeverity: 9,
		Description: "AWS 访问密钥 ID",
	},

	// AWS Key 环境变量格式
	{
		Name:        "AWS Access Key (环境变量)",
		Category:    CategoryCredential,
		Pattern:     `AWS_ACCESS_KEY[=:].*AKIA`,
		MinSeverity: 9,
		Description: "环境变量中的 AWS 密钥",
	},

	// 通用密码检测 - 宽松模式
	{
		Name:        "密码检测",
		Category:    CategoryCredential,
		Pattern:     `(?i)(password|passwd|pwd)[=:][^\n\r]+`,
		MinSeverity: 7,
		Description: "检测到可能的密码",
	},

	// 等号密码格式
	{
		Name:        "密码等号格式",
		Category:    CategoryCredential,
		Pattern:     `[=:][\s]*['"]?[^\s'"]+`,
		MinSeverity: 5,
		Description: "检测到等号格式的值",
	},

	// API Key
	{
		Name:        "API Key/Token",
		Category:    CategoryCredential,
		Pattern:     `(?i)(api[_-]?key|apikey|token|secret)[=:][^\n\r]+`,
		MinSeverity: 8,
		Description: "检测到可能的API密钥或令牌",
	},

	// 通用键值对检测
	{
		Name:        "敏感键值对",
		Category:    CategoryCredential,
		Pattern:     `(?i)(key|secret|token|pass|pwd)[=:][^\n\r]+`,
		MinSeverity: 6,
		Description: "检测到可能的敏感键值对",
	},

	// GitHub Token
	{
		Name:        "GitHub Token",
		Category:    CategoryCredential,
		Pattern:     `ghp_[0-9a-zA-Z]{36}|github_pat_[0-9a-zA-Z_]{82}`,
		MinSeverity: 9,
		Description: "GitHub 个人访问令牌",
	},

	// ========== 私钥类 ==========

	// 私钥检测
	{
		Name:        "私钥文件",
		Category:    CategoryPrivateKey,
		Pattern:     `-----BEGIN.*PRIVATE KEY-----`,
		MinSeverity: 10,
		Description: "检测到私钥文件",
	},

	// ========== 个人信息类 ==========

	// 身份证号
	{
		Name:        "中国身份证号",
		Category:    CategoryPersonal,
		Pattern:     `[1-9]\d{5}(18|19|20)\d{2}(0[1-9]|1[0-2])(0[1-9]|[12]\d|3[01])\d{3}[\dXx]`,
		MinSeverity: 8,
		Description: "中国居民身份证号码",
	},

	// 手机号
	{
		Name:        "手机号码",
		Category:    CategoryPersonal,
		Pattern:     `1[3-9]\d{9}`,
		MinSeverity: 5,
		Description: "中国大陆手机号码",
	},

	// 邮箱
	{
		Name:        "Email地址",
		Category:    CategoryPersonal,
		Pattern:     `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`,
		MinSeverity: 3,
		Description: "电子邮件地址",
	},

	// ========== 配置文件类 ==========

	// 配置文件中的密码
	{
		Name:        "配置文件中的密码",
		Category:    CategoryConfig,
		Pattern:     `(?i)(password|pass|pwd)[=:][^\n\r]+`,
		MinSeverity: 7,
		Description: "配置文件中的密码字段",
	},

	// 连接字符串
	{
		Name:        "连接字符串",
		Category:    CategoryConfig,
		Pattern:     `[a-zA-Z]+://[^:]+:[^@]+@`,
		MinSeverity: 9,
		Description: "数据库连接字符串（含密码）",
	},

	// 环境变量中的敏感信息
	{
		Name:        "环境变量中的敏感信息",
		Category:    CategoryConfig,
		Pattern:     `(?i)(KEY|SECRET|PASS|TOKEN)=.+`,
		MinSeverity: 8,
		Description: ".env 文件中的敏感变量",
	},
}
