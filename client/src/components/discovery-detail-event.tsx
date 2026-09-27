"use client";

import { useEffect } from "react";
import { trackDiscovery } from "@/components/discovery-events";

export default function DiscoveryDetailEvent({ listingID }: { listingID: string }) {
  useEffect(() => { trackDiscovery({ listing_id: listingID, event_type: "detail_view" }); }, [listingID]);
  return null;
}
