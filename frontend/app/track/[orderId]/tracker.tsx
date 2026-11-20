"use client";

import dynamic from "next/dynamic";
import Link from "next/link";
import { useEffect, useState } from "react";
import type { MapMarker } from "@/components/CityMap";
import { ApiError, getHistory, getOrder } from "@/lib/api";
import { ago } from "@/lib/geo";
import { STATUS_LABEL, STATUS_STEPS, type OrderDetail, type Position } from "@/lib/types";
import { useDriverPositions } from "@/lib/use-positions";

const CityMap = dynamic(() => import("@/components/CityMap").then((mod) => mod.CityMap), {
  ssr: false,
  loading: () => <div className="grid h-full place-items-center text-sm text-fog">Loading map…</div>,
});

function blurb(order: OrderDetail): string {
  const name = order.driverName ?? "Your driver";
  switch (order.status) {
    case "pending":
      return "Waiting for a driver to be assigned.";
    case "assigned":
      return `${name} has the job and is heading to the pickup.`;
    case "picked_up":
      return `The parcel is with ${name}.`;
    case "in_transit":
      return `${name} is on the way to the dropoff.`;
    case "delivered":
      return "Delivered. Thank you for using Lane.";
    default:
      return "";
  }
}

export function Tracker({ orderId }: { orderId: string }) {
  const [order, setOrder] = useState<OrderDetail | null>(null);
  const [missing, setMissing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [history, setHistory] = useState<Position[]>([]);
  const [now, setNow] = useState(() => Date.now());
  const { positions } = useDriverPositions();

  useEffect(() => {
    const timer = window.setInterval(() => setNow(Date.now()), 1000);
    return () => window.clearInterval(timer);
  }, []);

  useEffect(() => {
    let stop = false;
    const load = () => {
      getOrder(orderId)
        .then((detail) => {
          if (stop) return;
          setOrder(detail);
          setMissing(false);
        })
        .catch((err: unknown) => {
          if (stop) return;
          if (err instanceof ApiError && err.status === 404) {
            setMissing(true);
            return;
          }
          setError(err instanceof Error ? err.message : "Could not load this parcel");
        });
    };
    load();
    const timer = window.setInterval(load, 4000);
    return () => {
      stop = true;
      window.clearInterval(timer);
    };
  }, [orderId]);

  const driverId = order?.driverId;

  useEffect(() => {
    if (!driverId) return;
    let stop = false;
    const load = () => {
      getHistory(driverId)
        .then((points) => {
          if (!stop) setHistory(points);
        })
        .catch(() => {
          /* The live socket still moves the marker. */
        });
    };
    load();
    const timer = window.setInterval(load, 4000);
    return () => {
      stop = true;
      window.clearInterval(timer);
    };
  }, [driverId]);

  const live = driverId ? positions[driverId] : undefined;
  const fallback = history[history.length - 1] ?? order?.driverLocation;
  const current = live ?? fallback;

  const markers: MapMarker[] = [];
  if (order) {
    markers.push({
      id: "pickup",
      lat: order.pickupLat,
      lng: order.pickupLng,
      label: "Pickup",
      kind: "pickup",
      detail: order.pickupAddress,
    });
    markers.push({
      id: "dropoff",
      lat: order.dropoffLat,
      lng: order.dropoffLng,
      label: "Dropoff",
      kind: "dropoff",
      detail: order.dropoffAddress,
    });
  }
  if (current) {
    markers.push({
      id: "driver",
      lat: current.lat,
      lng: current.lng,
      label: order?.driverName ?? current.driverName ?? "Driver",
      kind: "driver",
      live: true,
    });
  }

  const trail = history.map((point) => [point.lat, point.lng] as [number, number]);
  if (current) {
    const last = trail[trail.length - 1];
    if (!last || last[0] !== current.lat || last[1] !== current.lng) {
      trail.push([current.lat, current.lng]);
    }
  }

  const step = order ? STATUS_STEPS.indexOf(order.status) : -1;

  return (
    <div className="min-h-dvh bg-paper text-ink">
      <header className="mx-auto flex max-w-5xl items-baseline justify-between px-4 py-5">
        <Link href="/" className="font-serif text-3xl text-ink">
          Lane
        </Link>
        <p className="text-xs uppercase tracking-[0.18em] text-ink/50">Parcel tracking</p>
      </header>
      <main className="mx-auto grid max-w-5xl gap-6 px-4 pb-12 lg:grid-cols-[1fr_1.1fr]">
        {missing ? (
          <section className="rounded-3xl bg-foam p-8 shadow-sm lg:col-span-2">
            <h1 className="font-serif text-4xl">We cannot find that parcel</h1>
            <p className="mt-2 text-ink/70">Check the link, or ask dispatch for a new one.</p>
          </section>
        ) : null}
        {!order && !missing ? (
          <p className="text-sm text-ink/60">{error ?? "Loading parcel…"}</p>
        ) : null}
        {order ? (
          <>
            <section className="rounded-3xl bg-foam p-6 shadow-sm">
              <p className="text-xs uppercase tracking-[0.16em] text-ink/45">Waybill {order.id.slice(0, 8)}</p>
              <h1 className="mt-2 font-serif text-4xl">{order.customerName}</h1>
              <p className="mt-3 text-lg text-ink/80">{blurb(order)}</p>
              <ol className="mt-6 space-y-2">
                {STATUS_STEPS.map((status, index) => {
                  const done = index <= step;
                  return (
                    <li key={status} className="flex items-center gap-3 text-sm">
                      <span className={`h-2.5 w-2.5 rounded-full ${done ? "bg-ink" : "bg-ink/15"}`} />
                      <span className={done ? "font-medium" : "text-ink/40"}>{STATUS_LABEL[status]}</span>
                    </li>
                  );
                })}
              </ol>
              <div className="mt-6 space-y-3 text-sm">
                <p>
                  <span className="block text-xs uppercase tracking-wide text-ink/45">Pickup</span>
                  {order.pickupAddress}
                </p>
                <p>
                  <span className="block text-xs uppercase tracking-wide text-ink/45">Dropoff</span>
                  {order.dropoffAddress}
                </p>
                <p>
                  <span className="block text-xs uppercase tracking-wide text-ink/45">Driver</span>
                  {order.driverName ?? "Not assigned yet"}
                  {current ? (
                    <span className="mt-1 block font-mono text-xs text-ink/60">
                      {current.lat.toFixed(5)}, {current.lng.toFixed(5)} · {ago(current.recordedAt, now)}
                    </span>
                  ) : (
                    <span className="mt-1 block text-ink/60">No GPS ping yet.</span>
                  )}
                </p>
              </div>
              {error ? <p className="mt-4 text-sm text-[#8a4b00]">{error}</p> : null}
            </section>
            <div className="h-[28rem] overflow-hidden rounded-3xl shadow-sm">
              <CityMap markers={markers} trails={trail.length > 1 ? [trail] : []} fit="once" />
            </div>
          </>
        ) : null}
      </main>
    </div>
  );
}
