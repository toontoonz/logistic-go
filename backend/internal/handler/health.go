package handler

import "github.com/gofiber/fiber/v2"

type healthReport struct {
	Status string `json:"status"`
}

func (a API) health(c *fiber.Ctx) error {
	ctx, cancel := requestCtx(c)
	defer cancel()
	if a.DB != nil {
		if err := a.DB.Ping(ctx); err != nil {
			return writeErr(c, fiber.StatusServiceUnavailable, "database unavailable")
		}
	}
	if a.Cache != nil {
		if err := a.Cache.Ping(ctx); err != nil {
			return writeErr(c, fiber.StatusServiceUnavailable, "redis unavailable")
		}
	}
	return writeData(c, fiber.StatusOK, healthReport{Status: "ok"})
}
