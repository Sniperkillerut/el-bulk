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
          --bg-page: ${t.bg_page || '#e6dac3'};
          --bg-header: ${t.bg_header || '#1a1f2e'};
          --bg-surface: ${t.bg_surface || '#fdfbf7'};
          --bg-card: ${t.bg_card || '#ffffff'};
          --text-main: ${t.text_main || '#3b3127'};
          --text-secondary: ${t.text_secondary || '#5c4e4d'};
          --text-muted: ${t.text_muted || '#8b795c'};
          --text-on-accent: ${t.text_on_accent || '#2c251d'};
          --text-on-header: ${t.text_on_header || '#ffffff'};
          --accent-primary: ${t.accent_primary || '#d4af37'};
          --accent-primary-hover: ${t.accent_primary_hover || '#b8961e'};
          --border-main: ${t.border_main || '#d4c5ab'};
          --border-focus: ${t.border_focus || '#3b3127'};
          --status-nm: ${t.status_nm || '#2e7d32'};
          --status-lp: ${t.status_lp || '#558b2f'};
          --status-mp: ${t.status_mp || '#ef6c00'};
          --status-hp: ${t.status_hp || '#c62828'};
          --status-dmg: ${t.status_dmg || '#455a44'};
          --btn-primary-bg: ${t.btn_primary_bg || (t.accent_primary || '#1a1f2e')};
          --btn-primary-text: ${t.btn_primary_text || (t.text_on_accent || '#ffffff')};
          --btn-secondary-bg: ${t.btn_secondary_bg || 'transparent'};
          --btn-secondary-text: ${t.btn_secondary_text || (t.text_main || '#3b3127')};
          --checkbox-border: ${t.checkbox_border || (t.text_muted || '#8b795c')};
          --checkbox-checked: ${t.checkbox_checked || (t.accent_primary || '#d4af37')};
          --radius-base: ${t.radius_base || '8px'};
          --padding-card: ${t.padding_card || '12px'};
          --gap-grid: ${t.gap_grid || '16px'};

          /* Legacy Aliases for Backward Compatibility */
          --ink-surface: ${t.bg_surface || '#fdfbf7'};
          --ink-card: ${t.bg_card || '#ffffff'};
          --ink-border: ${t.border_main || '#d4c5ab'};
          --ink-deep: ${t.text_main || '#3b3127'};
          --text-primary: ${t.text_main || '#3b3127'};
          --kraft-light: ${t.bg_page || '#e6dac3'};
          --kraft-paper: ${t.bg_page || '#e6dac3'};
          --gold: ${t.accent_primary || '#d4af37'};
          --gold-dark: ${t.accent_primary_hover || '#b8961e'};
          --hp-color: ${t.status_hp || '#c62828'};
          --nm-color: ${t.status_nm || '#2e7d32'};
          --lp-color: ${t.status_lp || '#558b2f'};
          --mp-color: ${t.status_mp || '#ef6c00'};
          --dmg-color: ${t.status_dmg || '#455a44'};
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
