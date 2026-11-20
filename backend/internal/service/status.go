package service

// Order statuses. The database CHECK constraint in
// migrations/001_init.sql lists the same values.
const (
	StatusPending   = "pending"
	StatusAssigned  = "assigned"
	StatusPickedUp  = "picked_up"
	StatusInTransit = "in_transit"
	StatusDelivered = "delivered"
)

// nextStatus is the only legal step forward. Assignment is not in this
// map: it goes through OrderService.Assign, which also stores a driver.
var nextStatus = map[string]string{
	StatusAssigned:  StatusPickedUp,
	StatusPickedUp:  StatusInTransit,
	StatusInTransit: StatusDelivered,
}

// CanTransition reports whether an order may move from -> to.
func CanTransition(from, to string) bool {
	next, ok := nextStatus[from]
	return ok && next == to
}
