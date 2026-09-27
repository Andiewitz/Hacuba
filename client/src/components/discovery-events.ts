"use client";

type EventType = "impression" | "card_click" | "detail_view" | "contact_reveal";
type DiscoveryEvent = { listing_id: string; event_type: EventType; query?: string };
let pending: DiscoveryEvent[] = [];
let timer: ReturnType<typeof setTimeout> | undefined;

export function trackDiscovery(event: DiscoveryEvent) {
  pending.push(event);
  if (pending.length >= 20) return flushDiscovery();
  if (!timer) timer = setTimeout(flushDiscovery, 4000);
}

export function flushDiscovery() {
  if (!pending.length) return;
  const events = pending.splice(0, 25);
  if (timer) { clearTimeout(timer); timer = undefined; }
  const body = JSON.stringify({ events });
  if (navigator.sendBeacon?.("/api/discovery/events", new Blob([body], { type: "application/json" }))) return;
  void fetch("/api/discovery/events", { method: "POST", headers: { "content-type": "application/json" }, body, keepalive: true, credentials: "same-origin" });
}
