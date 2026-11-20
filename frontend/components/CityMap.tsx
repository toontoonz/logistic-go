"use client";

import { useEffect, useRef } from "react";
import { MapContainer, Marker, Polyline, Popup, TileLayer, useMap, ZoomControl } from "react-leaflet";
import L from "leaflet";
import "leaflet/dist/leaflet.css";
import { MAP_CENTER } from "@/lib/routes";

export type MarkerKind = "driver" | "pickup" | "dropoff";

export type MapMarker = {
  id: string;
  lat: number;
  lng: number;
  label: string;
  kind: MarkerKind;
  live?: boolean;
  detail?: string;
};

function iconFor(marker: MapMarker): L.DivIcon {
  const letter = marker.label.trim().charAt(0).toUpperCase() || "•";
  const live = marker.live ? " pin-live" : "";
  return L.divIcon({
    className: "lane-pin",
    html: `<div class="pin pin-${marker.kind}${live}"><span>${escapeHtml(letter)}</span></div>`,
    iconSize: [36, 36],
    iconAnchor: [18, 18],
    popupAnchor: [0, -16],
  });
}

function escapeHtml(value: string): string {
  return value.replace(/[&<>"']/g, (ch) => {
    switch (ch) {
      case "&":
        return "&amp;";
      case "<":
        return "&lt;";
      case ">":
        return "&gt;";
      case '"':
        return "&quot;";
      default:
        return "&#39;";
    }
  });
}

function Frame({ points, mode }: { points: [number, number][]; mode: "once" | "off" }) {
  const map = useMap();
  const fitted = useRef(false);

  useEffect(() => {
    if (mode !== "once" || fitted.current || points.length === 0) return;
    fitted.current = true;
    if (points.length === 1) {
      map.setView(points[0], 14);
      return;
    }
    map.fitBounds(points, { padding: [48, 48], maxZoom: 15 });
  }, [map, mode, points]);

  return null;
}

export function CityMap({
  markers,
  trails = [],
  center = MAP_CENTER,
  zoom = 13,
  fit = "off",
}: {
  markers: MapMarker[];
  trails?: [number, number][][];
  center?: [number, number];
  zoom?: number;
  fit?: "once" | "off";
}) {
  const fitPoints = markers.map((marker) => [marker.lat, marker.lng] as [number, number]);

  return (
    <MapContainer center={center} zoom={zoom} className="h-full w-full" scrollWheelZoom zoomControl={false}>
      <ZoomControl position="bottomright" />
      <TileLayer
        attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
        url="https://tile.openstreetmap.org/{z}/{x}/{y}.png"
      />
      <Frame points={fitPoints} mode={fit} />
      {trails.map((trail, index) => (
        <Polyline
          key={`trail-${index}`}
          positions={trail}
          pathOptions={{ color: "#9adf1a", weight: 4, opacity: 0.85 }}
        />
      ))}
      {markers.map((marker) => (
        <Marker key={marker.id} position={[marker.lat, marker.lng]} icon={iconFor(marker)}>
          <Popup>
            <strong>{marker.label}</strong>
            {marker.detail ? <div>{marker.detail}</div> : null}
            <div>
              {marker.lat.toFixed(5)}, {marker.lng.toFixed(5)}
            </div>
          </Popup>
        </Marker>
      ))}
    </MapContainer>
  );
}
