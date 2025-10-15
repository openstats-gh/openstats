import config from "$lib/config";
import { Client } from "$lib/internalApi";
import type { Handle } from "@sveltejs/kit";

export const handle: Handle = async ({ event, resolve }) => {
  if (event.url.pathname.startsWith("/internal")) {
    const result = await fetch(new URL(event.url.pathname, config.apiBaseUrl), event.request);
    return result;
  }

  if (!event.locals.session) {
    const { data, error } = await Client.GET("/internal/session/", {
      baseUrl: config.apiBaseUrl,
      headers: event.request.headers,
      credentials: "include",
    });

    if (!error) {
      event.locals.session = data;
    } else if (error.status !== 401) {
      console.error("There was an error retrieving the session", error);
    }
  }

  return await resolve(event, {
    filterSerializedResponseHeaders(name) {
      return name === "content-length";
    },
  });
};
