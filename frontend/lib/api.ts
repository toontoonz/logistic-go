import type { Driver, Order, OrderDetail, OrderStatus, Position } from "./types";

export const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";
export const WS_URL = process.env.NEXT_PUBLIC_WS_URL ?? "ws://localhost:8080/ws/positions";

export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

export async function api<T>(path: string, init?: RequestInit): Promise<T> {
  let res: Response;
  try {
    res = await fetch(`${API_URL}${path}`, {
      ...init,
      headers: {
        "Content-Type": "application/json",
        ...(init?.headers ?? {}),
      },
      cache: "no-store",
    });
  } catch {
    throw new ApiError(0, `Cannot reach the API at ${API_URL}. Is docker compose up?`);
  }

  const body: { data?: T; error?: string } = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new ApiError(res.status, body.error ?? res.statusText);
  }
  return body.data as T;
}

export const getDrivers = () => api<Driver[]>("/api/drivers");
export const getDriver = (id: string) => api<Driver>(`/api/drivers/${id}`);

export type DemoStatus = { running: boolean; driverId?: string };

export const demoStatus = () => api<DemoStatus>("/api/demo");

export function startDemo(driverId: string) {
  return api<DemoStatus>("/api/demo/start", {
    method: "POST",
    body: JSON.stringify({ driverId }),
  });
}

export function stopDemo() {
  return api<DemoStatus>("/api/demo/stop", { method: "POST", body: "{}" });
}
export const getOrders = () => api<Order[]>("/api/orders");
export const getOrder = (id: string) => api<OrderDetail>(`/api/orders/${id}`);
export const getHistory = (driverId: string) =>
  api<Position[]>(`/api/drivers/${driverId}/history?limit=300`);

export function createDriver(input: { name: string; phone: string; vehicle: string }) {
  return api<Driver>("/api/drivers", { method: "POST", body: JSON.stringify(input) });
}

export function createOrder(input: {
  customerName: string;
  pickupAddress: string;
  dropoffAddress: string;
  pickupLat: number;
  pickupLng: number;
  dropoffLat: number;
  dropoffLng: number;
}) {
  return api<Order>("/api/orders", { method: "POST", body: JSON.stringify(input) });
}

export function assignOrder(orderId: string, driverId: string) {
  return api<Order>(`/api/orders/${orderId}/assign`, {
    method: "POST",
    body: JSON.stringify({ driverId }),
  });
}

export function updateOrderStatus(orderId: string, status: OrderStatus) {
  return api<Order>(`/api/orders/${orderId}/status`, {
    method: "PATCH",
    body: JSON.stringify({ status }),
  });
}

export function postLocation(driverId: string, lat: number, lng: number) {
  return api<Position>(`/api/drivers/${driverId}/location`, {
    method: "POST",
    body: JSON.stringify({ lat, lng }),
  });
}
