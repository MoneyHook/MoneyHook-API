package household

import (
	h "MoneyHook/MoneyHook-API/household"
	"context"
	"github.com/labstack/echo/v4"
)

func (v *Handler) Entries(c echo.Context) error {
	return v.run(c, 200, func(ctx context.Context, u, f string) (any, error) {
		return v.store.Entries(ctx, u, f, h.Filter{Month: c.QueryParam("month"), Payer: c.QueryParam("payer"), Kind: c.QueryParam("kind"), Cursor: c.QueryParam("cursor")})
	})
}
func (v *Handler) Entry(c echo.Context) error {
	return v.run(c, 200, func(ctx context.Context, u, f string) (any, error) {
		return v.store.Entry(ctx, u, f, c.Param("entryId"))
	})
}
func (v *Handler) Own(c echo.Context) error {
	in, e := read[h.EntryInput](c)
	if e != nil || c.Response().Committed {
		return e
	}
	return v.run(c, 201, func(ctx context.Context, u, f string) (any, error) {
		return v.store.CreateEntry(ctx, u, f, in, true, key(c))
	})
}
func (v *Handler) Proxy(c echo.Context) error {
	in, e := read[h.EntryInput](c)
	if e != nil || c.Response().Committed {
		return e
	}
	return v.run(c, 201, func(ctx context.Context, u, f string) (any, error) {
		return v.store.CreateEntry(ctx, u, f, in, false, key(c))
	})
}
func (v *Handler) Share(c echo.Context) error {
	in, e := read[h.EntryInput](c)
	if e != nil || c.Response().Committed {
		return e
	}
	return v.run(c, 200, func(ctx context.Context, u, f string) (any, error) {
		return v.store.Share(ctx, u, f, c.Param("transactionId"), in)
	})
}
func (v *Handler) Unshare(c echo.Context) error {
	return v.run(c, 204, func(ctx context.Context, u, f string) (any, error) {
		return nil, v.store.Unshare(ctx, u, f, c.Param("transactionId"), version(c))
	})
}
func (v *Handler) UpdateProxy(c echo.Context) error {
	in, e := read[h.EntryInput](c)
	if e != nil || c.Response().Committed {
		return e
	}
	return v.run(c, 200, func(ctx context.Context, u, f string) (any, error) {
		return v.store.UpdateProxy(ctx, u, f, c.Param("entryId"), in)
	})
}
func (v *Handler) DeleteProxy(c echo.Context) error {
	return v.run(c, 204, func(ctx context.Context, u, f string) (any, error) {
		return nil, v.store.DeleteProxy(ctx, u, f, c.Param("entryId"), version(c))
	})
}
func (v *Handler) Correct(c echo.Context) error {
	in, e := read[h.EntryInput](c)
	if e != nil || c.Response().Committed {
		return e
	}
	return v.run(c, 201, func(ctx context.Context, u, f string) (any, error) {
		return v.store.Correct(ctx, u, f, c.Param("entryId"), in)
	})
}

func (v *Handler) ShareStatus(c echo.Context) error {
	return v.run(c, 200, func(ctx context.Context, u, f string) (any, error) {
		return v.store.ShareStatus(ctx, u, f, c.Param("transactionId"))
	})
}
