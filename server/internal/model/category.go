package model

// seed用
type CategorySeed struct {
	Identifier   string
	CategoryName string
}

// DB用
type Category struct {
	CategoryID   int64
	Identifier   string
	CategoryName string
}

var DefaultCategories = []CategorySeed{
	// 食費
	{
		Identifier:   "food",
		CategoryName: "食費（スーパー）",
	},
	{
		Identifier:   "dining",
		CategoryName: "外食",
	},
	{
		Identifier:   "convenience",
		CategoryName: "コンビニ",
	},
	{
		Identifier:   "daily",
		CategoryName: "日用品",
	},

	// 固定費
	{
		Identifier:   "rent",
		CategoryName: "家賃",
	},
	{
		Identifier:   "utilities",
		CategoryName: "光熱費",
	},
	{
		Identifier:   "communication",
		CategoryName: "通信費",
	},
	{
		Identifier:   "subscription",
		CategoryName: "サブスクリプション",
	},

	// 移動
	{
		Identifier:   "transport",
		CategoryName: "交通費",
	},

	// 消費・娯楽
	{
		Identifier:   "shopping",
		CategoryName: "買い物",
	},
	{
		Identifier:   "entertainment",
		CategoryName: "娯楽",
	},

	// 健康
	{
		Identifier:   "health",
		CategoryName: "医療・健康",
	},

	// 教育
	{
		Identifier:   "education",
		CategoryName: "教育",
	},

	// 投資
	{
		Identifier:   "investment",
		CategoryName: "投資",
	},
	{
		Identifier:   "savings",
		CategoryName: "貯金",
	},

	// 収入
	{
		Identifier:   "salary",
		CategoryName: "給与",
	},
	{
		Identifier:   "bonus",
		CategoryName: "ボーナス",
	},
	{
		Identifier:   "allowance",
		CategoryName: "お小遣い",
	},
	{
		Identifier:   "reimbursement",
		CategoryName: "立替精算",
	},
	{
		Identifier:   "other_income",
		CategoryName: "その他収入",
	},

	// その他
	{
		Identifier:   "misc",
		CategoryName: "雑費",
	},
}
