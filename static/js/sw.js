const CACHE_NAME = 'scan-monitor-v3'; // Increment the version to trigger an update
const urlsToCache = [
    '/ui/', // This is the main HTML page at your start_url
    '/ui/js/manifest.json', // The manifest file
    'https://unpkg.com/@zxing/browser@latest/umd/zxing-browser.min.js',
    'https://code.responsivevoice.org/responsivevoice.js?key=YfaIuqy2'
];

// Install the service worker and cache the app shell
self.addEventListener('install', event => {
    event.waitUntil(
        caches.open(CACHE_NAME)
            .then(cache => {
                console.log('Opened cache');
                return cache.addAll(urlsToCache);
            })
    );
});

// Serve cached content when offline
self.addEventListener('fetch', event => {
    event.respondWith(
        caches.match(event.request)
            .then(response => {
                // Return response from cache, or fetch from network if not cached
                return response || fetch(event.request);
            })
    );
});

// Clean up old caches on activation
self.addEventListener('activate', event => {
    const cacheWhitelist = [CACHE_NAME];
    event.waitUntil(
        caches.keys().then(cacheNames => {
            return Promise.all(
                cacheNames.map(cacheName => {
                    if (cacheWhitelist.indexOf(cacheName) === -1) {
                        return caches.delete(cacheName);
                    }
                })
            );
        })
    );
});