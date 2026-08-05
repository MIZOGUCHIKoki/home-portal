package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	"kakeibo/internal/model"
)

// nullableString converts a nil or empty *string to a SQL NULL.
func nullableString(s *string) sql.NullString {
	if s == nil || *s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

// nullStringToPtr converts a SQL NULL string to a *string.
func nullStringToPtr(s sql.NullString) *string {
	if !s.Valid {
		return nil
	}
	return &s.String
}

// CreatePlaceRule inserts a new place rule
func CreatePlaceRule(db *sql.DB, text string, matchType model.PlaceRuleMatchType, identifier *string, isTransfer bool) (int64, error) {
	if text == "" {
		return 0, errors.New("text is required")
	}

	var placeRuleID int64

	err := db.QueryRow(
		`INSERT INTO place_rules (text, match_type, identifier, is_transfer)
         VALUES ($1, $2, $3, $4)
         RETURNING place_rule_id`,
		text,
		matchType,
		nullableString(identifier),
		isTransfer,
	).Scan(&placeRuleID)

	if err != nil {
		return 0, err
	}

	return placeRuleID, nil
}

// GetPlaceRuleByID retrieves a place rule by ID
func GetPlaceRuleByID(db *sql.DB, placeRuleID int64) (*model.PlaceRule, error) {
	if placeRuleID <= 0 {
		return nil, errors.New("placeRuleID is required")
	}

	item := &model.PlaceRule{}
	var identifier sql.NullString

	err := db.QueryRow(
		`SELECT place_rule_id, text, match_type, identifier, is_transfer
         FROM place_rules
         WHERE place_rule_id = $1`,
		placeRuleID,
	).Scan(&item.PlaceRuleID, &item.Text, &item.MatchType, &identifier, &item.IsTransfer)

	if err != nil {
		return nil, err
	}

	item.Identifier = nullStringToPtr(identifier)

	return item, nil
}

// GetPlaceRuleByTextAndMatchType retrieves a place rule by its unique
// (text, match_type) key.
func GetPlaceRuleByTextAndMatchType(db *sql.DB, text string, matchType model.PlaceRuleMatchType) (*model.PlaceRule, error) {
	if text == "" {
		return nil, errors.New("text is required")
	}

	item := &model.PlaceRule{}
	var identifier sql.NullString

	err := db.QueryRow(
		`SELECT place_rule_id, text, match_type, identifier, is_transfer
         FROM place_rules
         WHERE text = $1 AND match_type = $2`,
		text,
		matchType,
	).Scan(&item.PlaceRuleID, &item.Text, &item.MatchType, &identifier, &item.IsTransfer)

	if err != nil {
		return nil, err
	}

	item.Identifier = nullStringToPtr(identifier)

	return item, nil
}

// ListPlaceRules returns all place rules
func ListPlaceRules(db *sql.DB) ([]model.PlaceRule, error) {
	rows, err := db.Query(
		`SELECT place_rule_id, text, match_type, identifier, is_transfer
         FROM place_rules
         ORDER BY place_rule_id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.PlaceRule

	for rows.Next() {
		var item model.PlaceRule
		var identifier sql.NullString

		if err := rows.Scan(
			&item.PlaceRuleID,
			&item.Text,
			&item.MatchType,
			&identifier,
			&item.IsTransfer,
		); err != nil {
			return nil, err
		}

		item.Identifier = nullStringToPtr(identifier)

		items = append(items, item)
	}

	return items, rows.Err()
}

// UpdatePlaceRule updates a place rule by ID
func UpdatePlaceRule(db *sql.DB, placeRuleID int64, text string, matchType model.PlaceRuleMatchType, identifier *string, isTransfer bool) error {
	if placeRuleID <= 0 {
		return errors.New("placeRuleID is required")
	}
	if text == "" {
		return errors.New("text is required")
	}

	result, err := db.Exec(
		`UPDATE place_rules
         SET text = $1, match_type = $2, identifier = $3, is_transfer = $4, updated_at = CURRENT_TIMESTAMP
         WHERE place_rule_id = $5`,
		text,
		matchType,
		nullableString(identifier),
		isTransfer,
		placeRuleID,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// DeletePlaceRule deletes a place rule by ID
func DeletePlaceRule(db *sql.DB, placeRuleID int64) error {
	if placeRuleID <= 0 {
		return errors.New("placeRuleID is required")
	}

	result, err := db.Exec(
		`DELETE FROM place_rules WHERE place_rule_id = $1`,
		placeRuleID,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// EnsurePlaceRule ensures a place rule exists (by text + match_type)
func EnsurePlaceRule(db *sql.DB, seed model.PlaceRuleSeed) (int64, error) {
	item, err := GetPlaceRuleByTextAndMatchType(db, seed.Text, seed.MatchType)
	if err == nil {
		return item.PlaceRuleID, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	return CreatePlaceRule(db, seed.Text, seed.MatchType, &seed.Identifier, seed.IsTransfer)
}

// SeedDefaultPlaceRules applies initial master data
func SeedDefaultPlaceRules(db *sql.DB) error {
	for _, r := range model.DefaultPlaceRules {
		if _, err := EnsurePlaceRule(db, r); err != nil {
			return fmt.Errorf("place rule apply failed (%s): %w", r.Text, err)
		}
	}
	return nil
}

// matchTypeSpecificity ranks match types so a more specific match wins
// when several rules match the same place (higher is more specific).
func matchTypeSpecificity(mt model.PlaceRuleMatchType) int {
	switch mt {
	case model.PlaceRuleMatchExact:
		return 3
	case model.PlaceRuleMatchPrefix, model.PlaceRuleMatchSuffix:
		return 2
	default: // PlaceRuleMatchContains
		return 1
	}
}

func placeRuleMatches(place string, rule model.PlaceRule) bool {
	target := strings.ToLower(place)
	text := strings.ToLower(rule.Text)

	switch rule.MatchType {
	case model.PlaceRuleMatchExact:
		return target == text
	case model.PlaceRuleMatchPrefix:
		return strings.HasPrefix(target, text)
	case model.PlaceRuleMatchSuffix:
		return strings.HasSuffix(target, text)
	default: // PlaceRuleMatchContains
		return strings.Contains(target, text)
	}
}

// MatchPlaceRule finds the rule that best matches place, according to
// each rule's MatchType (0: 含まれる, 1: 全文一致, 2: 前方一致, 3: 後方一致).
// When several rules match, the most specific match type wins; ties are
// broken by the longest rule text. Returns nil if no rule matches.
func MatchPlaceRule(db *sql.DB, place string) (*model.PlaceRule, error) {
	if place == "" {
		return nil, nil
	}

	rules, err := ListPlaceRules(db)
	if err != nil {
		return nil, err
	}

	var candidates []model.PlaceRule
	for _, rule := range rules {
		if rule.Text == "" {
			continue
		}
		if placeRuleMatches(place, rule) {
			candidates = append(candidates, rule)
		}
	}

	if len(candidates) == 0 {
		return nil, nil
	}

	sort.Slice(candidates, func(i, j int) bool {
		si, sj := matchTypeSpecificity(candidates[i].MatchType), matchTypeSpecificity(candidates[j].MatchType)
		if si != sj {
			return si > sj
		}
		return len(candidates[i].Text) > len(candidates[j].Text)
	})

	best := candidates[0]
	return &best, nil
}

// ApplyPlaceRule looks up a rule matching t.Place and, when found,
// overwrites t.CategoryID / t.IsTransfer with the rule's values and
// marks t.Rule as automatically set.
// It is a no-op when the transaction has no place or no rule matches.
func ApplyPlaceRule(db *sql.DB, t *model.Transaction) error {
	if t.Place == nil || *t.Place == "" {
		return nil
	}

	rule, err := MatchPlaceRule(db, *t.Place)
	if err != nil {
		return err
	}
	if rule == nil {
		return nil
	}

	t.IsTransfer = rule.IsTransfer
	if rule.Identifier != nil {
		category, err := GetCategoryByIdentifier(db, *rule.Identifier)
		if err != nil {
			return err
		}
		t.CategoryID = &category.CategoryID
	}
	t.Rule = true

	return nil
}
