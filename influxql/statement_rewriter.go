package influxql

import (
	"errors"
)

// RewriteStatement rewrites stmt into a new statement, if applicable.
func RewriteStatement(stmt Statement) (Statement, error) {
	switch stmt := stmt.(type) {
	case *ShowFieldKeysStatement:
		return rewriteShowFieldKeysStatement(stmt)
	case *ShowMeasurementsStatement:
		return rewriteShowMeasurementsStatement(stmt)
	case *ShowTagKeysStatement:
		return rewriteShowTagKeysStatement(stmt)
	case *ShowTagValuesStatement:
		return rewriteShowTagValuesStatement(stmt)
	default:
		return stmt, nil
	}
}

func rewriteShowFieldKeysStatement(stmt *ShowFieldKeysStatement) (Statement, error) {
	var condition Expr
	if len(stmt.Sources) > 0 {
		if source, ok := stmt.Sources[0].(*Measurement); ok {
			if source.Regex != nil {
				condition = &BinaryExpr{
					Op:  EQREGEX,
					LHS: &VarRef{Val: "name"},
					RHS: &RegexLiteral{Val: source.Regex.Val},
				}
			} else if source.Name != "" {
				condition = &BinaryExpr{
					Op:  EQ,
					LHS: &VarRef{Val: "name"},
					RHS: &StringLiteral{Val: source.Name},
				}
			}
		}
	}

	return &SelectStatement{
		Fields: Fields([]*Field{
			{Expr: &VarRef{Val: "fieldKey"}},
		}),
		Sources: Sources([]Source{
			&Measurement{Name: "_fieldKeys"},
		}),
		Condition:  condition,
		Offset:     stmt.Offset,
		Limit:      stmt.Limit,
		SortFields: stmt.SortFields,
		OmitTime:   true,
		Dedupe:     true,
	}, nil
}

func rewriteShowMeasurementsStatement(stmt *ShowMeasurementsStatement) (Statement, error) {
	// Check for time in WHERE clause (not supported).
	if HasTimeExpr(stmt.Condition) {
		return nil, errors.New("SHOW MEASUREMENTS doesn't support time in WHERE clause")
	}

	condition := stmt.Condition
	if source, ok := stmt.Source.(*Measurement); ok {
		var expr Expr
		if source.Regex != nil {
			expr = &BinaryExpr{
				Op:  EQREGEX,
				LHS: &VarRef{Val: "name"},
				RHS: &RegexLiteral{Val: source.Regex.Val},
			}
		} else if source.Name != "" {
			expr = &BinaryExpr{
				Op:  EQ,
				LHS: &VarRef{Val: "name"},
				RHS: &StringLiteral{Val: source.Name},
			}
		}

		// Set condition or "AND" together.
		if condition == nil {
			condition = expr
		} else {
			condition = &BinaryExpr{Op: AND, LHS: expr, RHS: condition}
		}
	}

	return &SelectStatement{
		Fields: Fields([]*Field{
			{Expr: &VarRef{Val: "name"}},
		}),
		Sources: Sources([]Source{
			&Measurement{Name: "_measurements"},
		}),
		Condition:  condition,
		Offset:     stmt.Offset,
		Limit:      stmt.Limit,
		SortFields: stmt.SortFields,
		OmitTime:   true,
		Dedupe:     true,
	}, nil
}

func rewriteShowTagKeysStatement(stmt *ShowTagKeysStatement) (Statement, error) {
	// Check for time in WHERE clause (not supported).
	if HasTimeExpr(stmt.Condition) {
		return nil, errors.New("SHOW TAG KEYS doesn't support time in WHERE clause")
	}

	condition := stmt.Condition
	if len(stmt.Sources) > 0 {
		if source, ok := stmt.Sources[0].(*Measurement); ok {
			var expr Expr
			if source.Regex != nil {
				expr = &BinaryExpr{
					Op:  EQREGEX,
					LHS: &VarRef{Val: "name"},
					RHS: &RegexLiteral{Val: source.Regex.Val},
				}
			} else if source.Name != "" {
				expr = &BinaryExpr{
					Op:  EQ,
					LHS: &VarRef{Val: "name"},
					RHS: &StringLiteral{Val: source.Name},
				}
			}

			// Set condition or "AND" together.
			if condition == nil {
				condition = expr
			} else {
				condition = &BinaryExpr{Op: AND, LHS: expr, RHS: condition}
			}
		}
	}

	return &SelectStatement{
		Fields: []*Field{
			{Expr: &VarRef{Val: "tagKey"}},
		},
		Sources: []Source{
			&Measurement{Name: "_tagKeys"},
		},
		Condition:  condition,
		Offset:     stmt.Offset,
		Limit:      stmt.Limit,
		SortFields: stmt.SortFields,
		OmitTime:   true,
		Dedupe:     true,
	}, nil
}

func rewriteShowTagValuesStatement(stmt *ShowTagValuesStatement) (Statement, error) {
	// Check for time in WHERE clause (not supported).
	if HasTimeExpr(stmt.Condition) {
		return nil, errors.New("SHOW TAG VALUES doesn't support time in WHERE clause")
	}

	condition := stmt.Condition
	if len(stmt.Sources) > 0 {
		if source, ok := stmt.Sources[0].(*Measurement); ok {
			var expr Expr
			if source.Regex != nil {
				expr = &BinaryExpr{
					Op:  EQREGEX,
					LHS: &VarRef{Val: "name"},
					RHS: &RegexLiteral{Val: source.Regex.Val},
				}
			} else if source.Name != "" {
				expr = &BinaryExpr{
					Op:  EQ,
					LHS: &VarRef{Val: "name"},
					RHS: &StringLiteral{Val: source.Name},
				}
			}

			// Set condition or "AND" together.
			if condition == nil {
				condition = expr
			} else {
				condition = &BinaryExpr{Op: AND, LHS: expr, RHS: condition}
			}
		}
	}

	if len(stmt.TagKeys) > 0 {
		var expr Expr
		for _, tagKey := range stmt.TagKeys {
			if expr != nil {
				expr = &BinaryExpr{
					Op:  OR,
					LHS: expr,
					RHS: &BinaryExpr{
						Op:  EQ,
						LHS: &VarRef{Val: "key"},
						RHS: &StringLiteral{Val: tagKey},
					},
				}
			} else {
				expr = &BinaryExpr{
					Op:  EQ,
					LHS: &VarRef{Val: "key"},
					RHS: &StringLiteral{Val: tagKey},
				}
			}
		}

		// Wrap in parenthesis if we have more than 1 tag key.
		if len(stmt.TagKeys) > 1 {
			expr = &ParenExpr{Expr: expr}
		}

		// Set condition or "AND" together.
		if condition == nil {
			condition = expr
		} else {
			condition = &BinaryExpr{Op: AND, LHS: condition, RHS: expr}
		}
	}

	return &SelectStatement{
		Fields: []*Field{
			{Expr: &VarRef{Val: "key"}},
			{Expr: &VarRef{Val: "value"}},
		},
		Sources: []Source{
			&Measurement{Name: "_tags"},
		},
		Condition:  condition,
		Offset:     stmt.Offset,
		Limit:      stmt.Limit,
		SortFields: stmt.SortFields,
		OmitTime:   true,
		Dedupe:     true,
	}, nil
}
