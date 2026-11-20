import type { Metadata } from "next";
import { Tracker } from "./tracker";

export const metadata: Metadata = { title: "Parcel tracking" };

export default async function TrackPage({ params }: { params: Promise<{ orderId: string }> }) {
  const { orderId } = await params;
  return <Tracker orderId={orderId} />;
}
