// Disable SSR for all pages (required for Wails static embedding)
// Disable prerender because we have dynamic routes like /notes/[id]
export const ssr = false;
export const prerender = false;
