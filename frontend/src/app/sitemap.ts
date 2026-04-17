import { MetadataRoute } from 'next';

interface SitemapData {
  products: string[];
  notices: string[];
  tcgs: string[];
  collections: string[];
}

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const baseUrl = 'https://elbulk.com';
  const apiBase = process.env.INTERNAL_API_URL || process.env.NEXT_PUBLIC_API_URL || 'http://backend:8080';

  let dynamicData: SitemapData = {
    products: [],
    notices: [],
    tcgs: [],
    collections: [],
  };

  try {
    const res = await fetch(`${apiBase}/api/seo/sitemap-data`, {
      next: { revalidate: 3600 }, // Cache sitemap data for 1 hour
    });
    if (res.ok) {
        const rawData = await res.json();
        // The backend render.Success writes data directly, not wrapped in a "data" field
        if (rawData) {
            dynamicData = {
                products: rawData.products || [],
                notices: rawData.notices || [],
                tcgs: rawData.tcgs || [],
                collections: rawData.collections || [],
            };
        }
    }
  } catch (error) {
    console.error('Failed to fetch sitemap data:', error);
  }

  // 1. Define Static Routes
  const staticRoutes = [
    '',
    '/about',
    '/contact',
    '/shipping',
    '/refunds',
    '/terms',
    '/privacy',
    '/bounties',
    '/bulk',
    '/deck-importer',
    '/singles',
    '/sealed',
    '/store-exclusives',
  ].map((route) => ({
    url: `${baseUrl}${route}`,
    lastModified: new Date(),
    changeFrequency: 'daily' as const,
    priority: route === '' ? 1.0 : 0.8,
  }));

  // 2. Map Dynamic TCG Hubs
  const tcgRoutes = dynamicData.tcgs.flatMap((tcg) => [
    {
      url: `${baseUrl}/${tcg}/singles`,
      lastModified: new Date(),
      changeFrequency: 'daily' as const,
      priority: 0.9,
    },
    {
      url: `${baseUrl}/${tcg}/sealed`,
      lastModified: new Date(),
      changeFrequency: 'daily' as const,
      priority: 0.8,
    },
  ]);

  // 3. Map Dynamic Collections
  const collectionRoutes = dynamicData.collections.map((slug) => ({
    url: `${baseUrl}/collection/${slug}`,
    lastModified: new Date(),
    changeFrequency: 'weekly' as const,
    priority: 0.7,
  }));

  // 4. Map Dynamic Notices
  const noticeRoutes = dynamicData.notices.map((slug) => ({
    url: `${baseUrl}/notices/${slug}`,
    lastModified: new Date(),
    changeFrequency: 'weekly' as const,
    priority: 0.6,
  }));

  // 5. Map Dynamic Products
  const productRoutes = dynamicData.products.map((id) => ({
    url: `${baseUrl}/product/${id}`,
    lastModified: new Date(),
    changeFrequency: 'weekly' as const,
    priority: 0.5,
  }));

  return [
    ...staticRoutes,
    ...tcgRoutes,
    ...collectionRoutes,
    ...noticeRoutes,
    ...productRoutes,
  ];
}
