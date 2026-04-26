'use client';

import HomeSearchBar from '@/components/HomeSearchBar';
import { useLanguage } from '@/context/LanguageContext';

export default function HeroSection() {
  const { t } = useLanguage();

  return (
    <section className="relative w-full bg-bg-page py-12 md:py-20 border-b border-border-main/20 overflow-hidden">
      {/* Fixed Background Layer */}
      <div 
        className="absolute inset-0 z-0 animate-fade-in duration-1000"
        style={{ 
          backgroundImage: `url('/images/hero-bg.jpg')`,
          backgroundSize: 'cover',
          backgroundPosition: 'center',
        }}
      >
        {/* Enhanced Overlay for readability */}
        <div className="absolute inset-0 bg-bg-page/75 backdrop-blur-[4px] paper-texture opacity-90" />
        {/* Radial gradient for center focus and vignette effect */}
        <div className="absolute inset-0 bg-[radial-gradient(circle_at_center,transparent_0%,var(--bg-page)_100%)] opacity-40" />
        <div className="absolute inset-0 bg-gradient-to-b from-transparent via-bg-page/10 to-bg-page/40" />
      </div>

      <div className="centered-container relative z-30 flex flex-col items-center text-center px-6">
        <div className="animate-fade-up flex flex-col items-center w-full max-w-2xl">
          {/* Brand Header with shadow for contrast */}
          <h1 className="font-display text-fluid-h2 text-text-main mb-1 tracking-tight hero-text-glow">
            EL BULK
          </h1>
          <p className="text-sm md:text-lg text-text-main/70 font-bold uppercase tracking-[0.3em] mb-10 hero-text-glow">
            {t('pages.home.hero.tagline', 'Grow your collection')}
          </p>

          {/* Simple Search Bar Area */}
          <div className="w-full max-w-6xl">
            <div className="relative z-40 border border-text-main/10 rounded-md bg-white/40 backdrop-blur-xl shadow-2xl shadow-text-main/5">
              <HomeSearchBar placeholder={t('pages.home.hero.search_placeholder', 'Search for cards, sets, or accessories...')} />
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
