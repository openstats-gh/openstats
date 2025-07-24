import config from '$lib/config';
import type { Handle } from '@sveltejs/kit';

export const handle: Handle = async ({ event, resolve }) => {
  if (event.url.pathname.startsWith("/api")) {
    const newPath = event.url.pathname.substring("/api".length)
    console.log("/api request: " + newPath);
    return await fetch(new URL(newPath, config.apiBaseUrl), event.request)
  }

  return await resolve(event);
};