"use client";

import { useEffect, useState } from "react";
import { WS_URL } from "./api";
import type { Position } from "./types";

type SocketMessage = {
  type?: string;
  positions?: Position[];
  driverId?: string;
  driverName?: string;
  lat?: number;
  lng?: number;
  recordedAt?: string;
};

function isPosition(value: SocketMessage): value is Position {
  return (
    typeof value.driverId === "string" &&
    typeof value.lat === "number" &&
    typeof value.lng === "number" &&
    typeof value.recordedAt === "string"
  );
}

function mergePosition(prev: Record<string, Position>, next: Position): Record<string, Position> {
  const current = prev[next.driverId];
  if (current && Date.parse(current.recordedAt) > Date.parse(next.recordedAt)) {
    return prev;
  }
  return { ...prev, [next.driverId]: next };
}

export function useDriverPositions() {
  const [positions, setPositions] = useState<Record<string, Position>>({});
  const [trails, setTrails] = useState<Record<string, [number, number][]>>({});
  const [connected, setConnected] = useState(false);

  useEffect(() => {
    let socket: WebSocket | null = null;
    let stopped = false;
    let timer = 0;

    const remember = (point: Position) => {
      setPositions((prev) => mergePosition(prev, point));
      setTrails((prev) => {
        const trail = prev[point.driverId] ?? [];
        const last = trail[trail.length - 1];
        if (last && last[0] === point.lat && last[1] === point.lng) return prev;
        const nextPoint: [number, number] = [point.lat, point.lng];
        const nextTrail: [number, number][] = [...trail, nextPoint].slice(-80);
        return { ...prev, [point.driverId]: nextTrail };
      });
    };

    const connect = () => {
      socket = new WebSocket(WS_URL);
      socket.onopen = () => setConnected(true);
      socket.onclose = () => {
        setConnected(false);
        if (!stopped) timer = window.setTimeout(connect, 1000);
      };
      socket.onmessage = (event) => {
        let message: SocketMessage;
        try {
          message = JSON.parse(String(event.data)) as SocketMessage;
        } catch {
          return;
        }
        if (message.type === "snapshot" && Array.isArray(message.positions)) {
          for (const point of message.positions) remember(point);
          return;
        }
        if (message.type === "position" && isPosition(message)) {
          remember(message);
        }
      };
    };

    connect();
    return () => {
      stopped = true;
      window.clearTimeout(timer);
      socket?.close();
    };
  }, []);

  return { positions, trails, connected };
}
