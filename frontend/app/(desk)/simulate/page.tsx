"use client";

import dynamic from "next/dynamic";
import Link from "next/link";
import { useEffect, useRef, useState } from "react";
import type { MapMarker } from "@/components/CityMap";
import {
  assignOrder,
  createDriver,
  createOrder,
  demoStatus,
  getDriver,
  getDrivers,
  startDemo,
  stopDemo,
  updateOrderStatus,
} from "@/lib/api";
import { DEMO_ROUTE } from "@/lib/routes";
import type { Driver } from "@/lib/types";

const CityMap = dynamic(() => import("@/components/CityMap").then((mod) => mod.CityMap), { ssr: false });

export default function SimulatePage() {
  const [drivers, setDrivers] = useState<Driver[]>([]);
  const [driverId, setDriverId] = useState("");
  const [running, setRunning] = useState(false);
  const [sent, setSent] = useState(0);
  const [point, setPoint] = useState(DEMO_ROUTE[0]);
  const [orderId, setOrderId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  const timerRef = useRef<number | null>(null);

  function watch(id: string) {
    if (timerRef.current) window.clearInterval(timerRef.current);
    const pull = () => {
      getDriver(id)
        .then((driver) => {
          if (!driver.location) return;
          setPoint({ lat: driver.location.lat, lng: driver.location.lng });
          setSent((count) => count + 1);
        })
        .catch(() => {
          /* The server loop keeps publishing if this tab is closed. */
        });
    };
    pull();
    timerRef.current = window.setInterval(pull, 800);
  }

  useEffect(() => {
    getDrivers()
      .then((rows) => {
        setDrivers(rows);
        if (rows[0]) setDriverId((current) => current || rows[0].id);
      })
      .catch((err: unknown) => setError(err instanceof Error ? err.message : "Could not load drivers"));
    demoStatus()
      .then((status) => {
        if (!status.running || !status.driverId) return;
        setDriverId(status.driverId);
        setRunning(true);
        watch(status.driverId);
      })
      .catch(() => {
        /* Status is optional on first paint. */
      });
    return () => {
      if (timerRef.current) window.clearInterval(timerRef.current);
    };
    // watch is stable enough for mount; the interval is cleared on unmount only.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  async function stop() {
    setError(null);
    try {
      await stopDemo();
      if (timerRef.current) window.clearInterval(timerRef.current);
      timerRef.current = null;
      setRunning(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not stop");
    }
  }

  async function launch() {
    setError(null);
    setBusy(true);
    try {
      let id = driverId;
      if (!id) {
        const driver = await createDriver({
          name: "Mali",
          phone: "0812345678",
          vehicle: "Scooter",
        });
        id = driver.id;
        setDrivers((prev) => [...prev, driver]);
        setDriverId(id);
      }
      const created = await createOrder({
        customerName: "Suda",
        pickupAddress: "Siam Paragon",
        dropoffAddress: "Thong Lo",
        pickupLat: DEMO_ROUTE[0].lat,
        pickupLng: DEMO_ROUTE[0].lng,
        dropoffLat: DEMO_ROUTE[DEMO_ROUTE.length - 1].lat,
        dropoffLng: DEMO_ROUTE[DEMO_ROUTE.length - 1].lng,
      });
      await assignOrder(created.id, id);
      await updateOrderStatus(created.id, "picked_up");
      await updateOrderStatus(created.id, "in_transit");
      setOrderId(created.id);
      setPoint(DEMO_ROUTE[0]);
      await startDemo(id);
      setRunning(true);
      setSent(0);
      watch(id);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Could not start the demo");
    } finally {
      setBusy(false);
    }
  }

  const markers: MapMarker[] = [
    { id: "pickup", lat: DEMO_ROUTE[0].lat, lng: DEMO_ROUTE[0].lng, label: "Pickup", kind: "pickup" },
    {
      id: "dropoff",
      lat: DEMO_ROUTE[DEMO_ROUTE.length - 1].lat,
      lng: DEMO_ROUTE[DEMO_ROUTE.length - 1].lng,
      label: "Dropoff",
      kind: "dropoff",
    },
    {
      id: "driver",
      lat: point.lat,
      lng: point.lng,
      label: drivers.find((driver) => driver.id === driverId)?.name ?? "Mali",
      kind: "driver",
      live: running,
      detail: "Simulator",
    },
  ];

  return (
    <div className="mx-auto grid max-w-6xl gap-6 px-4 py-6 lg:grid-cols-[24rem_1fr]">
      <section className="rounded-2xl border border-line bg-panel p-5">
        <h1 className="font-serif text-4xl leading-none">Driver simulator</h1>
        <p className="mt-3 text-sm leading-relaxed text-fog">
          This page stands in for a phone. The API keeps posting GPS along a Bangkok route after
          you leave, so the live map and the tracking page keep moving. Press Stop to end it.
          Demo pace is faster than a scooter.
        </p>
        <label className="mt-5 block text-xs text-fog">
          Driver
          <select className="field mt-1" value={driverId} onChange={(e) => setDriverId(e.target.value)}>
            <option value="">Create Mali on launch</option>
            {drivers.map((driver) => (
              <option key={driver.id} value={driver.id}>
                {driver.name} · {driver.vehicle || "no vehicle"}
              </option>
            ))}
          </select>
        </label>
        <div className="mt-5 flex gap-2">
          <button type="button" className="btn-primary" onClick={launch} disabled={running || busy}>
            Start demo route
          </button>
          <button type="button" className="btn-ghost" onClick={stop} disabled={!running}>
            Stop
          </button>
        </div>
        <dl className="mt-5 grid grid-cols-2 gap-3 text-sm">
          <div className="rounded-xl bg-black/25 p-3">
            <dt className="text-xs text-fog">Pings sent</dt>
            <dd className="font-mono text-2xl">{sent}</dd>
          </div>
          <div className="rounded-xl bg-black/25 p-3">
            <dt className="text-xs text-fog">Current fix</dt>
            <dd className="font-mono text-sm">
              {point.lat.toFixed(5)}
              <br />
              {point.lng.toFixed(5)}
            </dd>
          </div>
        </dl>
        {orderId ? (
          <p className="mt-4 text-sm">
            Customer page:{" "}
            <Link href={`/track/${orderId}`} className="text-signal underline-offset-2 hover:underline">
              /track/{orderId.slice(0, 8)}…
            </Link>
          </p>
        ) : null}
        {error ? <p className="mt-3 text-sm text-lantern">{error}</p> : null}
      </section>
      <div className="h-[32rem] overflow-hidden rounded-2xl border border-line">
        <CityMap markers={markers} trails={[DEMO_ROUTE.map((stop) => [stop.lat, stop.lng])]} fit="once" />
      </div>
    </div>
  );
}
