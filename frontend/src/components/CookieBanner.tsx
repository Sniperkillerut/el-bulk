'use client';

import React, { useState, useEffect } from 'react';
import { useLanguage } from '@/context/LanguageContext';

type CookiePreferences = {
  essential: boolean;
  analytics: boolean;
  marketing: boolean;
};

const CookieBanner = () => {
  const [isVisible, setIsVisible] = useState(false);
  const [showPreferences, setShowPreferences] = useState(false);
  const [prefs, setPrefs] = useState<CookiePreferences>(() => {
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem('cookie-preferences');
      if (saved) {
        try {
          return JSON.parse(saved) as CookiePreferences;
        } catch (e) {
          console.error('Error parsing cookie prefs', e);
        }
      }
    }
    return {
      essential: true,
      analytics: false,
      marketing: false,
    };
  });
  
  const { t } = useLanguage();

  const updateGA4Consent = (preferences: CookiePreferences) => {
    if (typeof window !== 'undefined') {
      // Google Analytics & Ads Consent Mode v2
      if (window.gtag) {
        window.gtag('consent', 'update', {
          'analytics_storage': preferences.analytics ? 'granted' : 'denied',
          'ad_storage': preferences.marketing ? 'granted' : 'denied',
          'ad_user_data': preferences.marketing ? 'granted' : 'denied',
          'ad_personalization': preferences.marketing ? 'granted' : 'denied',
          'personalization_storage': preferences.marketing ? 'granted' : 'denied',
        });
      }

      // Meta Pixel Consent
      const fbq = window.fbq;
      if (typeof fbq === 'function') {
        fbq('consent', preferences.marketing ? 'grant' : 'revoke');
      }

      // Hotjar Consent (Linked to Analytics)
      const hj = window.hj;
      if (typeof hj === 'function') {
        hj('consent', preferences.analytics ? 'grant' : 'retract');
      }
    }
  };

  useEffect(() => {
    const savedPrefs = localStorage.getItem('cookie-preferences');
    if (savedPrefs) {
      try {
        const parsed = JSON.parse(savedPrefs) as CookiePreferences;
        updateGA4Consent(parsed);
      } catch (e) {
        console.error('Failed to parse cookie preferences', e);
      }
    } else {
      // Show banner if no preferences exist
      const timer = setTimeout(() => setIsVisible(true), 1500);
      return () => clearTimeout(timer);
    }
  }, []);

  const handleAcceptAll = () => {
    const allAccepted = { essential: true, analytics: true, marketing: true };
    setPrefs(allAccepted);
    localStorage.setItem('cookie-preferences', JSON.stringify(allAccepted));
    updateGA4Consent(allAccepted);
    setIsVisible(false);
  };

  const handleDeclineAll = () => {
    const allDeclined = { essential: true, analytics: false, marketing: false };
    setPrefs(allDeclined);
    localStorage.setItem('cookie-preferences', JSON.stringify(allDeclined));
    updateGA4Consent(allDeclined);
    setIsVisible(false);
  };

  const handleSavePreferences = () => {
    localStorage.setItem('cookie-preferences', JSON.stringify(prefs));
    updateGA4Consent(prefs);
    setIsVisible(false);
  };

  if (!isVisible) {
    return (
      <button
        onClick={() => setIsVisible(true)}
        className="fixed bottom-4 left-4 z-[var(--z-drawer)] p-2 rounded-full bg-bg-header/20 backdrop-blur-md border border-white/10 text-white/40 hover:text-white hover:bg-bg-header/40 transition-all duration-300 group"
        title={t('pages.cookies.manage', 'Manage Cookies')}
      >
        <svg 
          viewBox="0 0 24 24" 
          fill="none" 
          stroke="currentColor" 
          strokeWidth="2" 
          strokeLinecap="round" 
          strokeLinejoin="round" 
          className="w-5 h-5 group-hover:rotate-12 transition-transform"
        >
          <path d="M12 2a10 10 0 1 0 10 10 4 4 0 0 1-5-3 4 4 0 0 1-3-5 4 4 0 0 1-2 8z" />
          <path d="M12 11v.01" />
          <path d="M16 14v.01" />
          <path d="M8 14v.01" />
          <path d="M12 17v.01" />
        </svg>
      </button>
    );
  }

  return (
    <div 
      className="fixed bottom-0 left-0 right-0 z-[var(--z-drawer)] p-4 md:p-6 animate-fade-up"
      style={{ animationDuration: 'var(--duration-slow)', animationTimingFunction: 'var(--ease-out)' }}
    >
      <div 
        className="max-w-4xl mx-auto backdrop-blur-xl bg-bg-header/85 border border-white/10 rounded-2xl p-6 shadow-2xl overflow-hidden transition-all duration-500 ease-in-out"
        style={{ maxHeight: showPreferences ? '600px' : '300px' }}
      >
        {!showPreferences ? (
          <div className="flex flex-col md:flex-row items-center justify-between gap-6">
            <div className="flex-1">
              <div className="flex items-center gap-3 mb-2">
                <span className="w-2 h-2 rounded-full bg-accent-primary animate-pulse" />
                <h3 className="text-text-on-header font-display tracking-widest text-lg m-0 uppercase">
                  {t('pages.cookies.title', 'COOKIES')}
                </h3>
              </div>
              <p className="text-text-on-header/80 text-sm leading-relaxed m-0">
                {t('pages.cookies.description', 'We use cookies to improve your experience. You can choose to accept all or customize your preferences.')}
              </p>
            </div>
            
            <div className="flex flex-wrap items-center justify-end gap-3 shrink-0">
              <button
                onClick={() => setShowPreferences(true)}
                className="btn-secondary text-white border-white/10 hover:bg-white/5 py-2 px-4 text-xs"
              >
                {t('pages.cookies.customize', 'CUSTOMIZE')}
              </button>
              <button
                onClick={handleDeclineAll}
                className="btn-secondary text-white/60 border-white/10 hover:bg-white/5 py-2 px-4 text-xs"
              >
                {t('pages.cookies.decline_all', 'DECLINE ALL')}
              </button>
              <button
                onClick={handleAcceptAll}
                className="btn-primary flex items-center justify-center min-w-[140px] relative overflow-hidden group py-2 px-6"
              >
                <span className="relative z-10 font-bold">{t('pages.cookies.accept_all', 'ACCEPT ALL')}</span>
                <div className="absolute inset-0 bg-white/20 translate-x-[-100%] group-hover:translate-x-[100%] transition-transform duration-500 ease-in-out" />
              </button>
            </div>
          </div>
        ) : (
          <div className="animate-fade-up">
            <div className="flex items-center justify-between mb-6">
              <h3 className="text-text-on-header font-display tracking-widest text-xl m-0">
                {t('pages.cookies.preferences', 'COOKIE PREFERENCES')}
              </h3>
              <button 
                onClick={() => setShowPreferences(false)}
                className="text-text-on-header/50 hover:text-white transition-colors"
              >
                ✕
              </button>
            </div>

            <div className="space-y-4 mb-8">
              {/* Essential */}
              <div className="flex items-start justify-between p-4 bg-white/5 rounded-xl border border-white/5">
                <div className="pr-4">
                  <div className="text-text-on-header font-bold text-sm mb-1 uppercase tracking-wider">{t('pages.cookies.essential_title', 'NECESSARY')}</div>
                  <div className="text-text-on-header/60 text-xs">{t('pages.cookies.essential_desc', 'Essential for the site to function (Auth, Theme, Cart).')}</div>
                </div>
                <div className="text-accent-primary text-[10px] font-mono font-bold uppercase py-2 tracking-tighter opacity-70">Always Required</div>
              </div>

              {/* Analytics */}
              <div className="flex items-start justify-between p-4 bg-white/5 rounded-xl border border-white/5 hover:border-white/10 transition-colors">
                <div className="pr-4">
                  <div className="text-text-on-header font-bold text-sm mb-1 uppercase tracking-wider">{t('pages.cookies.analytics_title', 'ANALYTICS')}</div>
                  <div className="text-text-on-header/60 text-xs">{t('pages.cookies.analytics_desc', 'Help us understand how people use the store to improve user experience.')}</div>
                </div>
                <div 
                  onClick={() => setPrefs({...prefs, analytics: !prefs.analytics})}
                  className={`w-12 h-6 rounded-full relative cursor-pointer transition-colors duration-300 ${prefs.analytics ? 'bg-accent-primary' : 'bg-white/20'}`}
                >
                  <div className={`absolute top-1 w-4 h-4 bg-white rounded-full transition-transform duration-300 ${prefs.analytics ? 'translate-x-7' : 'translate-x-1'}`} />
                </div>
              </div>

              {/* Marketing */}
              <div className="flex items-start justify-between p-4 bg-white/5 rounded-xl border border-white/5 hover:border-white/10 transition-colors">
                <div className="pr-4">
                  <div className="text-text-on-header font-bold text-sm mb-1 uppercase tracking-wider">{t('pages.cookies.marketing_title', 'MARKETING')}</div>
                  <div className="text-text-on-header/60 text-xs">{t('pages.cookies.marketing_desc', 'Used for personalized advertising and social features.')}</div>
                </div>
                <div 
                  onClick={() => setPrefs({...prefs, marketing: !prefs.marketing})}
                  className={`w-12 h-6 rounded-full relative cursor-pointer transition-colors duration-300 ${prefs.marketing ? 'bg-accent-primary' : 'bg-white/20'}`}
                >
                  <div className={`absolute top-1 w-4 h-4 bg-white rounded-full transition-transform duration-300 ${prefs.marketing ? 'translate-x-7' : 'translate-x-1'}`} />
                </div>
              </div>
            </div>

            <div className="flex justify-end gap-4">
               <button
                onClick={() => setShowPreferences(false)}
                className="btn-secondary text-white/60 border-white/10 hover:bg-white/5"
              >
                {t('pages.common.back', 'BACK')}
              </button>
              <button
                onClick={handleSavePreferences}
                className="btn-primary min-w-[200px] font-bold"
              >
                {t('pages.cookies.save', 'SAVE MY CHOICES')}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default CookieBanner;
