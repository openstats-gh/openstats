import type { paths } from "$lib/schema";
import createClient from "openapi-fetch";

export const Client = createClient<paths>({
  baseUrl: "http://localhost:5173/",
});
