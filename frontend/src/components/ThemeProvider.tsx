'use client';

import React from 'react';
import { ThemeProvider as NextThemesProvider } from 'next-themes';
import { type ThemeProviderProps } from 'next-themes';
import { Theme } from '@/lib/types';

interface ExtendedThemeProviderProps extends ThemeProviderProps {
  allThemes: Theme[];
}

const generateThemeCSS = (t: Theme, selector: string) => `
  ${selector} {
    --bg-page: ${t.bg_page};
    --bg-header: ${t.bg_header};
    --bg-surface: ${t.bg_surface};
    --bg-card: ${t.bg_card || t.bg_surface};
    --text-main: ${t.text_main};
    --text-secondary: ${t.text_secondary};
    --text-muted: ${t.text_muted};
    --text-on-accent: ${t.text_on_accent};
    --text-on-header: ${t.text_on_header};
    --accent-primary: ${t.accent_primary};
    --accent-primary-hover: ${t.accent_primary_hover};
    --accent-header: ${t.accent_header || t.accent_primary || '#ffffff'};
    --status-hp-header: ${t.status_hp_header || t.status_hp || '#ef4444'};
    --border-main: ${t.border_main};
    --border-focus: ${t.border_focus || t.accent_primary};
    --status-nm: ${t.status_nm};
    --status-lp: ${t.status_lp};
    --status-mp: ${t.status_mp};
    --status-hp: ${t.status_hp};
    --status-dmg: ${t.status_dmg};
    --btn-primary-bg: ${t.btn_primary_bg || t.accent_primary};
    --btn-primary-text: ${t.btn_primary_text || t.text_on_accent};
    --btn-secondary-bg: ${t.btn_secondary_bg || 'transparent'};
    --btn-secondary-text: ${t.btn_secondary_text || t.accent_primary};
    --checkbox-border: ${t.checkbox_border || t.border_main};
    --checkbox-checked: ${t.checkbox_checked || t.accent_primary};
    --radius-base: ${t.radius_base || '8px'};
    --padding-card: ${t.padding_card || '12px'};
    --gap-grid: ${t.gap_grid || '16px'};
    --bg-image-url: ${t.bg_image_url ? `url("${t.bg_image_url}")` : 'none'};
    --font-heading: ${t.font_heading || 'var(--font-heading)'};
    --font-body: ${t.font_body || 'var(--font-body)'};
    --accent-secondary: ${t.accent_secondary || 'transparent'};
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
`;

export function ThemeProvider({ children, allThemes = [], ...props }: ExtendedThemeProviderProps) {
  const defaultThemeObj = allThemes.find(t => t.id === props.defaultTheme || t.name === props.defaultTheme);
  
  const rootStyles = defaultThemeObj ? generateThemeCSS(defaultThemeObj, ':root') : '';
  const allThemesStyles = allThemes.map(t => generateThemeCSS(t, `[data-theme='${t.id}'], [data-theme='${t.name}']`)).join('\n');

  const dynamicStyles = (
    <style id="dynamic-themes" dangerouslySetInnerHTML={{
      __html: `
        ${rootStyles}
        ${allThemesStyles}
      `
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
