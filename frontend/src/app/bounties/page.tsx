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
    <div className="min-h-screen flex flex-col bg-bg-page transition-colors duration-500">
      <main className="flex-1 w-full relative z-10 pt-16 md:pt-24 pb-20">
        <div className="centered-container px-4">
          <PublicBountiesClient initialBounties={bounties} />
        </div>
      </main>
    </div>
  );
}
