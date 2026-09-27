"use client";

import { useEffect, useState } from "react";
import { CheckCircle2, Inbox } from "lucide-react";
import { useAppSelector } from "@/lib/hooks";

type Report = {
  id: string;
  category: string;
  listing_reference?: string;
	listing_id?: string;
  contact_email?: string;
  description: string;
  status: "open" | "triaged" | "resolved";
  created_at: string;
  updated_at: string;
};

export default function SupportQueuePage() {
  const { user, accessToken, csrfToken, initialized } = useAppSelector((state) => state.auth);
  const [reports, setReports] = useState<Report[]>([]);
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(true);
	const [reasons, setReasons] = useState<Record<string, string>>({});
  const isStaff = user?.role === "staff" && Boolean(accessToken);

  useEffect(() => {
    if (!isStaff || !accessToken) return;
    let cancelled = false;
    fetch("/api/support/reports", { headers: { Authorization: `Bearer ${accessToken}` } })
      .then(async (response) => {
        if (!response.ok) throw new Error("Couldn't load reports.");
        return response.json() as Promise<{ reports: Report[] }>;
      })
      .then((value) => {
        if (!cancelled) setReports(value.reports);
      })
      .catch((error) => {
        if (!cancelled) setMessage(error instanceof Error ? error.message : "Couldn't load reports.");
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => { cancelled = true; };
  }, [accessToken, isStaff]);

  const update = async (id: string, status: "triaged" | "resolved") => {
    if (!accessToken) return;
    setMessage("");
    const response = await fetch(`/api/support/reports/${id}`, {
      method: "PATCH",
      headers: { Authorization: `Bearer ${accessToken}`, "Content-Type": "application/json" },
      body: JSON.stringify({ status }),
    });
    if (!response.ok) {
      setMessage("Couldn't update this report.");
      return;
    }
    const updated = (await response.json()) as Report;
    setReports((current) => current.map((report) => (report.id === id ? { ...report, ...updated } : report)));
  };

	const moderate = async (report: Report, action: "hide" | "restore") => {
		if (!accessToken || !csrfToken || !report.listing_id) return;
		const reason = reasons[report.id]?.trim() ?? "";
		if (action === "hide" && !reason) {
			setMessage("Add a clear reason before hiding this listing.");
			return;
		}
		setMessage("");
		const response = await fetch(`/api/listings/staff/${report.listing_id}/moderate`, {
			method: "POST",
			credentials: "same-origin",
			headers: { Authorization: `Bearer ${accessToken}`, "X-CSRF-Token": csrfToken, "Content-Type": "application/json" },
			body: JSON.stringify({ report_id: report.id, action, reason }),
		});
		if (!response.ok) {
			setMessage("Couldn’t apply that listing action.");
			return;
		}
		setReasons((current) => ({ ...current, [report.id]: "" }));
		setMessage(action === "hide" ? "Listing hidden. Mark the report triaged or resolved when ready." : "Listing restored.");
	};

  if (!initialized || (isStaff && loading)) return <LoadingQueue />;
  if (!isStaff) {
    return (
      <main className="min-h-screen bg-background px-6 py-[var(--space-16)]">
        <section className="mx-auto max-w-2xl rounded-[var(--radius-lg)] border border-border bg-card p-[var(--space-8)] text-center shadow-[var(--shadow-sm)]">
          <h1 className="font-heading text-[2.25rem] font-semibold text-foreground">Staff access required</h1>
          <p className="mt-3 text-sm leading-6 text-muted-foreground">This queue contains private support reports and is available only to designated staff.</p>
        </section>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-background px-6 py-[var(--space-16)] md:px-10 lg:px-20">
      <div className="mx-auto max-w-6xl">
        <p className="text-sm font-semibold text-[var(--color-terracotta-ink)]">Staff workspace</p>
        <h1 className="mt-2 font-heading text-[2.75rem] font-bold leading-[1.15] text-foreground">Support queue</h1>
        <p className="mt-3 max-w-[62ch] text-[1.125rem] leading-[1.6] text-muted-foreground">Private reports awaiting a status decision. Actions are recorded in the Support audit log.</p>
        {message && <p role="alert" className="mt-6 text-sm text-[var(--color-terracotta-ink)]">{message}</p>}
        <div className="mt-[var(--space-8)] space-y-[var(--space-4)]">
          {reports.length === 0 ? (
            <section className="rounded-[var(--radius-lg)] border border-border bg-card p-[var(--space-8)] text-center shadow-[var(--shadow-sm)]">
              <Inbox className="mx-auto h-8 w-8 text-muted-foreground" />
              <h2 className="mt-4 font-heading text-[1.75rem] font-semibold text-foreground">No reports yet</h2>
            </section>
          ) : reports.map((report) => (
            <article key={report.id} className="rounded-[var(--radius-lg)] border border-border bg-card p-[var(--space-6)] shadow-[var(--shadow-sm)]">
              <div className="flex flex-wrap items-start justify-between gap-4">
                <div>
                  <p className="text-sm font-semibold capitalize text-[var(--color-terracotta-ink)]">{report.category} · {report.status}</p>
                  <p className="mt-3 max-w-[72ch] whitespace-pre-wrap text-sm leading-6 text-foreground">{report.description}</p>
                  {report.listing_reference && <p className="mt-3 break-all text-sm text-muted-foreground">Listing: {report.listing_reference}</p>}
                  {report.contact_email && <p className="mt-1 text-sm text-muted-foreground">Reply: {report.contact_email}</p>}
				  {report.listing_id && <div className="mt-5 max-w-[52ch] rounded-[var(--radius-md)] border border-border bg-background p-4"><p className="text-sm font-semibold text-foreground">Listing moderation</p><p className="mt-1 text-sm leading-6 text-muted-foreground">A listing action is recorded with this report. The seller sees the decision reason, not reporter details.</p><label className="mt-3 block text-sm font-semibold text-foreground">Reason<textarea value={reasons[report.id] ?? ""} onChange={(event) => setReasons((current) => ({ ...current, [report.id]: event.target.value }))} maxLength={500} className="mt-2 min-h-20 w-full rounded-[var(--radius-sm)] border border-border bg-card px-3 py-2 text-sm font-normal text-foreground outline-none focus:ring-2 focus:ring-[var(--color-terracotta)]" placeholder="Explain the moderation decision to the seller." /></label><div className="mt-3 flex flex-wrap gap-2"><button type="button" onClick={() => moderate(report, "hide")} disabled={report.status === "resolved" || !(reasons[report.id] ?? "").trim()} className="rounded-[var(--radius-md)] bg-[var(--color-terracotta)] px-4 py-3 text-sm font-semibold text-black disabled:opacity-50">Hide listing</button><button type="button" onClick={() => moderate(report, "restore")} className="rounded-[var(--radius-md)] border border-border px-4 py-3 text-sm font-semibold text-foreground">Restore listing</button></div></div>}
                </div>
                {report.status !== "resolved" && (
                  <div className="flex gap-2">
                    {report.status === "open" && <button onClick={() => update(report.id, "triaged")} className="rounded-[var(--radius-md)] border border-border px-4 py-3 text-sm font-semibold text-foreground">Mark triaged</button>}
                    <button onClick={() => update(report.id, "resolved")} className="inline-flex items-center gap-2 rounded-[var(--radius-md)] bg-[var(--color-terracotta)] px-4 py-3 text-sm font-semibold text-black"><CheckCircle2 className="h-4 w-4" />Resolve</button>
                  </div>
                )}
              </div>
            </article>
          ))}
        </div>
      </div>
    </main>
  );
}

function LoadingQueue() {
  return <main className="min-h-screen bg-background px-6 py-[var(--space-16)] text-muted-foreground">Loading support queue…</main>;
}
