// LOCKED: Azion Edge Function entry point
// Do not modify without understanding Azion deployment

import app from './index';

addEventListener('fetch', (event: FetchEvent) => {
  setAzionArgs(event.args);
  event.respondWith(app.fetch(event.request));
});
