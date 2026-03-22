import type { Metadata } from 'next';
import { Inter } from 'next/font/google';
import './globals.css';
import Navbar from '@/components/Navbar';
import { CartProvider } from '@/lib/CartContext';
import ProductModalManager from '@/components/ProductModalManager';

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
        <CartProvider>
          <Navbar />
          <ProductModalManager />
          <main style={{ minHeight: 'calc(100vh - 64px)' }}>
            {children}
          </main>
          <footer style={{ background: 'var(--ink-navy)', borderTop: '1px solid var(--ink-border)', padding: '2rem 1rem' }}>
            <div className="centered-container flex flex-col md:flex-row items-center justify-between gap-4">
              <div>
                <span className="font-display text-xl text-gold">EL BULK</span>
                <span className="text-xs ml-2" style={{ color: 'var(--text-muted)', fontFamily: 'Space Mono, monospace' }}>TCG STORE</span>
              </div>
              <p className="text-xs text-center" style={{ color: 'var(--text-muted)' }}>
                We buy bulk. We sell singles. We love cardboard.
              </p>
              <p className="text-xs" style={{ color: 'var(--text-muted)', fontFamily: 'Space Mono, monospace' }}>
                © {new Date().getFullYear()} El Bulk
              </p>
            </div>
          </footer>
        </CartProvider>
      </body>
    </html>
  );
}
