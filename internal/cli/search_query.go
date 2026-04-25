package cli

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/maolonglong/workflowy-go/pkg/workflowy"
)

type compiledSearchQuery struct {
	segments []*searchExpr
}

var searchNow = time.Now

func compileSearchQuery(raw string) (*compiledSearchQuery, error) {
	tokens, err := tokenizeSearchQuery(raw)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, fmt.Errorf("workflowy: invalid search query: query is empty")
	}

	parser := &searchParser{tokens: tokens}
	query, err := parser.parseQuery()
	if err != nil {
		return nil, err
	}
	if parser.hasNext() {
		return nil, fmt.Errorf("workflowy: invalid search query: unexpected token %q", parser.peek().value)
	}
	return query, nil
}

func (q *compiledSearchQuery) Filter(nodes []*workflowy.Node) []*workflowy.Node {
	lookup := make(map[string]*workflowy.Node, len(nodes))
	for _, node := range nodes {
		lookup[node.ID] = node
	}

	matches := make([]*workflowy.Node, 0, len(nodes))
	for _, node := range nodes {
		if q.matchesNode(node, lookup) {
			matches = append(matches, node)
		}
	}
	return matches
}

func (q *compiledSearchQuery) matchesNode(node *workflowy.Node, lookup map[string]*workflowy.Node) bool {
	if len(q.segments) == 0 {
		return false
	}

	searchNode := newSearchNode(node)
	if !q.segments[len(q.segments)-1].Match(searchNode) {
		return false
	}
	if len(q.segments) == 1 {
		return true
	}

	ancestors := ancestorChain(node, lookup)
	want := 0
	for _, ancestor := range ancestors {
		if q.segments[want].Match(newSearchNode(ancestor)) {
			want++
			if want == len(q.segments)-1 {
				return true
			}
		}
	}
	return false
}

func ancestorChain(node *workflowy.Node, lookup map[string]*workflowy.Node) []*workflowy.Node {
	var reversed []*workflowy.Node
	seen := map[string]struct{}{node.ID: {}}
	current := node
	for current.ParentID != nil && *current.ParentID != "" {
		parent, ok := lookup[*current.ParentID]
		if !ok {
			break
		}
		if _, ok := seen[parent.ID]; ok {
			break
		}
		reversed = append(reversed, parent)
		seen[parent.ID] = struct{}{}
		current = parent
	}

	ancestors := make([]*workflowy.Node, len(reversed))
	for i := range reversed {
		ancestors[i] = reversed[len(reversed)-1-i]
	}
	return ancestors
}

type searchNode struct {
	node         *workflowy.Node
	nameLower    string
	noteLower    string
	contentLower string
}

func newSearchNode(node *workflowy.Node) *searchNode {
	name := strings.ToLower(node.Name)
	note := ""
	if node.Note != nil {
		note = strings.ToLower(*node.Note)
	}
	content := name
	if note != "" {
		content += "\n" + note
	}
	return &searchNode{
		node:         node,
		nameLower:    name,
		noteLower:    note,
		contentLower: content,
	}
}

type searchExpr struct {
	clauses []*searchClause
}

func (e *searchExpr) Match(node *searchNode) bool {
	for _, clause := range e.clauses {
		if clause.Match(node) {
			return true
		}
	}
	return false
}

type searchClause struct {
	terms []*searchTerm
}

func (c *searchClause) Match(node *searchNode) bool {
	for _, term := range c.terms {
		matched := term.matcher.Match(node)
		if term.negated {
			matched = !matched
		}
		if !matched {
			return false
		}
	}
	return true
}

type searchTerm struct {
	negated bool
	matcher searchMatcher
}

type searchMatcher interface {
	Match(node *searchNode) bool
}

type textMatcher struct {
	needle string
}

func (m textMatcher) Match(node *searchNode) bool {
	return strings.Contains(node.contentLower, m.needle)
}

type noteMatcher struct{}

func (noteMatcher) Match(node *searchNode) bool {
	return node.node.Note != nil
}

type layoutMatcher struct {
	layout workflowy.LayoutMode
}

func (m layoutMatcher) Match(node *searchNode) bool {
	return node.node.Data.LayoutMode == m.layout
}

type completedMatcher struct{}

func (completedMatcher) Match(node *searchNode) bool {
	return node.node.CompletedAt != nil || (node.node.Completed != nil && *node.node.Completed)
}

type timeField int

