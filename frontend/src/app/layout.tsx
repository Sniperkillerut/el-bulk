import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import './globals.css';
import Navbar from '@/components/Navbar';
import StorefrontLayoutWrapper from '@/components/StorefrontLayoutWrapper';
import { CartProvider } from '@/lib/CartContext';
import ProductModalManager from '@/components/ProductModalManager';
import BountyModalManager from '@/components/BountyModalManager';
import RemoteLogManager from '@/components/RemoteLogManager';

import { UserProvider } from '@/context/UserContext';

const inter = Inter({ subsets: ['latin'] });

export const metadata: Metadata = {
  title: 'El Bulk — TCG Store',
  description: 'Your local Magic: The Gathering, Pokémon, Lorcana and One Piece card shop. Buy singles, sealed product, and sell us your bulk.',
  keywords: ['MTG', 'Magic the Gathering', 'Pokemon', 'Lorcana', 'TCG', 'card store', 'singles', 'sealed', 'bulk'],
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={inter.className} suppressHydrationWarning>
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link
          href="https://fonts.googleapis.com/css2?family=Bebas+Neue&family=Space+Mono:wght@400;700&display=swap"
          rel="stylesheet"
        />
        <style>{`
          .centered-container {
            max-width: 1280px !important;
            margin-left: auto !important;
            margin-right: auto !important;
            width: 100% !important;
            display: block !important;
          }
          .responsive-stack {
            display: flex !important;
            flex-direction: column !important;
            width: 100% !important;
          }
          @media (min-width: 640px) {
            .responsive-stack {
              flex-direction: row !important;
            }
          }
          /* Fix Hero section alignment */
          section.box-lid > div.centered-container {
            display: flex !important;
            justify-content: center !important;
          }
          @media (min-width: 1024px) {
             section.box-lid > div.centered-container {
               justify-content: flex-start !important;
             }
          }
        `}</style>
      </head>
      <body suppressHydrationWarning>
        <RemoteLogManager />
        <UserProvider>
          <CartProvider>
            <StorefrontLayoutWrapper>
              {children}
            </StorefrontLayoutWrapper>
          </CartProvider>
        </UserProvider>
      </body>
    </html>
  );
}
