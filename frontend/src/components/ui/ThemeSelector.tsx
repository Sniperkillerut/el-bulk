'use client';

import React, { useEffect, useState } from 'react';
import { useTheme } from 'next-themes';
import { fetchThemes } from '@/lib/api_themes';
import { Theme } from '@/lib/types';
import Dropdown from './Dropdown';
import { useLanguage } from '@/context/LanguageContext';

export default function ThemeSelector() {
  const { theme: activeTheme, setTheme } = useTheme();
  const { t } = useLanguage();
  const [themes, setThemes] = useState<Theme[]>([]);
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    fetchThemes().then(res => {
      setThemes(res || []);
      setMounted(true);
    }).catch(console.error);
  }, []);

  if (!mounted) return null;

  const activeThemeName = (themes || []).find(t => t.id === activeTheme)?.name || 'Theme';

  return (
    <Dropdown
      trigger={
        <button className="flex items-center gap-2 px-3 py-1.5 rounded-md hover:bg-white/5 transition-colors text-xs font-medium uppercase tracking-wider text-text-on-header/80 hover:text-text-on-header group">
          <svg 
            width="16" 
            height="16" 
            viewBox="0 0 24 24" 
            fill="none" 
            stroke="currentColor" 
            strokeWidth="2" 
            strokeLinecap="round" 
            strokeLinejoin="round" 
            className="text-accent-header"
          >
            <circle cx="13.5" cy="6.5" r=".5" fill="currentColor"/>
            <circle cx="17.5" cy="10.5" r=".5" fill="currentColor"/>
            <circle cx="8.5" cy="7.5" r=".5" fill="currentColor"/>
            <circle cx="6.5" cy="12.5" r=".5" fill="currentColor"/>
            <path d="M12 2C6.5 2 2 6.5 2 12s4.5 10 10 10c.926 0 1.707-.484 2.103-1.206.35-.637.303-1.387-.215-2.037a2.4 2.4 0 0 1 .592-3.366c.647-.49 1.551-.577 2.22-.214.72.392 1.205 1.173 1.205 2.103 0 3.145 2.554 5.7 5.7 5.7h.393a10 10 0 0 0 0-20z"/>
          </svg>
          <span className="hidden sm:inline">{activeThemeName}</span>
        </button>
      }
      align="end"
      data-theme-area="theme-selector"
    >
      <div className="py-1 w-48 max-h-64 overflow-y-auto bg-bg-surface border border-border-main shadow-xl">
        <div className="px-3 py-2 text-[10px] uppercase tracking-widest text-text-muted border-b border-border-main/50 mb-1">
          {t('components.theme_selector.title', 'Select Theme')}
        </div>
        {themes.map((t) => (
          <button
            key={t.id}
            onClick={() => setTheme(t.id)}
            className={`
              w-full text-left px-4 py-2.5 text-xs transition-colors flex items-center justify-between
              ${activeTheme === t.id 
                ? 'bg-accent-primary/10 text-accent-primary font-bold' 
                : 'text-text-secondary hover:bg-white/5 hover:text-text-main'}
            `}
          >
            {t.name}
            <div 
              className="w-3 h-3 rounded-full border border-white/10" 
              style={{ backgroundColor: t.accent_primary }} 
            />
          </button>
        ))}
      </div>
    </Dropdown>
  );
}
