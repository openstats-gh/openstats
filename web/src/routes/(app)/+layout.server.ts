// import { redirect } from "@sveltejs/kit";
import type { LayoutServerLoad } from "./$types";

export const load: LayoutServerLoad = ({ locals }) => {
  if (!locals.session) {
    // todo: redirect here if not the main page, not sign-in
    // redirect(302, "/");
  }
  return {
    session: locals.session,
  };
};
