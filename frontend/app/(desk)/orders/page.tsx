"use client";

import dynamic from "next/dynamic";
import Link from "next/link";
import { FormEvent, useEffect, useState } from "react";
import type { MapMarker } from "@/components/CityMap";
import { assignOrder, createOrder, getDrivers, getOrders, updateOrderStatus } from "@/lib/api";
import { PRESETS } from "@/lib/routes";
import { NEXT_STATUS, STATUS_LABEL, type Driver, type Order } from "@/lib/types";

const CityMap = dynamic(() => import("@/components/CityMap").then((mod) => mod.CityMap), { ssr: false });

type Draft = {
  customerName: string;
  pickupAddress: string;
  dropoffAddress: string;
  pickupLat: string;
  pickupLng: string;
  dropoffLat: string;
  dropoffLng: string;
};

const EMPTY: Draft = {
  customerName: "",
  pickupAddress: "",
  dropoffAddress: "",
  pickupLat: "",
  pickupLng: "",
  dropoffLat: "",
  dropoffLng: "",
};

export default function OrdersPage() {
  const [orders, setOrders] = useState<Order[]>([]);
  const [drivers, setDrivers] = useState<Driver[]>([]);
  const [draft, setDraft] = useState<Draft>(EMPTY);
  const [choice, setChoice] = useState<Record<string, string>>({});
  const [error, setError] = useState<string | null>(null);
  const [notice, setNotice] = useState<string | null>(null);

  async function reload() {
    const [nextOrders, nextDrivers] = await Promise.all([getOrders(), getDrivers()]);
    setOrders(nextOrders);
    setDrivers(nextDrivers);
  }

  useEffect(() => {
    reload().catch((err: unknown) => setError(err instanceof Error ? err.message : "Could not load orders"));
  }, []);

  function setField(key: keyof Draft, value: string) {
    setDraft((prev) => ({ ...prev, [key]: value }));
  }

  async function onCreate(event: FormEvent) {
    event.preventDefault();
    setError(null);
    setNotice(null);
    try {
      const order = await createOrder({
        customerName: draft.customerName,
        pickupAddress: draft.pickupAddress,
        dropoffAddress: draft.dropoffAddress,
        pickupLat: Number(draft.pickupLat),
        pickupLng: Number(draft.pickupLng),
        dropoffLat: Number(draft.dropoffLat),
        dropoffLng: Number(draft.dropoffLng),
      });
      setDraft(EMPTY);
      setNotice(`Created. Tracking link: /track/${order.id}`);
      await reload();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not create order");
    }
  }

  async function onAssign(orderId: string) {
    const driverId = choice[orderId];
    if (!driverId) {
      setError("Pick a driver first");
      return;
    }
    setError(null);
    try {
      await assignOrder(orderId, driverId);
      await reload();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not assign");
    }
  }

  async function onAdvance(order: Order) {
    const next = NEXT_STATUS[order.status];
    if (!next) return;
    setError(null);
    try {
      await updateOrderStatus(order.id, next.status);
      await reload();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not update status");
    }
  }

  const preview: MapMarker[] = [];
  const pickupLat = Number(draft.pickupLat);
  const pickupLng = Number(draft.pickupLng);
  const dropoffLat = Number(draft.dropoffLat);
  const dropoffLng = Number(draft.dropoffLng);
  if (draft.pickupLat.trim() !== "" && draft.pickupLng.trim() !== "" && Number.isFinite(pickupLat) && Number.isFinite(pickupLng)) {
    preview.push({
      id: "pickup",
      lat: pickupLat,
      lng: pickupLng,
      label: "Pickup",
      kind: "pickup",
    });
  }
  if (draft.dropoffLat.trim() !== "" && draft.dropoffLng.trim() !== "" && Number.isFinite(dropoffLat) && Number.isFinite(dropoffLng)) {
    preview.push({
      id: "dropoff",
      lat: dropoffLat,
      lng: dropoffLng,
      label: "Dropoff",
      kind: "dropoff",
    });
  }

  return (
    <div className="mx-auto grid max-w-6xl gap-6 px-4 py-6 lg:grid-cols-[22rem_1fr]">
      <form onSubmit={onCreate} className="h-fit rounded-2xl border border-line bg-panel p-4">
        <h1 className="font-serif text-3xl">New parcel</h1>
        <p className="mt-1 text-sm text-fog">Coordinates are WGS84. Presets fill a real Bangkok pair.</p>
        <div className="mt-3 flex flex-wrap gap-2">
          {PRESETS.map((preset) => (
            <button
              key={preset.label}
              type="button"
              className="rounded-full border border-line px-3 py-1 text-xs text-fog hover:border-signal hover:text-signal"
              onClick={() =>
                setDraft({
                  customerName: preset.customerName,
                  pickupAddress: preset.pickupAddress,
                  dropoffAddress: preset.dropoffAddress,
                  pickupLat: String(preset.pickupLat),
                  pickupLng: String(preset.pickupLng),
                  dropoffLat: String(preset.dropoffLat),
                  dropoffLng: String(preset.dropoffLng),
                })
              }
            >
              {preset.label}
            </button>
          ))}
        </div>
        <div className="mt-4 space-y-2">
          <label className="block text-xs text-fog">
            Customer
            <input className="field mt-1" value={draft.customerName} onChange={(e) => setField("customerName", e.target.value)} required />
          </label>
          <label className="block text-xs text-fog">
            Pickup address
            <input className="field mt-1" value={draft.pickupAddress} onChange={(e) => setField("pickupAddress", e.target.value)} required />
          </label>
          <div className="grid grid-cols-2 gap-2">
            <input className="field" inputMode="decimal" placeholder="Pickup lat" value={draft.pickupLat} onChange={(e) => setField("pickupLat", e.target.value)} required />
            <input className="field" inputMode="decimal" placeholder="Pickup lng" value={draft.pickupLng} onChange={(e) => setField("pickupLng", e.target.value)} required />
          </div>
          <label className="block text-xs text-fog">
            Dropoff address
            <input className="field mt-1" value={draft.dropoffAddress} onChange={(e) => setField("dropoffAddress", e.target.value)} required />
          </label>
          <div className="grid grid-cols-2 gap-2">
            <input className="field" inputMode="decimal" placeholder="Dropoff lat" value={draft.dropoffLat} onChange={(e) => setField("dropoffLat", e.target.value)} required />
            <input className="field" inputMode="decimal" placeholder="Dropoff lng" value={draft.dropoffLng} onChange={(e) => setField("dropoffLng", e.target.value)} required />
          </div>
        </div>
        {preview.length > 0 ? (
          <div className="mt-3 h-40 overflow-hidden rounded-xl">
            <CityMap
              key={`${draft.pickupLat},${draft.pickupLng},${draft.dropoffLat},${draft.dropoffLng}`}
              markers={preview}
              fit="once"
              zoom={12}
            />
          </div>
        ) : null}
        <button type="submit" className="btn-primary mt-4 w-full">
          Create order
        </button>
        {notice ? <p className="mt-3 text-sm text-signal">{notice}</p> : null}
        {error ? <p className="mt-3 text-sm text-lantern">{error}</p> : null}
      </form>

      <section className="space-y-3">
        <div className="flex items-end justify-between">
          <h2 className="font-serif text-3xl">Board</h2>
          <button type="button" className="btn-ghost" onClick={() => reload().catch((err: unknown) => setError(err instanceof Error ? err.message : "Could not refresh"))}>
            Refresh
          </button>
        </div>
        {orders.length === 0 ? (
          <p className="rounded-2xl border border-dashed border-line px-4 py-8 text-sm text-fog">No parcels yet.</p>
        ) : (
          orders.map((order) => {
            const next = NEXT_STATUS[order.status];
            const canAssign = order.status === "pending" || order.status === "assigned";
            return (
              <article key={order.id} className="rounded-2xl border border-line bg-panel p-4">
                <div className="flex flex-wrap items-start justify-between gap-3">
                  <div>
                    <p className="font-medium">{order.customerName}</p>
                    <p className="text-sm text-fog">
                      {order.pickupAddress} → {order.dropoffAddress}
                    </p>
                    <p className="mt-1 text-xs text-fog">
                      {order.driverName ? `Driver ${order.driverName}` : "Unassigned"}
                    </p>
                  </div>
                  <span className="rounded-full bg-black/30 px-2 py-1 text-xs text-signal">{STATUS_LABEL[order.status]}</span>
                </div>
                <div className="mt-3 flex flex-wrap items-center gap-2">
                  {canAssign ? (
                    <>
                      <select
                        className="field max-w-48"
                        value={choice[order.id] ?? order.driverId ?? ""}
                        onChange={(e) => setChoice((prev) => ({ ...prev, [order.id]: e.target.value }))}
                      >
                        <option value="">Choose driver</option>
                        {drivers.map((driver) => (
                          <option key={driver.id} value={driver.id}>
                            {driver.name}
                          </option>
                        ))}
                      </select>
                      <button type="button" className="btn-ghost" onClick={() => onAssign(order.id)}>
                        Assign
                      </button>
                    </>
                  ) : null}
                  {next ? (
                    <button type="button" className="btn-primary" onClick={() => onAdvance(order)}>
                      {next.label}
                    </button>
                  ) : null}
                  <Link href={`/track/${order.id}`} className="text-sm text-signal underline-offset-2 hover:underline">
                    Customer tracking
                  </Link>
                </div>
              </article>
            );
          })
        )}
      </section>
    </div>
  );
}
