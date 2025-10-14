import { Client } from "$lib/internalApi";
import type { LayoutServerLoad } from "./$types";

export const load: LayoutServerLoad = async ({ fetch }) => {
  const { data, error } = await Client.GET("/internal/session/", {
    fetch: fetch,
  });

  if (error && error.status !== 401) {
    console.error("Error getting session", error);
  }

  return {
    session: data,
  };
};
