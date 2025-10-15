// See https://svelte.dev/docs/kit/types#app.d.ts

import type { SessionResponseBody } from "$lib/schema";

// for information about these interfaces
declare global {
  namespace App {
    // interface Error {}
    interface Locals {
      session: SessionResponseBody;
    }
    // interface PageData {}
    // interface PageState {}
    // interface Platform {}
  }
}

export {};
