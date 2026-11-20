package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/toontoonz/logistic-go/backend/internal/service"
)

type createDriverBody struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Vehicle string `json:"vehicle"`
}

type locationBody struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func (a API) createDriver(c *fiber.Ctx) error {
	var body createDriverBody
	if err := c.BodyParser(&body); err != nil {
		return writeErr(c, fiber.StatusBadRequest, "invalid JSON")
	}
	ctx, cancel := requestCtx(c)
	defer cancel()
	driver, err := a.Drivers.Create(ctx, service.CreateDriverInput{
		Name:    body.Name,
		Phone:   body.Phone,
		Vehicle: body.Vehicle,
	})
	if err != nil {
		return writeServiceErr(c, err)
	}
	return writeData(c, fiber.StatusCreated, driver)
}

func (a API) listDrivers(c *fiber.Ctx) error {
	ctx, cancel := requestCtx(c)
	defer cancel()
	drivers, err := a.Drivers.List(ctx)
	if err != nil {
		return writeServiceErr(c, err)
	}
	return writeData(c, fiber.StatusOK, drivers)
}

func (a API) getDriver(c *fiber.Ctx) error {
	ctx, cancel := requestCtx(c)
	defer cancel()
	driver, err := a.Drivers.Get(ctx, c.Params("id"))
	if err != nil {
		return writeServiceErr(c, err)
	}
	return writeData(c, fiber.StatusOK, driver)
}

func (a API) postLocation(c *fiber.Ctx) error {
	var body locationBody
	if err := c.BodyParser(&body); err != nil {
		return writeErr(c, fiber.StatusBadRequest, "invalid JSON")
	}
	ctx, cancel := requestCtx(c)
	defer cancel()
	pos, err := a.Locations.Ingest(ctx, c.Params("id"), body.Lat, body.Lng)
	if err != nil {
		return writeServiceErr(c, err)
	}
	return writeData(c, fiber.StatusCreated, pos)
}

func (a API) locationHistory(c *fiber.Ctx) error {
	ctx, cancel := requestCtx(c)
	defer cancel()
	limit := c.QueryInt("limit", 100)
	points, err := a.Locations.History(ctx, c.Params("id"), limit)
	if err != nil {
		return writeServiceErr(c, err)
	}
	return writeData(c, fiber.StatusOK, points)
}
