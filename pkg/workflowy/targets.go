package workflowy

import "context"

// TargetsService handles operations on WorkFlowy targets.
type TargetsService struct {
	s *service
}

// List returns a call that lists all available targets.
func (ts *TargetsService) List() *TargetsListCall {
	return &TargetsListCall{s: ts.s}
}

// TargetsListCall lists all targets.
type TargetsListCall struct {
	s *service
}

// Do executes the list targets call.
func (c *TargetsListCall) Do(ctx context.Context) ([]*Target, error) {
	req, err := c.s.newRequest("GET", "/targets", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Targets []*Target `json:"targets"`
	}
	if err := c.s.do(ctx, req, &resp); err != nil {
		return nil, err
	}
	return resp.Targets, nil
}
