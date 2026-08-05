package dto

// MatchType: 0=含まれる, 1=全文一致, 2=前方一致, 3=後方一致
type CreatePlaceRuleRequest struct {
	Text       string  `json:"text"`
	MatchType  int     `json:"match_type"`
	Identifier *string `json:"identifier"`
	IsTransfer bool    `json:"is_transfer"`
}

type UpdatePlaceRuleRequest struct {
	ID         int64   `json:"id"`
	Text       string  `json:"text"`
	MatchType  int     `json:"match_type"`
	Identifier *string `json:"identifier"`
	IsTransfer bool    `json:"is_transfer"`
}

type PlaceRuleResponse struct {
	ID         int64   `json:"id"`
	Text       string  `json:"text"`
	MatchType  int     `json:"match_type"`
	Identifier *string `json:"identifier"`
	IsTransfer bool    `json:"is_transfer"`
}