const (
	timeFieldCreated timeField = iota
	timeFieldModified
)

type recentTimestampMatcher struct {
	field timeField
	since time.Time
}

func (m recentTimestampMatcher) Match(node *searchNode) bool {
	var ts time.Time
	switch m.field {
	case timeFieldCreated:
		ts = node.node.CreatedAt.Time
	case timeFieldModified:
		ts = node.node.ModifiedAt.Time
	default:
		return false
	}
	if ts.IsZero() {
		return false
	}
	return !ts.Before(m.since)
}

type searchTokenKind int

const (
	searchTokenWord searchTokenKind = iota
	searchTokenQuoted
	searchTokenMinus
	searchTokenGT
	searchTokenOR
)

type searchToken struct {
	kind  searchTokenKind
	value string
}

func tokenizeSearchQuery(raw string) ([]searchToken, error) {
	var tokens []searchToken
	for i := 0; i < len(raw); {
		switch raw[i] {
		case ' ', '\t', '\n', '\r':
			i++
		case '>':
			tokens = append(tokens, searchToken{kind: searchTokenGT, value: ">"})
			i++
		case '-':
			tokens = append(tokens, searchToken{kind: searchTokenMinus, value: "-"})
			i++
		case '"':
			i++
			start := i
			for i < len(raw) && raw[i] != '"' {
				i++
			}
			if i >= len(raw) {
				return nil, fmt.Errorf("workflowy: invalid search query: unterminated quoted phrase")
			}
			tokens = append(tokens, searchToken{kind: searchTokenQuoted, value: raw[start:i]})
			i++
		default:
			start := i
			for i < len(raw) && !isSearchDelimiter(raw[i]) {
				i++
			}
			word := raw[start:i]
			kind := searchTokenWord
			if strings.EqualFold(word, "OR") {
				kind = searchTokenOR
			}
			tokens = append(tokens, searchToken{kind: kind, value: word})
		}
	}
	return tokens, nil
}

func isSearchDelimiter(ch byte) bool {
	switch ch {
	case ' ', '\t', '\n', '\r', '>', '"':
		return true
	default:
		return false
	}
}

type searchParser struct {
	tokens []searchToken
	pos    int
}

func (p *searchParser) parseQuery() (*compiledSearchQuery, error) {
	var segments []*searchExpr
	for {
		if !p.hasNext() {
			break
		}
		if p.peek().kind == searchTokenGT {
			return nil, fmt.Errorf("workflowy: invalid search query: empty segment before %q", p.peek().value)
		}

		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		segments = append(segments, expr)

		if !p.hasNext() {
			break
		}
		if p.peek().kind != searchTokenGT {
			break
		}
		p.next()
		if !p.hasNext() {
			return nil, fmt.Errorf("workflowy: invalid search query: empty segment after %q", ">")
		}
	}

	return &compiledSearchQuery{segments: segments}, nil
}

func (p *searchParser) parseExpr() (*searchExpr, error) {
	clause, err := p.parseClause()
	if err != nil {
		return nil, err
	}

	expr := &searchExpr{clauses: []*searchClause{clause}}
	for p.hasNext() && p.peek().kind == searchTokenOR {
		p.next()
		nextClause, err := p.parseClause()
		if err != nil {
			return nil, err
		}
		expr.clauses = append(expr.clauses, nextClause)
	}
	return expr, nil
}

func (p *searchParser) parseClause() (*searchClause, error) {
	var terms []*searchTerm
	for p.hasNext() {
		switch p.peek().kind {
		case searchTokenOR, searchTokenGT:
			if len(terms) == 0 {
				return nil, fmt.Errorf("workflowy: invalid search query: missing search term before %q", p.peek().value)
			}
			return &searchClause{terms: terms}, nil
		default:
			term, err := p.parseTerm()
			if err != nil {
				return nil, err
			}
			terms = append(terms, term)
		}
	}
	if len(terms) == 0 {
		return nil, fmt.Errorf("workflowy: invalid search query: missing search term")
	}
	return &searchClause{terms: terms}, nil
}

func (p *searchParser) parseTerm() (*searchTerm, error) {
	negated := false
	for p.hasNext() && p.peek().kind == searchTokenMinus {
		negated = !negated
		p.next()
	}
	if !p.hasNext() {
		return nil, fmt.Errorf("workflowy: invalid search query: dangling %q", "-")
	}

	token := p.next()
	matcher, err := compileTokenMatcher(token)
	if err != nil {
		return nil, err
	}
	return &searchTerm{negated: negated, matcher: matcher}, nil
}

