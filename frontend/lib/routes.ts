import type { LatLng } from "./geo";

// A short eastbound path across central Bangkok, used by the simulator
// and as a one-click order preset. Coordinates are WGS84.
export const DEMO_ROUTE: LatLng[] = [
  { lat: 13.7462, lng: 100.5347 },
  { lat: 13.7398, lng: 100.5412 },
  { lat: 13.7308, lng: 100.5416 },
  { lat: 13.7278, lng: 100.5518 },
  { lat: 13.7372, lng: 100.5602 },
  { lat: 13.7306, lng: 100.5694 },
  { lat: 13.7241, lng: 100.5788 },
];

export const MAP_CENTER: [number, number] = [13.7355, 100.556];

export type PlacePreset = {
  label: string;
  customerName: string;
  pickupAddress: string;
  dropoffAddress: string;
  pickupLat: number;
  pickupLng: number;
  dropoffLat: number;
  dropoffLng: number;
};

export const PRESETS: PlacePreset[] = [
  {
    label: "Siam → Thong Lo",
    customerName: "Suda",
    pickupAddress: "Siam Paragon",
    dropoffAddress: "Thong Lo",
    pickupLat: DEMO_ROUTE[0].lat,
    pickupLng: DEMO_ROUTE[0].lng,
    dropoffLat: DEMO_ROUTE[DEMO_ROUTE.length - 1].lat,
    dropoffLng: DEMO_ROUTE[DEMO_ROUTE.length - 1].lng,
  },
  {
    label: "Chatuchak → Asok",
    customerName: "Arthit",
    pickupAddress: "Chatuchak Market",
    dropoffAddress: "Terminal 21 Asok",
    pickupLat: 13.7999,
    pickupLng: 100.5503,
    dropoffLat: 13.7374,
    dropoffLng: 100.5604,
  },
  {
    label: "ICONSIAM → Lumphini",
    customerName: "Ploy",
    pickupAddress: "ICONSIAM",
    dropoffAddress: "Lumphini Park",
    pickupLat: 13.7266,
    pickupLng: 100.5102,
    dropoffLat: 13.7308,
    dropoffLng: 100.5416,
  },
];
