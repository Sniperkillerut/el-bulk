import type { Metadata } from 'next';
import { Inter, Bebas_Neue, Space_Mono } from 'next/font/google';
import './globals.css';
import StorefrontLayoutWrapper from '@/components/StorefrontLayoutWrapper';
import { CartProvider } from '@/lib/CartContext';
import RemoteLogManager from '@/components/RemoteLogManager';
import Script from 'next/script';

import { UserProvider } from '@/context/UserContext';
import { UIProvider } from '@/context/UIContext';
import { ToastProvider } from '@/context/ToastContext';
import './foil-effects.css';

const inter = Inter({ subsets: ['latin'], variable: '--font-inter' });
const bebas = Bebas_Neue({ weight: '400', subsets: ['latin'], variable: '--font-bebas' });
const spaceMono = Space_Mono({ weight: ['400', '700'], subsets: ['latin'], variable: '--font-mono' });
import { ThemeProvider } from '@/components/ThemeProvider';
import { fetchPublicSettings } from '@/lib/api';
import { fetchThemes } from '@/lib/api_themes';
import { Suspense } from 'react';

declare global {
  interface Window {
    gtag: (command: string, action: string, params: object) => void;
    fbq: (command: string, action: string, params?: object | string) => void;
    hj: (command: string, action: string, params?: string) => void;
  }
}

export const metadata: Metadata = {
  title: 'El Bulk — TCG Store',
  description: 'Your local Magic: The Gathering, Pokémon, Lorcana and One Piece card shop in Bogotá. Buy singles, sealed product, and sell us your bulk with secure evaluation.',
  keywords: ['MTG', 'Magic the Gathering', 'Pokemon', 'Lorcana', 'TCG', 'card store', 'singles', 'sealed', 'bulk', 'Bogota', 'Colombia'],
  authors: [{ name: 'El Bulk Collective' }],
  openGraph: {
    title: 'El Bulk — TCG Store',
    description: 'Premier destination for TCG enthusiasts. Secure buying, selling, and trading in Bogotá.',
    url: 'https://elbulk.com',
    siteName: 'El Bulk',
    images: [
      {
        url: '/og-image.png',
        width: 1200,
        height: 630,
        alt: 'El Bulk Storefront',
      },
    ],
    locale: 'es_CO',
    type: 'website',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'El Bulk — TCG Store',
    description: 'Premier destination for TCG enthusiasts. Secure buying, selling, and trading in Bogotá.',
    images: ['/og-image.png'],
  },
  robots: {
    index: true,
    follow: true,
    googleBot: {
      index: true,
      follow: true,
      'max-video-preview': -1,
      'max-image-preview': 'large',
      'max-snippet': -1,
    },
  },
};

import { LanguageProvider } from '@/context/LanguageContext';

/**
 * RootProviders handles asynchronous configuration fetching.
 * Wrapped in Suspense in the main layout to satisfy Next.js 15 Dynamic IO checks.
 */
