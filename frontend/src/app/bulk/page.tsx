import type { Metadata } from 'next';
import BulkPageClient from './BulkPageClient';

export const metadata: Metadata = {
  title: 'Sell Your Bulk — El Bulk TCG Store',
  description: 'Bring your bulk commons, uncommons, and rares. We pay cash on the spot.',
};

export default function BulkPage() {
  return <BulkPageClient />;
}
