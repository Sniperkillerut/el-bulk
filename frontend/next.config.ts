import type { NextConfig } from 'next';

const nextConfig: NextConfig = {
  output: 'standalone',
  images: {
    remotePatterns: [
      { protocol: 'https', hostname: 'cards.scryfall.io' },
      { protocol: 'https', hostname: 'assets.pokemon.com' },
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
  turbopack: {},
  webpack: (config, { dev, isServer }) => {
    if (dev) {
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
