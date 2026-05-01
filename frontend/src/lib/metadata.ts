import { Product } from './types';

export function formatProductDescription(product: Product): string {
  const technicalInfo = `[${product.set_name}] ${product.condition} ${product.foil_treatment}`;
  const priceInfo = `COP ${product.price.toLocaleString('en-US')}`;
  const stockInfo = product.stock > 0 ? `${product.stock} in stock` : 'Out of stock';
  
  return `Buy ${product.name} (${technicalInfo}) — ${priceInfo} | ${stockInfo}. Secure shipping from El Bulk Bogotá.`;
}
