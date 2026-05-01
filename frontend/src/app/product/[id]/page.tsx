import { Metadata } from 'next';
import { fetchProduct } from '@/lib/api';
import ProductDetailClient from './ProductDetailClient';
import { formatProductDescription } from '@/lib/metadata';

interface Props {
  params: Promise<{ id: string }>;
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { id } = await params;
  try {
    const product = await fetchProduct(id);
    const title = `${product.name} - ${product.set_name} | El Bulk TCG`;
    const description = formatProductDescription(product);
    
    return {
      title,
      description,
      openGraph: {
        title,
        description,
        images: product.image_url ? [product.image_url] : ['/og-image.png'],
        type: 'website',
      },
      twitter: {
        card: 'summary_large_image',
        title,
        description,
        images: product.image_url ? [product.image_url] : ['/og-image.png'],
      },
    };
  } catch {
    return {
      title: 'Product Not Found | El Bulk TCG',
    };
  }
}

export default async function ProductDetailPage({ params }: Props) {
  const { id } = await params;
  let product;
  try {
    product = await fetchProduct(id);
  } catch {
    return <ProductDetailClient product={null} error={true} />;
  }

  if (!product) {
    return <ProductDetailClient product={null} error={true} />;
  }

  // Step 4.2: Structured Data (JSON-LD)
  const jsonLd = {
    '@context': 'https://schema.org',
    '@type': 'Product',
    name: product.name,
    image: product.image_url,
    description: product.description || `TCG Card: ${product.name} from ${product.set_name}`,
    brand: {
      '@type': 'Brand',
      name: product.tcg.toUpperCase(),
    },
    offers: {
      '@type': 'Offer',
      price: product.price,
      priceCurrency: 'COP',
      availability: product.stock > 0 ? 'https://schema.org/InStock' : 'https://schema.org/OutOfStock',
    },
    additionalProperty: [
      { '@type': 'PropertyValue', name: 'Set', value: product.set_name },
      { '@type': 'PropertyValue', name: 'Condition', value: product.condition },
      { '@type': 'PropertyValue', name: 'Rarity', value: product.rarity },
    ],
  };

  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
      />
      <ProductDetailClient product={product} />
    </>
  );
}
