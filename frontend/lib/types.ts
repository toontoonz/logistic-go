// JSON shapes match the Go structs in backend/internal/model/model.go.

export type Position = {
  driverId: string;
  driverName?: string;
  lat: number;
  lng: number;
  recordedAt: string;
};

export type Driver = {
  id: string;
  name: string;
  phone: string;
  vehicle: string;
  createdAt: string;
  location?: Position;
};

export type OrderStatus =
  | "pending"
  | "assigned"
  | "picked_up"
  | "in_transit"
  | "delivered";

export type Order = {
  id: string;
  customerName: string;
  pickupAddress: string;
  dropoffAddress: string;
  pickupLat: number;
  pickupLng: number;
  dropoffLat: number;
  dropoffLng: number;
  status: OrderStatus;
  driverId?: string;
  driverName?: string;
  createdAt: string;
  updatedAt: string;
};

export type OrderDetail = Order & {
  driverLocation?: Position;
};

export const STATUS_LABEL: Record<OrderStatus, string> = {
  pending: "Pending",
  assigned: "Assigned",
  picked_up: "Picked up",
  in_transit: "In transit",
  delivered: "Delivered",
};

export const STATUS_STEPS: OrderStatus[] = [
  "pending",
  "assigned",
  "picked_up",
  "in_transit",
  "delivered",
];

export const NEXT_STATUS: Partial<Record<OrderStatus, { status: OrderStatus; label: string }>> = {
  assigned: { status: "picked_up", label: "Mark picked up" },
  picked_up: { status: "in_transit", label: "Start transit" },
  in_transit: { status: "delivered", label: "Mark delivered" },
};
