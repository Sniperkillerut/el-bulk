'use client';

import HomeSearchBar from '@/components/HomeSearchBar';
import { useLanguage } from '@/context/LanguageContext';

export default function HeroSection() {
  const { t } = useLanguage();

  return (
    <section className="relative w-full overflow-hidden bg-bg-kraft py-12 md:py-20 border-b border-border-plum/20">
      {/* Fixed Background Layer */}
      <div 
        className="absolute inset-0 z-0 animate-fade-in duration-1000"
        style={{ 
          backgroundImage: `url('/images/hero-bg.jpg')`,
          backgroundSize: 'cover',
          backgroundPosition: 'center',
        }}
      >
        {/* Subtle Craft Overlay to blend with the cardboard aesthetic */}
        <div className="absolute inset-0 bg-bg-kraft/60 backdrop-blur-[2px] paper-texture opacity-90" />
      </div>

      <div className="centered-container relative z-30 flex flex-col items-center text-center px-6">
        <div className="animate-fade-up flex flex-col items-center">
          {/* Humble Brand Header */}
          <h1 className="font-display text-fluid-h2 text-ink-plum mb-1 tracking-tight">
            EL BULK
          </h1>
          <p className="text-sm md:text-lg text-ink-plum/60 font-bold uppercase tracking-[0.3em] mb-8">
            {t('pages.home.hero.tagline', 'Grow your collection')}
          </p>

          {/* Simple Search Bar Area */}
          <div className="w-full max-w-2xl">
            <div className="border border-ink-plum/20 rounded-md overflow-hidden bg-white/40 backdrop-blur-md shadow-lg shadow-ink-plum/5">
              <HomeSearchBar placeholder={t('pages.home.hero.search_placeholder', 'Search for cards, sets, or accessories...')} />
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
