'use client';

import React, { useEffect, useState } from 'react';
import { ThemeProvider as NextThemesProvider } from 'next-themes';
import { type ThemeProviderProps } from 'next-themes';
import { fetchThemes } from '@/lib/api_themes';
import { Theme } from '@/lib/types';

export function ThemeProvider({ children, ...props }: ThemeProviderProps) {
  const [themes, setThemes] = useState<Theme[]>([]);

  useEffect(() => {
    fetchThemes().then(setThemes).catch(console.error);
  }, []);

  const dynamicStyles = (
    <style id="dynamic-themes" dangerouslySetInnerHTML={{
      __html: themes.map(t => `
        [data-theme='${t.id}'] {
          --bg-page: ${t.bg_page};
          --bg-header: ${t.bg_header};
          --bg-surface: ${t.bg_surface};
          --text-main: ${t.text_main};
          --text-secondary: ${t.text_secondary};
          --text-muted: ${t.text_muted};
          --text-on-accent: ${t.text_on_accent};
          --accent-primary: ${t.accent_primary};
          --accent-primary-hover: ${t.accent_primary_hover};
          --border-main: ${t.border_main};
          --status-nm: ${t.status_nm};
          --status-lp: ${t.status_lp};
          --status-mp: ${t.status_mp};
          --status-hp: ${t.status_hp};
          --radius-base: ${t.radius_base};
          --padding-card: ${t.padding_card};
          --gap-grid: ${t.gap_grid};
        }
      `).join('\n')
    }} />
  );

  return (
    <NextThemesProvider 
      attribute="data-theme" 
      defaultTheme="00000000-0000-0000-0000-000000000001" 
      storageKey="el-bulk-theme"
      {...props}
    >
      {dynamicStyles}
      {children}
    </NextThemesProvider>
  );
}