async function RootProviders({ children }: { children: React.ReactNode }) {
  let defaultTheme = '00000000-0000-0000-0000-000000000001'; // Default: Cardboard
  let themes: import('@/lib/types').Theme[] = [];

  try {
    // Sequential fetching with force-cache to satisfy Next.js 15 build-time checks
    const settings = await fetchPublicSettings({ cache: 'force-cache' });
    if (settings && settings.default_theme_id) {
      defaultTheme = settings.default_theme_id;
    }
    
    const fetchedThemes = await fetchThemes({ cache: 'force-cache' });
    if (fetchedThemes) {
      themes = fetchedThemes;
    }
  } catch {
    // Fallback to defaults if API is unreachable during build
  }

  return (
    <ThemeProvider allThemes={themes} defaultTheme={defaultTheme}>
      <LanguageProvider>
        <UserProvider>
          <UIProvider>
            <CartProvider>
              <ToastProvider>
                <StorefrontLayoutWrapper>
                  {children}
                </StorefrontLayoutWrapper>
              </ToastProvider>
            </CartProvider>
          </UIProvider>
        </UserProvider>
      </LanguageProvider>
    </ThemeProvider>
  );
}


 
export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={`${inter.variable} ${bebas.variable} ${spaceMono.variable}`} suppressHydrationWarning>
      <body suppressHydrationWarning>
        <RemoteLogManager />
        
        <Suspense fallback={<div className="min-h-screen bg-bg-page animate-pulse" />}>
          <RootProviders>
            {children}
          </RootProviders>
        </Suspense>

        {/* Scripts at the end of body for stability */}
        <Script
          id="consent-mgmt"
          strategy="afterInteractive"
          dangerouslySetInnerHTML={{
            __html: `
              window.dataLayer = window.dataLayer || [];
              function gtag(){dataLayer.push(arguments);}
              gtag('consent', 'default', {
                'ad_storage': 'denied',
                'analytics_storage': 'denied',
                'ad_user_data': 'denied',
                'ad_personalization': 'denied',
                'personalization_storage': 'denied',
                'functionality_storage': 'granted',
                'security_storage': 'granted'
              });

              ${process.env.NEXT_PUBLIC_META_PIXEL_ID ? `
              !function(f,b,e,v,n,t,s)
              {if(f.fbq)return;n=f.fbq=function(){n.callMethod?
              n.callMethod.apply(n,arguments):n.queue.push(arguments)};
              if(!f._fbq)f._fbq=n;n.push=n;n.loaded=!0;n.version='2.0';
              n.queue=[];t=b.createElement(e);t.async=!0;
              t.src=v;s=b.getElementsByTagName(e)[0];
              s.parentNode.insertBefore(t,s)}(window, document,'script',
              'https://connect.facebook.net/en_US/fbevents.js');
              fbq('init', '${process.env.NEXT_PUBLIC_META_PIXEL_ID}');
              fbq('track', 'PageView');
              ` : ''}

              ${process.env.NEXT_PUBLIC_HOTJAR_ID ? `
              // ContentSquare Tag Manager (Migration from Hotjar)
              (function() {
                var s = document.createElement('script');
                s.async = true;
                s.src = 'https://t.contentsquare.net/uxa/' + '${process.env.NEXT_PUBLIC_HOTJAR_ID}' + '.js';
                document.getElementsByTagName('head')[0].appendChild(s);
              })();
              ` : ''}
            `,
          }}
        />

        {process.env.NEXT_PUBLIC_GA_ID && (
          <>
            <Script
              id="google-analytics-src"
              strategy="afterInteractive"
              src={`https://www.googletagmanager.com/gtag/js?id=${process.env.NEXT_PUBLIC_GA_ID}`}
            />
            <Script
              id="google-analytics-init"
              strategy="afterInteractive"
              dangerouslySetInnerHTML={{
                __html: `
                  window.dataLayer = window.dataLayer || [];
                  function gtag(){dataLayer.push(arguments);}
                  gtag('js', new Date());
                  gtag('config', '${process.env.NEXT_PUBLIC_GA_ID}', {
                    page_path: window.location.pathname,
                  });
                `,
              }}
            />
          </>
        )}

        {process.env.NEXT_PUBLIC_GOOGLE_ADS_ID && (
          <Script
            id="google-ads"
            strategy="afterInteractive"
            src={`https://www.googletagmanager.com/gtag/js?id=${process.env.NEXT_PUBLIC_GOOGLE_ADS_ID}`}
          />
        )}

        {process.env.NEXT_PUBLIC_META_PIXEL_ID && (
          <noscript suppressHydrationWarning>
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              height="1"
              width="1"
              style={{ display: 'none' }}
              src={`https://www.facebook.com/tr?id=${process.env.NEXT_PUBLIC_META_PIXEL_ID}&ev=PageView&noscript=1`}
              alt=""
            />
          </noscript>
        )}
      </body>
    </html>
  );
}
