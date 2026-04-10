import type { NextConfig } from 'next';

const isProd = process.env.APP_ENV === 'production' || process.env.NODE_ENV === 'production';
const isDev = !isProd;

const nextConfig: NextConfig = {
  output: 'standalone',
  reactStrictMode: true,
  poweredByHeader: false,
  compress: isProd,
  images: {
    remotePatterns: [
      { protocol: 'https', hostname: 'cards.scryfall.io' },
      { protocol: 'https', hostname: 'assets.pokemon.com' },
      { protocol: 'https', hostname: 'lh3.googleusercontent.com' },
      { protocol: 'https', hostname: 'www.gravatar.com' },
      { protocol: 'https', hostname: 'm.media-amazon.com' },
      { protocol: 'https', hostname: 'images.pokemontcg.io' },
    ],
  },
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'http://backend:8080/api/:path*',
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
  experimental: {
    cacheComponents: true,
    viewTransition: true,
  },
  turbopack: isDev ? {} : undefined,
  webpack: (config, { dev }) => {
    if (dev && isDev) {
      config.watchOptions = {
        poll: 800,
        aggregateTimeout: 300,
        ignored: ['**/node_modules', '**/.next'],
      };
    }
    return config;
  },
};

export default nextConfig;
