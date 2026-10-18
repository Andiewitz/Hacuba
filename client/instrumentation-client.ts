import * as Sentry from "@sentry/nextjs";

const dsn = process.env.NEXT_PUBLIC_SENTRY_DSN;

Sentry.init({
  dsn,
  enabled: Boolean(dsn),
  tracesSampleRate: 0.1,
  beforeSend(event) {
    if (event.request) {
      delete event.request.cookies;
      delete event.request.headers;
      if (event.request.url) {
        event.request.url = event.request.url.split("?")[0];
      }
    }
    return event;
  },
});

export const onRouterTransitionStart = Sentry.captureRouterTransitionStart;
