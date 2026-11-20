export type LatLng = { lat: number; lng: number };

const EARTH_M = 6_371_000;

export function haversine(a: LatLng, b: LatLng): number {
  const p1 = (a.lat * Math.PI) / 180;
  const p2 = (b.lat * Math.PI) / 180;
  const dp = ((b.lat - a.lat) * Math.PI) / 180;
  const dl = ((b.lng - a.lng) * Math.PI) / 180;
  const h = Math.sin(dp / 2) ** 2 + Math.cos(p1) * Math.cos(p2) * Math.sin(dl / 2) ** 2;
  return 2 * EARTH_M * Math.asin(Math.min(1, Math.sqrt(h)));
}

export function pathLength(path: LatLng[]): number {
  let total = 0;
  for (let i = 1; i < path.length; i += 1) {
    total += haversine(path[i - 1], path[i]);
  }
  return total;
}

export function pointAlong(path: LatLng[], distance: number): LatLng {
  if (path.length === 0) return { lat: 13.736, lng: 100.555 };
  if (path.length === 1 || distance <= 0) return path[0];
  let left = distance;
  for (let i = 1; i < path.length; i += 1) {
    const segment = haversine(path[i - 1], path[i]);
    if (left <= segment) {
      const t = segment === 0 ? 0 : left / segment;
      return {
        lat: path[i - 1].lat + (path[i].lat - path[i - 1].lat) * t,
        lng: path[i - 1].lng + (path[i].lng - path[i - 1].lng) * t,
      };
    }
    left -= segment;
  }
  return path[path.length - 1];
}

export function ago(iso: string, now: number): string {
  const seconds = Math.max(0, Math.round((now - Date.parse(iso)) / 1000));
  if (seconds < 60) return `${seconds}s ago`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  return `${Math.floor(minutes / 60)}h ago`;
}
