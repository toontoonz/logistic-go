package handler

import "github.com/gofiber/fiber/v2"

type startDemoBody struct {
	DriverID string `json:"driverId"`
}

func (a API) demoStatus(c *fiber.Ctx) error {
	if a.Demo == nil {
		return writeErr(c, fiber.StatusNotFound, "demo simulator is not enabled")
	}
	return writeData(c, fiber.StatusOK, a.Demo.Status())
}

func (a API) startDemo(c *fiber.Ctx) error {
	if a.Demo == nil {
		return writeErr(c, fiber.StatusNotFound, "demo simulator is not enabled")
	}
	var body startDemoBody
	if err := c.BodyParser(&body); err != nil {
		return writeErr(c, fiber.StatusBadRequest, "invalid JSON")
	}
	if err := a.Demo.Start(body.DriverID); err != nil {
		return writeServiceErr(c, err)
	}
	return writeData(c, fiber.StatusOK, a.Demo.Status())
}

func (a API) stopDemo(c *fiber.Ctx) error {
	if a.Demo == nil {
		return writeErr(c, fiber.StatusNotFound, "demo simulator is not enabled")
	}
	a.Demo.Stop()
	return writeData(c, fiber.StatusOK, a.Demo.Status())
}
