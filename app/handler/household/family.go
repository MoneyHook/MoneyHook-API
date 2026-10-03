package household

import (
	h "MoneyHook/MoneyHook-API/household"
	"context"
	"github.com/labstack/echo/v4"
)

func (v *Handler) List(c echo.Context) error {
	return v.run(c, 200, func(ctx context.Context, u, f string) (any, error) { return v.store.List(ctx, u) })
}
func (v *Handler) Get(c echo.Context) error {
	return v.run(c, 200, func(ctx context.Context, u, f string) (any, error) { return v.store.Get(ctx, u, f) })
}
func (v *Handler) Create(c echo.Context) error {
	in, e := read[h.FamilyInput](c)
	if e != nil || c.Response().Committed {
		return e
	}
	return v.run(c, 201, func(ctx context.Context, u, f string) (any, error) { return v.store.Create(ctx, u, in, key(c)) })
}
func (v *Handler) Rename(c echo.Context) error {
	in, e := read[h.FamilyInput](c)
	if e != nil || c.Response().Committed {
		return e
	}
	return v.run(c, 200, func(ctx context.Context, u, f string) (any, error) { return v.store.Rename(ctx, u, f, in) })
}
func (v *Handler) Members(c echo.Context) error {
	return v.run(c, 200, func(ctx context.Context, u, f string) (any, error) { return v.store.Members(ctx, u, f) })
}
func (v *Handler) RenameMember(c echo.Context) error {
	in, e := read[h.MemberInput](c)
	if e != nil || c.Response().Committed {
		return e
	}
	return v.run(c, 204, func(ctx context.Context, u, f string) (any, error) {
		return nil, v.store.RenameMember(ctx, u, f, in.DisplayName)
	})
}
func (v *Handler) Transfer(c echo.Context) error {
	in, e := read[h.MemberInput](c)
	if e != nil || c.Response().Committed {
		return e
	}
	return v.run(c, 204, func(ctx context.Context, u, f string) (any, error) { return nil, v.store.Transfer(ctx, u, f, in) })
}
func (v *Handler) Leave(c echo.Context) error {
	in, e := read[h.Version](c)
	if e != nil || c.Response().Committed {
		return e
	}
	return v.run(c, 204, func(ctx context.Context, u, f string) (any, error) {
		return nil, v.store.Leave(ctx, u, f, "", in.ExpectedVersion, false)
	})
}
func (v *Handler) RemoveMember(c echo.Context) error {
	return v.run(c, 204, func(ctx context.Context, u, f string) (any, error) {
		return nil, v.store.Leave(ctx, u, f, c.Param("memberId"), version(c), false)
	})
}
func (v *Handler) Archive(c echo.Context) error {
	in, e := read[h.Version](c)
	if e != nil || c.Response().Committed {
		return e
	}
	return v.run(c, 204, func(ctx context.Context, u, f string) (any, error) {
		return nil, v.store.Leave(ctx, u, f, "", in.ExpectedVersion, true)
	})
}
