import type { Metadata } from 'next';
import { Inter, Bebas_Neue, Space_Mono, Cinzel, Playfair_Display, Outfit, Roboto, Montserrat } from 'next/font/google';
import './globals.css';
import StorefrontLayoutWrapper from '@/components/StorefrontLayoutWrapper';
import { CartProvider } from '@/lib/CartContext';
import RemoteLogManager from '@/components/RemoteLogManager';
import Script from 'next/script';

import { UserProvider } from '@/context/UserContext';
import { UIProvider } from '@/context/UIContext';
import './foil-effects.css';

const inter = Inter({ subsets: ['latin'], variable: '--font-inter' });
const bebas = Bebas_Neue({ weight: '400', subsets: ['latin'], variable: '--font-bebas' });
const spaceMono = Space_Mono({ weight: ['400', '700'], subsets: ['latin'], variable: '--font-mono' });
const cinzel = Cinzel({ subsets: ['latin'], variable: '--font-cinzel' });
const playfair = Playfair_Display({ subsets: ['latin'], variable: '--font-playfair' });
const outfit = Outfit({ subsets: ['latin'], variable: '--font-outfit' });
const roboto = Roboto({ weight: ['400', '700'], subsets: ['latin'], variable: '--font-roboto' });
const montserrat = Montserrat({ subsets: ['latin'], variable: '--font-montserrat' });
import { ThemeProvider } from '@/components/ThemeProvider';
import { fetchPublicSettings } from '@/lib/api';

declare global {
  interface Window {
    gtag: (command: string, action: string, params: object) => void;
    fbq: (command: string, action: string, params?: object | string) => void;
    hj: (command: string, action: string, params?: string) => void;
  }
}

export const metadata: Metadata = {
  title: 'El Bulk — TCG Store',
  description: 'Your local Magic: The Gathering, Pokémon, Lorcana and One Piece card shop. Buy singles, sealed product, and sell us your bulk.',
  keywords: ['MTG', 'Magic the Gathering', 'Pokemon', 'Lorcana', 'TCG', 'card store', 'singles', 'sealed', 'bulk'],
};

import { LanguageProvider } from '@/context/LanguageContext';

export default async function RootLayout({ children }: { children: React.ReactNode }) {
  let defaultTheme = '00000000-0000-0000-0000-000000000001'; // Default to Cardboard
  try {
    const settings = await fetchPublicSettings();
    if (settings.default_theme_id) {
      defaultTheme = settings.default_theme_id;
    }
  } catch (err) {
    console.error('Failed to fetch default theme settings', err);
  }
  return (
    <html lang="en" className={`${inter.variable} ${bebas.variable} ${spaceMono.variable} ${cinzel.variable} ${playfair.variable} ${outfit.variable} ${roboto.variable} ${montserrat.variable}`} suppressHydrationWarning>
      <body suppressHydrationWarning>
        <RemoteLogManager />
        
        <ThemeProvider defaultTheme={defaultTheme}>
          <LanguageProvider>
            <UserProvider>
              <UIProvider>
                <CartProvider>
                  <StorefrontLayoutWrapper>
                    {children}
                  </StorefrontLayoutWrapper>
                </CartProvider>
              </UIProvider>
            </UserProvider>
          </LanguageProvider>
        </ThemeProvider>

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
              (function(h,o,t,j,a,r){
                h.hj=h.hj||function(){(h.hj.q=h.hj.q||[]).push(arguments)};
                h._hjSettings={hjid:${process.env.NEXT_PUBLIC_HOTJAR_ID},hjsv:6};
                a=o.getElementsByTagName('head')[0];
                r=o.createElement('script');r.async=1;
                r.src=t+h._hjSettings.hjid+j+h._hjSettings.hjsv;
                a.appendChild(r);
              })(window,document,'https://static.hotjar.com/c/hotjar-','.js?sv=');
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
