import { Product } from './types';
import { Metadata } from 'next';
import { fetchProduct } from './api';

export function formatProductDescription(product: Product): string {
  const technicalInfo = `[${product.set_name}] ${product.condition} ${product.foil_treatment}`;
  const priceInfo = `COP ${product.price.toLocaleString('en-US')}`;
  const stockInfo = product.stock > 0 ? `${product.stock} in stock` : 'Out of stock';
  
  return `Buy ${product.name} (${technicalInfo}) — ${priceInfo} | ${stockInfo}. Secure shipping from El Bulk Bogotá.`;
}

export async function getSharedProductMetadata(productId: string | null): Promise<Metadata | null> {
  if (!productId) return null;

  const siteUrl = process.env.NEXT_PUBLIC_SITE_URL || 'https://elbulk.com';

  try {
    const product = await fetchProduct(productId);
    if (!product) {
      console.warn(`[Metadata] Product not found: ${productId}`);
      return null;
    }

    const title = `${product.name} - ${product.set_name} | El Bulk TCG`;
    const description = formatProductDescription(product);
    
    // Ensure image URL is absolute for crawlers
    let imageUrl = product.image_url;
    if (imageUrl && !imageUrl.startsWith('http')) {
      imageUrl = `${siteUrl}${imageUrl.startsWith('/') ? '' : '/'}${imageUrl}`;
    }
    
    // Fallback if no product image
    const finalImageUrl = imageUrl || `${siteUrl}/og-image.png`;

    return {
      title,
      description,
      openGraph: {
        title,
        description,
        images: [finalImageUrl],
        type: 'website',
        url: `${siteUrl}/product/${product.id}`, // Canonical link to the product
      },
      twitter: {
        card: 'summary_large_image',
        title,
        description,
        images: [finalImageUrl],
      },
    };
  } catch (error) {
    console.error(`[Metadata] Error fetching for ${productId}:`, error);
    return null;
  }
}
