import type { Metadata } from 'next';
import { Inter, Bebas_Neue, Space_Mono } from 'next/font/google';
import './globals.css';
import StorefrontLayoutWrapper from '@/components/StorefrontLayoutWrapper';
import { CartProvider } from '@/lib/CartContext';
import RemoteLogManager from '@/components/RemoteLogManager';

import { UserProvider } from '@/context/UserContext';
import { UIProvider } from '@/context/UIContext';
import './foil-effects.css';

const inter = Inter({ subsets: ['latin'], variable: '--font-inter' });
const bebas = Bebas_Neue({ weight: '400', subsets: ['latin'], variable: '--font-bebas' });
const spaceMono = Space_Mono({ weight: ['400', '700'], subsets: ['latin'], variable: '--font-mono' });

export const metadata: Metadata = {
  title: 'El Bulk — TCG Store',
  description: 'Your local Magic: The Gathering, Pokémon, Lorcana and One Piece card shop. Buy singles, sealed product, and sell us your bulk.',
  keywords: ['MTG', 'Magic the Gathering', 'Pokemon', 'Lorcana', 'TCG', 'card store', 'singles', 'sealed', 'bulk'],
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={`${inter.variable} ${bebas.variable} ${spaceMono.variable}`} suppressHydrationWarning>
      <head>
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
          <UIProvider>
            <CartProvider>
              <StorefrontLayoutWrapper>
                {children}
              </StorefrontLayoutWrapper>
            </CartProvider>
          </UIProvider>
        </UserProvider>
      </body>
    </html>
  );
}
