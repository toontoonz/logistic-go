"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

const LINKS = [
  { href: "/", label: "Live map" },
  { href: "/orders", label: "Orders" },
  { href: "/simulate", label: "Simulator" },
];

export function DeskNav() {
  const pathname = usePathname();

  return (
    <header className="sticky top-0 z-[800] flex h-14 items-center justify-between border-b border-line bg-ink/90 px-4 backdrop-blur">
      <div className="flex items-baseline gap-3">
        <Link href="/" className="font-serif text-2xl tracking-tight text-signal">
          Lane
        </Link>
        <span className="hidden text-xs uppercase tracking-[0.16em] text-fog sm:inline">
          Dispatch desk · no login yet
        </span>
      </div>
      <nav className="flex items-center gap-1 text-sm">
        {LINKS.map((link) => {
          const active = pathname === link.href;
          return (
            <Link
              key={link.href}
              href={link.href}
              className={`rounded-full px-3 py-1.5 ${active ? "bg-signal text-ink" : "text-fog hover:text-white"}`}
            >
              {link.label}
            </Link>
          );
        })}
      </nav>
    </header>
  );
}
