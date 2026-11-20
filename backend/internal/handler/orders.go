package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/toontoonz/logistic-go/backend/internal/service"
)

type createOrderBody struct {
	CustomerName   string  `json:"customerName"`
	PickupAddress  string  `json:"pickupAddress"`
	DropoffAddress string  `json:"dropoffAddress"`
	PickupLat      float64 `json:"pickupLat"`
	PickupLng      float64 `json:"pickupLng"`
	DropoffLat     float64 `json:"dropoffLat"`
	DropoffLng     float64 `json:"dropoffLng"`
}

type assignBody struct {
	DriverID string `json:"driverId"`
}

type statusBody struct {
	Status string `json:"status"`
}

func (a API) createOrder(c *fiber.Ctx) error {
	var body createOrderBody
	if err := c.BodyParser(&body); err != nil {
		return writeErr(c, fiber.StatusBadRequest, "invalid JSON")
	}
	ctx, cancel := requestCtx(c)
	defer cancel()
	order, err := a.Orders.Create(ctx, service.CreateOrderInput{
		CustomerName:   body.CustomerName,
		PickupAddress:  body.PickupAddress,
		DropoffAddress: body.DropoffAddress,
		PickupLat:      body.PickupLat,
		PickupLng:      body.PickupLng,
		DropoffLat:     body.DropoffLat,
		DropoffLng:     body.DropoffLng,
	})
	if err != nil {
		return writeServiceErr(c, err)
	}
	return writeData(c, fiber.StatusCreated, order)
}

func (a API) listOrders(c *fiber.Ctx) error {
	ctx, cancel := requestCtx(c)
	defer cancel()
	orders, err := a.Orders.List(ctx)
	if err != nil {
		return writeServiceErr(c, err)
	}
	return writeData(c, fiber.StatusOK, orders)
}

func (a API) getOrder(c *fiber.Ctx) error {
	ctx, cancel := requestCtx(c)
	defer cancel()
	order, err := a.Orders.Get(ctx, c.Params("id"))
	if err != nil {
		return writeServiceErr(c, err)
	}
	return writeData(c, fiber.StatusOK, order)
}

func (a API) assignOrder(c *fiber.Ctx) error {
	var body assignBody
	if err := c.BodyParser(&body); err != nil {
		return writeErr(c, fiber.StatusBadRequest, "invalid JSON")
	}
	ctx, cancel := requestCtx(c)
	defer cancel()
	order, err := a.Orders.Assign(ctx, c.Params("id"), body.DriverID)
	if err != nil {
		return writeServiceErr(c, err)
	}
	return writeData(c, fiber.StatusOK, order)
}

func (a API) updateOrderStatus(c *fiber.Ctx) error {
	var body statusBody
	if err := c.BodyParser(&body); err != nil {
		return writeErr(c, fiber.StatusBadRequest, "invalid JSON")
	}
	ctx, cancel := requestCtx(c)
	defer cancel()
	order, err := a.Orders.UpdateStatus(ctx, c.Params("id"), body.Status)
	if err != nil {
		return writeServiceErr(c, err)
	}
	return writeData(c, fiber.StatusOK, order)
}
