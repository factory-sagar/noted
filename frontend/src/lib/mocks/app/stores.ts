import { writable } from 'svelte/store';

export const page = writable({
  url: new URL('http://localhost/'),
  params: {},
  route: { id: null },
  status: 200,
  error: null,
  data: {},
  form: null
});

export const navigating = writable(null);
export const updated = writable(false);

export const getStores = () => ({
  page,
  navigating,
  updated
});
