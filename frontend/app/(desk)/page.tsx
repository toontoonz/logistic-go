"use client";

import dynamic from "next/dynamic";
import { FormEvent, useEffect, useMemo, useState } from "react";
import type { MapMarker } from "@/components/CityMap";
import { ApiError, createDriver, getDrivers } from "@/lib/api";
import { ago } from "@/lib/geo";
import type { Driver } from "@/lib/types";
import { useDriverPositions } from "@/lib/use-positions";

const CityMap = dynamic(() => import("@/components/CityMap").then((mod) => mod.CityMap), {
  ssr: false,
  loading: () => <div className="grid h-full place-items-center text-sm text-fog">Loading map…</div>,
});

export default function LiveMapPage() {
  const { positions, trails, connected } = useDriverPositions();
  const [drivers, setDrivers] = useState<Driver[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [adding, setAdding] = useState(false);
  const [now, setNow] = useState(() => Date.now());

  useEffect(() => {
    const timer = window.setInterval(() => setNow(Date.now()), 1000);
    return () => window.clearInterval(timer);
  }, []);

  useEffect(() => {
    let stop = false;
    const load = () => {
      getDrivers()
        .then((rows) => {
          if (!stop) setDrivers(rows);
        })
        .catch((err: unknown) => {
          if (!stop) setError(err instanceof ApiError ? err.message : "Could not load drivers");
        });
    };
    load();
    const timer = window.setInterval(load, 3000);
    return () => {
      stop = true;
      window.clearInterval(timer);
    };
  }, []);

  const rows = useMemo<Driver[]>(() => {
    const seen = new Set(drivers.map((driver) => driver.id));
    const extras = Object.values(positions)
      .filter((point) => !seen.has(point.driverId))
      .map((point) => ({
        id: point.driverId,
        name: point.driverName || "Driver",
        phone: "",
        vehicle: "",
        createdAt: point.recordedAt,
        location: point,
      }));
    return [...drivers, ...extras];
  }, [drivers, positions]);

  const markers = useMemo<MapMarker[]>(() => {
    return rows.flatMap((driver) => {
      const live = positions[driver.id] ?? driver.location;
      if (!live) return [];
      const fresh = now - Date.parse(live.recordedAt) < 30_000;
      return [
        {
          id: driver.id,
          lat: live.lat,
          lng: live.lng,
          label: driver.name,
          kind: "driver" as const,
          live: fresh,
          detail: driver.vehicle || "Driver",
        },
      ];
    });
  }, [rows, positions, now]);

  const trailLines = useMemo(() => Object.values(trails).filter((trail) => trail.length > 1), [trails]);

  async function onCreate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const form = new FormData(event.currentTarget);
    setError(null);
    try {
      const driver = await createDriver({
        name: String(form.get("name") ?? ""),
        phone: String(form.get("phone") ?? ""),
        vehicle: String(form.get("vehicle") ?? ""),
      });
      setDrivers((prev) => [...prev, driver].sort((a, b) => a.name.localeCompare(b.name)));
      event.currentTarget.reset();
      setAdding(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not add driver");
    }
  }

  return (
    <div className="relative h-[calc(100dvh-3.5rem)]">
      <CityMap markers={markers} trails={trailLines} />
      <aside className="absolute top-4 left-4 z-[500] flex max-h-[calc(100%-2rem)] w-[min(100%-2rem,22rem)] flex-col gap-3 overflow-auto">
        <section className="rounded-2xl border border-line bg-panel/95 p-4 shadow-2xl backdrop-blur">
          <div className="flex items-center justify-between gap-3">
            <div>
              <p className="font-serif text-2xl leading-none">Drivers</p>
              <p className="mt-1 text-xs text-fog">Latest fix from Redis, trail from the live socket.</p>
            </div>
            <span
              className={`rounded-full px-2 py-1 text-xs ${connected ? "bg-signal/15 text-signal" : "bg-lantern/15 text-lantern"}`}
            >
              {connected ? "Live" : "Reconnecting"}
            </span>
          </div>
          <ul className="mt-4 space-y-2">
            {rows.length === 0 ? (
              <li className="rounded-xl border border-dashed border-line px-3 py-4 text-sm text-fog">
                No drivers yet. Add one, then open the simulator to move them.
              </li>
            ) : (
              rows.map((driver) => {
                const live = positions[driver.id] ?? driver.location;
                const fresh = live ? now - Date.parse(live.recordedAt) < 30_000 : false;
                return (
                  <li key={driver.id} className="rounded-xl bg-black/25 px-3 py-2">
                    <div className="flex items-center justify-between gap-2">
                      <p className="font-medium">{driver.name}</p>
                      <span className={`h-2 w-2 rounded-full ${fresh ? "bg-signal" : "bg-fog/50"}`} />
                    </div>
                    <p className="text-xs text-fog">
                      {[driver.vehicle, driver.phone].filter(Boolean).join(" · ") || "Profile loading"}
                    </p>
                    <p className="mt-1 font-mono text-[11px] text-fog">
                      {live
                        ? `${live.lat.toFixed(5)}, ${live.lng.toFixed(5)} · ${ago(live.recordedAt, now)}`
                        : "No GPS yet"}
                    </p>
                  </li>
                );
              })
            )}
          </ul>
          {adding ? (
            <form onSubmit={onCreate} className="mt-4 space-y-2">
              <input name="name" required placeholder="Name" className="field" />
              <input name="phone" required placeholder="Phone" className="field" />
              <input name="vehicle" placeholder="Vehicle" className="field" />
              <div className="flex gap-2">
                <button type="submit" className="btn-primary">
                  Save driver
                </button>
                <button type="button" className="btn-ghost" onClick={() => setAdding(false)}>
                  Cancel
                </button>
              </div>
            </form>
          ) : (
            <button type="button" className="btn-primary mt-4 w-full" onClick={() => setAdding(true)}>
              Add driver
            </button>
          )}
          {error ? <p className="mt-3 text-sm text-lantern">{error}</p> : null}
        </section>
      </aside>
    </div>
  );
}
