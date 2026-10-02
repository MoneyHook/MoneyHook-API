package household

import (
	h "MoneyHook/MoneyHook-API/household"
	"context"
	"github.com/labstack/echo/v4"
)

func (v *Handler) Invitations(c echo.Context) error {
	return v.run(c, 200, func(ctx context.Context, u, f string) (any, error) { return v.store.Invitations(ctx, u, f) })
}
func (v *Handler) Issue(c echo.Context) error {
	in, e := read[struct{}](c)
	if e != nil || c.Response().Committed {
		return e
	}
	_ = in
	return v.run(c, 201, func(ctx context.Context, u, f string) (any, error) { return v.store.Issue(ctx, u, f, "", key(c)) })
}
func (v *Handler) Reissue(c echo.Context) error {
	in, e := read[struct{}](c)
	if e != nil || c.Response().Committed {
		return e
	}
	_ = in
	return v.run(c, 201, func(ctx context.Context, u, f string) (any, error) {
		return v.store.Issue(ctx, u, f, c.Param("invitationId"), key(c))
	})
}
func (v *Handler) Revoke(c echo.Context) error {
	return v.run(c, 204, func(ctx context.Context, u, f string) (any, error) {
		return nil, v.store.Revoke(ctx, u, f, c.Param("invitationId"))
	})
}
func (v *Handler) Preview(c echo.Context) error {
	in, e := read[h.Credential](c)
	if e != nil || c.Response().Committed {
		return e
	}
	return v.run(c, 200, func(ctx context.Context, u, f string) (any, error) { return v.store.Preview(ctx, u, peer(c), in) })
}
func (v *Handler) Accept(c echo.Context) error {
	in, e := read[h.Credential](c)
	if e != nil || c.Response().Committed {
		return e
	}
	return v.run(c, 201, func(ctx context.Context, u, f string) (any, error) {
		return v.store.Accept(ctx, u, peer(c), in, key(c))
	})
}
