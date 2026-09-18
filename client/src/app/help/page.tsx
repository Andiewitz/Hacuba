import Link from "next/link";

const faqs = [
  {
    q: "How do I save a home I like?",
    a: "Log in or create an account from the profile menu, then tap the heart on any listing. Your saved homes follow you across devices.",
  },
  {
    q: "How do listings get verified?",
    a: "Every listing is reviewed against our photo, location, and ownership checks before it appears in search results.",
  },
  {
    q: "Who do I contact about a listing?",
    a: "Open the listing and use the contact card to message the seller directly. Never send money before viewing a property in person.",
  },
];

export default function HelpPage() {
  return (
    <div className="min-h-screen bg-background">
      <main className="mx-auto max-w-3xl px-6 pt-12 pb-24 md:px-10">
        <p className="text-sm font-semibold text-muted-foreground">
          Help center
        </p>
        <h1 className="mt-2 font-heading text-4xl leading-tight font-bold">
          How can we help?
        </h1>
        <p className="mt-4 max-w-[65ch] leading-7 text-muted-foreground">
          Answers to the questions we hear most. Still stuck? Reach us through
          the project repository and we will point you the right way.
        </p>

        <div className="mt-8 flex flex-col gap-4">
          {faqs.map((item) => (
            <section
              key={item.q}
              className="rounded-lg border border-border bg-card p-6 text-card-foreground shadow-sm"
            >
              <h2 className="font-heading text-lg font-medium">{item.q}</h2>
              <p className="mt-2 max-w-[70ch] text-sm leading-6 text-muted-foreground">
                {item.a}
              </p>
            </section>
          ))}
        </div>

        <div className="mt-8 rounded-lg bg-primary p-6 text-primary-foreground shadow-md">
          <h2 className="font-heading text-lg font-medium">
            Something looks wrong?
          </h2>
          <p className="mt-2 text-sm leading-6 opacity-90">
            Report a listing or a bug so we can take a look as soon as possible.
          </p>
          <Link
            href="https://github.com/Andiewitz/Hacuba"
            target="_blank"
            rel="noreferrer"
            className="mt-4 inline-flex h-10 items-center rounded-md bg-background px-4 text-sm font-semibold text-foreground transition-colors hover:shadow-md"
          >
            Open GitHub
          </Link>
        </div>
      </main>
    </div>
  );
}
