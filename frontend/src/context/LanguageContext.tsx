'use client';

import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { fetchTranslations } from '@/lib/api';

type Locale = string;

interface LanguageContextType {
  locale: Locale;
  setLocale: (locale: Locale) => void;
  t: (key: string, fallback?: string, params?: Record<string, string | number>) => string;
  isLoading: boolean;
  availableLocales: Locale[];
}

const LanguageContext = createContext<LanguageContextType | undefined>(undefined);

export const LanguageProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [locale, setLocaleState] = useState<Locale>('en');
  const [translations, setTranslations] = useState<Record<string, string>>({});
  const [allTranslations, setAllTranslations] = useState<Record<Locale, Record<string, string>>>({});
  const [isLoading, setIsLoading] = useState(true);

  const availableLocales = Object.keys(allTranslations).length > 0 ? Object.keys(allTranslations) : ['en', 'es'];

  const loadTranslations = useCallback(async () => {
    try {
      setIsLoading(true);
      const data = await fetchTranslations();
      // data is Record<Locale, Record<Key, Value>>
      setAllTranslations(data as Record<Locale, Record<string, string>>);
    } catch (error) {
      console.error('Failed to fetch translations:', error);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    loadTranslations();
    
    // Load preferred locale from localStorage
    const savedLocale = localStorage.getItem('el-bulk-locale');
    if (savedLocale) {
      setLocaleState(savedLocale);
    } else {
      // Try browser language
      const browserLang = navigator.language.split('-')[0];
      if (['en', 'es'].includes(browserLang)) {
        setLocaleState(browserLang);
      }
    }
  }, [loadTranslations]);

  useEffect(() => {
    if (allTranslations[locale]) {
      setTranslations(allTranslations[locale]);
    } else if (allTranslations['en']) {
      // Fallback to English if current locale not found
      setTranslations(allTranslations['en']);
    }
  }, [locale, allTranslations]);

  const setLocale = (newLocale: Locale) => {
    setLocaleState(newLocale);
    localStorage.setItem('el-bulk-locale', newLocale);
  };

  const t = useCallback((key: string, fallback?: string, params?: Record<string, string | number>): string => {
    let result = key;

    if (translations[key]) {
      result = translations[key];
    } else if (locale !== 'en' && allTranslations['en'] && allTranslations['en'][key]) {
      // If not found in current translations, try English specifically
      result = allTranslations['en'][key];
    } else {
      result = fallback || key;
    }

    // Handle interpolation: replace {key} with value from params
    if (params) {
      Object.entries(params).forEach(([k, v]) => {
        result = result.replace(new RegExp(`\\{${k}\\}`, 'g'), String(v));
      });
    }

    return result;
  }, [translations, locale, allTranslations]);

  return (
    <LanguageContext.Provider value={{ locale, setLocale, t, isLoading, availableLocales }}>
      {children}
    </LanguageContext.Provider>
  );
};

export const useLanguage = () => {
  const context = useContext(LanguageContext);
  if (context === undefined) {
    throw new Error('useLanguage must be used within a LanguageProvider');
  }
  return context;
};
