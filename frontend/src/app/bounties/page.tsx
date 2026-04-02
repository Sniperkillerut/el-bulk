import { Metadata } from 'next';
import { fetchBounties } from '@/lib/api';
import PublicBountiesClient from '@/app/bounties/bounties-client';

export const metadata: Metadata = {
  title: 'Wanted Cards - El Bulk TCG',
  description: 'Cards we are actively looking to buy. Check our bounty list and sell to us!',
};

export const revalidate = 60; // Cache for 60 seconds

export default async function BountiesPage() {
  const bounties = await fetchBounties({ active: true }).catch(() => []);

  return (
    <div className="min-h-screen flex flex-col pt-24 bg-kraft-paper">
      <main className="flex-1 w-full relative z-10 pb-20">
        <div className="section-container pt-8 md:pt-16 pb-12">
          <PublicBountiesClient initialBounties={bounties} />
          
        </div>
      </main>
    </div>
  );
}
