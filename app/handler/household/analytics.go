package household

import (
	h "MoneyHook/MoneyHook-API/household"
	"context"
	"github.com/labstack/echo/v4"
	"sort"
	"strconv"
	"time"
)

func (v *Handler) Analytics(c echo.Context) error {
	return v.run(c, 200, func(ctx context.Context, u, f string) (any, error) {
		if c.Param("group") != "overview" && c.Param("group") != "categories" && c.Param("group") != "payers" {
			return nil, h.Invalid
		}
		result := h.Summary{Groups: []h.SummaryGroup{}}
		groups := map[string]h.SummaryGroup{}
		cursor := ""
		for {
			p, e := v.store.Entries(ctx, u, f, h.Filter{Month: c.QueryParam("month"), Cursor: cursor})
			if e != nil {
				return nil, e
			}
			for _, entry := range p.Entries {
				if entry.Excluded {
					continue
				}
				if entry.Sign == 1 {
					result.Income += entry.Amount
				} else {
					result.Expense += entry.Amount
					id := entry.CategoryID
					name := entry.CategoryName
					if c.Param("group") == "payers" {
						id = "common"
						if entry.Payer.MemberID != nil {
							id = *entry.Payer.MemberID
						}
						name = entry.Payer.DisplayName
					}
					g := groups[id]
					g.ID = id
					g.Name = name
					g.Amount += entry.Amount
					groups[id] = g
				}
			}
			if p.NextCursor == nil {
				break
			}
			cursor = *p.NextCursor
		}
		result.Balance = result.Income - result.Expense
		for _, g := range groups {
			result.Groups = append(result.Groups, g)
		}
		sort.Slice(result.Groups, func(i, j int) bool { return result.Groups[i].Amount > result.Groups[j].Amount })
		return result, nil
	})
}
func (v *Handler) Duplicates(c echo.Context) error {
	return v.run(c, 200, func(ctx context.Context, u, f string) (any, error) {
		date := c.QueryParam("date")
		parsed, err := time.Parse("2006-01-02", date)
		if err != nil || parsed.Format("2006-01-02") != date {
			return nil, h.Invalid
		}
		amount, e := strconv.ParseInt(c.QueryParam("amount"), 10, 64)
		if e != nil || amount < 1 || amount > 9999999 {
			return nil, h.Invalid
		}
		sign, e := strconv.Atoi(c.QueryParam("sign"))
		if e != nil || (sign != -1 && sign != 1) {
			return nil, h.Invalid
		}
		if exclude := c.QueryParam("exclude"); exclude != "" && !h.ID(exclude) {
			return nil, h.Invalid
		}
		out := []h.Entry{}
		cursor := ""
		for {
			p, e := v.store.Entries(ctx, u, f, h.Filter{Month: date[:7] + "-01", Payer: c.QueryParam("payer"), Cursor: cursor})
			if e != nil {
				return nil, e
			}
			for _, r := range p.Entries {
				if !r.Excluded && r.Date == date && r.Amount == amount && r.Sign == sign && r.ID != c.QueryParam("exclude") {
					out = append(out, r)
				}
			}
			if p.NextCursor == nil {
				break
			}
			cursor = *p.NextCursor
		}
		return out, nil
	})
}