func compileTokenMatcher(token searchToken) (searchMatcher, error) {
	switch token.kind {
	case searchTokenQuoted:
		needle := strings.ToLower(strings.TrimSpace(token.value))
		if needle == "" {
			return nil, fmt.Errorf("workflowy: invalid search query: empty quoted phrase")
		}
		return textMatcher{needle: needle}, nil
	case searchTokenWord:
		return compileWordMatcher(token.value)
	default:
		return nil, fmt.Errorf("workflowy: invalid search query: unexpected token %q", token.value)
	}
}

func compileWordMatcher(word string) (searchMatcher, error) {
	lower := strings.ToLower(strings.TrimSpace(word))
	if lower == "" {
		return nil, fmt.Errorf("workflowy: invalid search query: empty search term")
	}

	if key, value, ok := strings.Cut(lower, ":"); ok {
		switch key {
		case "is":
			return compileIsMatcher(value)
		case "has":
			return compileHasMatcher(value)
		case "changed":
			return compileRecentTimestampMatcher(value, timeFieldModified)
		case "created":
			return compileRecentTimestampMatcher(value, timeFieldCreated)
		case "text", "highlight", "in", "date-after":
			return nil, fmt.Errorf("workflowy: invalid search query: %s: is not supported by exported Workflowy data", key)
		default:
			return nil, fmt.Errorf("workflowy: invalid search query: unsupported operator %q", key+":")
		}
	}

	return textMatcher{needle: lower}, nil
}

func compileIsMatcher(value string) (searchMatcher, error) {
	switch value {
	case "todo":
		return layoutMatcher{layout: workflowy.LayoutTodo}, nil
	case "complete":
		return completedMatcher{}, nil
	case "bullets":
		return layoutMatcher{layout: workflowy.LayoutBullets}, nil
	case "h1":
		return layoutMatcher{layout: workflowy.LayoutH1}, nil
	case "h2":
		return layoutMatcher{layout: workflowy.LayoutH2}, nil
	case "h3":
		return layoutMatcher{layout: workflowy.LayoutH3}, nil
	case "code-block":
		return layoutMatcher{layout: workflowy.LayoutCodeBlock}, nil
	case "quote-block":
		return layoutMatcher{layout: workflowy.LayoutQuoteBlock}, nil
	default:
		return nil, fmt.Errorf("workflowy: invalid search query: is:%s is not supported by exported Workflowy data", value)
	}
}

func compileHasMatcher(value string) (searchMatcher, error) {
	switch value {
	case "note":
		return noteMatcher{}, nil
	default:
		return nil, fmt.Errorf("workflowy: invalid search query: has:%s is not supported by exported Workflowy data", value)
	}
}

var relativeAgePattern = regexp.MustCompile(`^(\d+)(m|h|d|w)$`)

func compileRecentTimestampMatcher(value string, field timeField) (searchMatcher, error) {
	age, err := parseRelativeAge(value)
	if err != nil {
		operator := "created"
		if field == timeFieldModified {
			operator = "changed"
		}
		return nil, fmt.Errorf("workflowy: invalid search query: %s", operator+":"+err.Error())
	}
	return recentTimestampMatcher{
		field: field,
		since: searchNow().Add(-age),
	}, nil
}

func parseRelativeAge(value string) (time.Duration, error) {
	match := relativeAgePattern.FindStringSubmatch(strings.ToLower(strings.TrimSpace(value)))
	if len(match) != 3 {
		return 0, fmt.Errorf("%s must use a relative age like 30m, 12h, 7d, or 2w", value)
	}

	n, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, fmt.Errorf("invalid age %q: %w", value, err)
	}

	switch match[2] {
	case "m":
		return time.Duration(n) * time.Minute, nil
	case "h":
		return time.Duration(n) * time.Hour, nil
	case "d":
		return time.Duration(n) * 24 * time.Hour, nil
	case "w":
		return time.Duration(n) * 7 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("%s must use a relative age like 30m, 12h, 7d, or 2w", value)
	}
}

func (p *searchParser) hasNext() bool {
	return p.pos < len(p.tokens)
}

func (p *searchParser) peek() searchToken {
	return p.tokens[p.pos]
}

func (p *searchParser) next() searchToken {
	token := p.tokens[p.pos]
	p.pos++
	return token
}
