import { Metadata } from 'next';
import { fetchBounties } from '@/lib/api';
import PublicBountiesClient from './bounties-client';

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
          
          <div className="max-w-4xl mx-auto text-center mb-16 animate-in slide-in-from-bottom-5 fade-in duration-700">
            <h1 className="font-display text-5xl md:text-7xl xl:text-8xl tracking-tighter text-ink-deep leading-[0.85] mb-6">
              WANTED<span className="text-gold"> / </span><br/>BOUNTIES
            </h1>
            <p className="text-text-secondary md:text-lg max-w-2xl mx-auto tracking-wide">
              We are actively looking to buy the cards below. If you have them, reach out to us! 
              Can't find what you are looking for? Send us a card request!
            </p>
          </div>
          
          <PublicBountiesClient initialBounties={bounties} />
          
        </div>
      </main>
    </div>
  );
}
