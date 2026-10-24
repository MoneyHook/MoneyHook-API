package household

import (
	h "MoneyHook/MoneyHook-API/household"
	"context"
	"github.com/labstack/echo/v4"
	"strconv"
	"time"
)

func (v *Handler) Analytics(c echo.Context) error {
	return v.run(c, 200, func(ctx context.Context, u, f string) (any, error) {
		if c.Param("group") != "overview" && c.Param("group") != "categories" && c.Param("group") != "payers" {
			return nil, h.Invalid
		}
		return v.store.Analytics(ctx, u, f, c.QueryParam("month"), c.Param("group"))
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
		return v.store.Duplicates(ctx, u, f, h.DuplicateFilter{Date: date, Amount: amount, Sign: sign, Payer: c.QueryParam("payer"), Exclude: c.QueryParam("exclude")})
	})
}
