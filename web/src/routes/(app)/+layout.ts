import { api } from "$lib/api";
import { redirect } from "@sveltejs/kit";
import type { LayoutLoad } from "./$types";

export const load: LayoutLoad = async ({ fetch }) => {
  const session = await api.with(fetch).getSessionSummary();

  if (!session) {
    redirect(307, "/login");
  } else {
    return session;
  }
};
