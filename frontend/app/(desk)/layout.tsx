import { DeskNav } from "@/components/DeskNav";

export default function DeskLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-dvh bg-ink text-white">
      <DeskNav />
      {children}
    </div>
  );
}
