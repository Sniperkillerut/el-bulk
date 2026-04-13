import type { NextConfig } from 'next';

const isProd = process.env.APP_ENV === 'production' || process.env.NODE_ENV === 'production';
const isDev = !isProd;

const nextConfig: NextConfig = {
  output: 'standalone',
  reactStrictMode: true,
  poweredByHeader: false,
  compress: isProd,
  turbopack: {},
  images: {
    remotePatterns: [
      { protocol: 'https', hostname: 'cards.scryfall.io' },
      { protocol: 'https', hostname: 'assets.pokemon.com' },
      { protocol: 'https', hostname: 'lh3.googleusercontent.com' },
      { protocol: 'https', hostname: 'www.gravatar.com' },
      { protocol: 'https', hostname: 'm.media-amazon.com' },
      { protocol: 'https', hostname: 'images.pokemontcg.io' },
      { protocol: 'https', hostname: 'images.ygoprodeck.com' },
      { protocol: 'https', hostname: 'lorcana-api.com' },
      { protocol: 'https', hostname: 'en.onepiece-cardgame.com' },
      { protocol: 'https', hostname: 'api.dicebear.com' },
      { protocol: 'https', hostname: 'images.unsplash.com' },
      { protocol: 'https', hostname: 'cdn.gamegenic.com' },
      { protocol: 'https', hostname: 'cdn.shopify.com' },
      { protocol: 'https', hostname: 'ultimateguard.com' },
      { protocol: 'https', hostname: 'www.dragonshield.com' },
      { protocol: 'https', hostname: 'images.cardboardcrack.com' },
      { protocol: 'https', hostname: 'coordinadora.com' },
      { protocol: 'https', hostname: 'interrapidisimo.com' },
      { protocol: 'https', hostname: 'tcc.com.co' },
      { protocol: 'https', hostname: 'www.servientrega.com' },
      { protocol: 'https', hostname: 'storage.googleapis.com' },
    ],
  },
  async rewrites() {
    // INTERNAL_API_URL: Used for server-side requests (SSR).
    // In Cloud Run, this should be the backend's internal or public URL.
    // If missing, we fall back to NEXT_PUBLIC_API_URL to ensure production stability.
    const apiBase = process.env.INTERNAL_API_URL || 
                    process.env.NEXT_PUBLIC_API_URL || 
                    'http://backend:8080';
    return [
      {
        source: '/api/:path*',
        destination: `${apiBase}/api/:path*`,
      },
    ];
  },
  async headers() {
    return [
      {
        source: '/(.*)',
        headers: [
          { key: 'X-Frame-Options', value: 'DENY' },
          { key: 'X-Content-Type-Options', value: 'nosniff' },
          { key: 'Referrer-Policy', value: 'strict-origin-when-cross-origin' },
          { key: 'X-XSS-Protection', value: '1; mode=block' },
        ],
      },
    ];
  },
  devIndicators: isDev ? { position: 'bottom-right' } : false,
  cacheComponents: true,
  experimental: {
    viewTransition: true,
  },
};

if (isDev) {
  nextConfig.turbopack = {};
  nextConfig.webpack = (config) => {
    config.watchOptions = {
      poll: 800,
      aggregateTimeout: 300,
      ignored: ['**/node_modules', '**/.next'],
    };
    return config;
  };
}

export default nextConfig;
