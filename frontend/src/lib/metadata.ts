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

  try {
    const product = await fetchProduct(productId);
    if (!product) return null;

    const title = `${product.name} — El Bulk TCG`;
    const description = formatProductDescription(product);
    const imageUrl = product.image_url;

    return {
      title,
      description,
      openGraph: {
        title,
        description,
        images: imageUrl ? [{ url: imageUrl, alt: product.name }] : [],
      },
      twitter: {
        card: 'summary_large_image',
        title,
        description,
        images: imageUrl ? [imageUrl] : [],
      },
    };
  } catch (error) {
    console.error('Failed to fetch metadata for product:', productId, error);
    return null;
  }
}
