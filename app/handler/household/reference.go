package household

import (
	h "MoneyHook/MoneyHook-API/household"
	"context"
	"github.com/labstack/echo/v4"
)

func (v *Handler) Payments(c echo.Context) error {
	return v.run(c, 200, func(ctx context.Context, u, f string) (any, error) { return v.store.References(ctx, u, f, "payments") })
}
func (v *Handler) SavePayments(c echo.Context) error {
	in, e := read[h.ReferenceInput](c)
	if e != nil || c.Response().Committed {
		return e
	}
	return v.run(c, 200, func(ctx context.Context, u, f string) (any, error) {
		return v.store.SaveReference(ctx, u, f, "payments", c.Param("referenceId"), in)
	})
}
func (v *Handler) Subcategories(c echo.Context) error {
	return v.run(c, 200, func(ctx context.Context, u, f string) (any, error) {
		return v.store.References(ctx, u, f, "subcategories")
	})
}
func (v *Handler) SaveSubcategories(c echo.Context) error {
	in, e := read[h.ReferenceInput](c)
	if e != nil || c.Response().Committed {
		return e
	}
	return v.run(c, 200, func(ctx context.Context, u, f string) (any, error) {
		return v.store.SaveReference(ctx, u, f, "subcategories", c.Param("referenceId"), in)
	})
}
