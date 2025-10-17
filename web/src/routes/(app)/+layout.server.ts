import type { LayoutServerLoad } from "./$types";

export const load: LayoutServerLoad = ({ locals }) => {
  if (!locals.session) {
    return {
      session: null,
    };
  }
  return {
    session: locals.session,
  };
};
