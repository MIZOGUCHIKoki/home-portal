package model

// PlaceRuleMatchType decides how PlaceRule.Text is compared against a
// transaction's place.
type PlaceRuleMatchType int

const (
	PlaceRuleMatchContains PlaceRuleMatchType = 0 // 含まれる
	PlaceRuleMatchExact    PlaceRuleMatchType = 1 // 全文一致
	PlaceRuleMatchPrefix   PlaceRuleMatchType = 2 // 前方一致
	PlaceRuleMatchSuffix   PlaceRuleMatchType = 3 // 後方一致
)

// PlaceRule maps text matched against a transaction's place to a
// category and/or transfer flag, so imports can auto-categorize
// transactions without manual editing.
type PlaceRule struct {
	PlaceRuleID int64
	Text        string
	MatchType   PlaceRuleMatchType
	Identifier  *string
	IsTransfer  bool
}

// seed用
type PlaceRuleSeed struct {
	Text       string
	MatchType  PlaceRuleMatchType
	Identifier string
	IsTransfer bool
}

// seed用（同じ条件のtextをまとめて書くためのグループ）
type PlaceRuleSeedGroup struct {
	MatchType  PlaceRuleMatchType
	Identifier string
	IsTransfer bool
	Texts      []string
}

var defaultPlaceRuleGroups = []PlaceRuleSeedGroup{
	{
		MatchType:  PlaceRuleMatchContains,
		Identifier: "cafe",
		Texts:      []string{"イデカフェ", "スターバックス", "ＤＣＳ", "タリーズコーヒー"},
	},
	{
		MatchType:  PlaceRuleMatchExact,
		Identifier: "communication",
		Texts:      []string{"ｐｏｖｏご利用料金", "楽天モバイル通信料", "インターネットイニシアティブ"},
	},
	{
		MatchType:  PlaceRuleMatchExact,
		IsTransfer: true,
		Texts: []string{
			"口座振替　ビューカード",
			"口座振替　ミツビシユーエフジェイニコス",
			"口座振替　（カ）ジエーシービー",
			"振込＊ミゾグチ　コウキ",
		},
	},
	{
		MatchType:  PlaceRuleMatchContains,
		IsTransfer: true,
		Texts: []string{
			"ことら送金　ミゾグチ　コウキ",
			"ＡＴＭ",
			"フリカエ　ＰＡＹＰＡＹ",
		},
	},
	{
		MatchType:  PlaceRuleMatchExact,
		Identifier: "interest",
		Texts:      []string{"利息"},
	},
	{
		MatchType:  PlaceRuleMatchContains,
		Identifier: "transport",
		Texts:      []string{"ＪＡＬ", "ＥＴＣ"},
	},
	{
		MatchType:  PlaceRuleMatchContains,
		Identifier: "accommodation",
		Texts:      []string{"東横ＩＮＮ"},
	},
	{
		MatchType:  PlaceRuleMatchExact,
		Identifier: "education",
		Texts:      []string{"株式会社レアジョブ"},
	},
	{
		MatchType:  PlaceRuleMatchExact,
		Identifier: "subscription",
		Texts: []string{
			"ＮＥＴＦＬＩＸ．ＣＯＭ",
			"ＡＰＰＬＥ．ＣＯＭ／ＢＩＬＬ",
		},
	},
	{
		MatchType:  PlaceRuleMatchExact,
		Identifier: "dining",
		Texts: []string{
			"サイゼリヤ",
			"鳥貴族",
		},
	},
	{
		MatchType:  PlaceRuleMatchContains,
		Identifier: "supermarket",
		Texts: []string{
			"イトーヨーカドー",
			"東武ストア",
			"ベルク"},
	},
	{
		MatchType:  PlaceRuleMatchContains,
		Identifier: "tax",
		Texts:      []string{"地方税", "国税"},
	},
	{
		MatchType:  PlaceRuleMatchContains,
		Identifier: "convenience",
		Texts: []string{
			"セブン－イレブン",
			"ファミリーマート",
			"ローソン",
			"ＮｅｗＤａｙｓ・ＫＩＯＳＫ",
		},
	},
	{
		MatchType:  PlaceRuleMatchExact,
		Identifier: "investment",
		Texts:      []string{"ＳＢＩハイブリッド預金"},
	},
}

var DefaultPlaceRules = expandPlaceRuleGroups(defaultPlaceRuleGroups)

func expandPlaceRuleGroups(groups []PlaceRuleSeedGroup) []PlaceRuleSeed {
	var seeds []PlaceRuleSeed
	for _, g := range groups {
		for _, text := range g.Texts {
			seeds = append(seeds, PlaceRuleSeed{
				Text:       text,
				MatchType:  g.MatchType,
				Identifier: g.Identifier,
				IsTransfer: g.IsTransfer,
			})
		}
	}
	return seeds
}
