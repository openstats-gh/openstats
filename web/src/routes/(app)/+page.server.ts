import type { LayoutServerLoad } from "./(pages)/$types";

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
