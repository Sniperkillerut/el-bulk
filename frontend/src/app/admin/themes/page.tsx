'use client';

import React, { useEffect, useState } from 'react';
import { useAdmin } from '@/hooks/useAdmin';
import { 
  fetchThemes, createTheme, updateTheme, deleteTheme, 
} from '@/lib/api_themes';
import { getAdminSettings, updateAdminSettings } from '@/lib/api';
import { Theme, ThemeInput, Settings } from '@/lib/types';

// Simple SVG Icons to replace lucide-react
const Icons = {
  Plus: () => <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>,
  Save: () => <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v13a2 2 0 0 1-2 2z"></path><polyline points="17 21 17 13 7 13 7 21"></polyline><polyline points="7 3 7 8 15 8"></polyline></svg>,
  Trash: () => <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path><line x1="10" y1="11" x2="10" y2="17"></line><line x1="14" y1="11" x2="14" y2="17"></line></svg>,
  Refresh: () => <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polyline points="23 4 23 10 17 10"></polyline><polyline points="1 20 1 14 7 14"></polyline><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path></svg>,
  Check: () => <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg>,
  Palette: () => <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="13.5" cy="6.5" r=".5" fill="currentColor"></circle><circle cx="17.5" cy="10.5" r=".5" fill="currentColor"></circle><circle cx="8.5" cy="7.5" r=".5" fill="currentColor"></circle><circle cx="6.5" cy="12.5" r=".5" fill="currentColor"></circle><path d="M12 2C6.5 2 2 6.5 2 12s4.5 10 10 10c.926 0 1.707-.484 2.103-1.206.35-.637.303-1.387-.215-2.037a2.4 2.4 0 0 1 .592-3.366c.647-.49 1.551-.577 2.22-.214.72.392 1.205 1.173 1.205 2.103 0 3.145 2.554 5.7 5.7 5.7h.393a10 10 0 0 0 0-20z"></path></svg>,
  Shield: () => <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path></svg>,
  Alert: () => <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path><line x1="12" y1="9" x2="12" y2="13"></line><line x1="12" y1="17" x2="12.01" y2="17"></line></svg>,
  Grid: () => <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="3" y="3" width="7" height="7"></rect><rect x="14" y="3" width="7" height="7"></rect><rect x="14" y="14" width="7" height="7"></rect><rect x="3" y="14" width="7" height="7"></rect></svg>,
  Type: () => <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polyline points="4 7 4 4 20 4 20 7"></polyline><line x1="9" y1="20" x2="15" y2="20"></line><line x1="12" y1="4" x2="12" y2="20"></line></svg>,
  Radius: () => <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path></svg>,
  Maximize: () => <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M15 3h6v6"></path><path d="M9 21H3v-6"></path><path d="M21 3l-7 7"></path><path d="M3 21l7-7"></path></svg>,
  Chevron: ({ open }: { open: boolean }) => <svg className={`transition-transform duration-200 ${open ? 'rotate-180' : ''}`} width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><polyline points="6 9 12 15 18 9"></polyline></svg>,
  Monitor: () => <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect><line x1="8" y1="21" x2="16" y2="21"></line><line x1="12" y1="17" x2="12" y2="21"></line></svg>,
  Smartphone: () => <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="5" y="2" width="14" height="20" rx="2" ry="2"></rect><line x1="12" y1="18" x2="12" y2="18"></line></svg>,
};

const AVAILABLE_FONTS = [
  { name: 'Default', value: '' },
  { name: 'Inter (Sans)', value: 'Inter, sans-serif' },
  { name: 'Bebas Neue (Display)', value: 'Bebas Neue, sans-serif' },
  { name: 'Space Mono (Mono)', value: 'Space Mono, monospace' },
  { name: 'Cinzel (Classical)', value: 'Cinzel, serif' },
  { name: 'Playfair Display (Elegant)', value: 'Playfair Display, serif' },
  { name: 'Outfit (Modern)', value: 'Outfit, sans-serif' },
  { name: 'Roboto (Standard)', value: 'Roboto, sans-serif' },
  { name: 'Montserrat (Bold)', value: 'Montserrat, sans-serif' },
];

function FontSelector({ label, value, onChange, helperText }: { label: string, value: string, onChange: (val: string) => void, helperText?: string }) {
  return (
    <div className="flex flex-col p-2.5 rounded bg-bg-header/30 hover:bg-bg-header/50 transition-colors border border-border-main/50 space-y-1">
      <div className="flex items-center justify-between">
        <span className="text-[10px] font-bold text-text-secondary uppercase">{label}</span>
        {helperText && <span className="text-[8px] font-mono text-text-muted italic">{helperText}</span>}
      </div>
      <select 
        value={value} 
        onChange={e => onChange(e.target.value)}
        className="bg-bg-page border border-border-main/30 rounded px-2 py-1.5 text-[11px] outline-none focus:border-accent-primary transition-colors text-text-main font-medium"
        style={{ fontFamily: value || 'inherit' }}
      >
        {AVAILABLE_FONTS.map(f => (
          <option key={f.value} value={f.value} style={{ fontFamily: f.value || 'inherit' }}>
            {f.name}
          </option>
        ))}
      </select>
    </div>
  );
}

function ImageUploadInput({ label, value, onChange, helperText }: { label: string, value: string, onChange: (val: string) => void, helperText?: string }) {
  const [isUploading, setIsUploading] = useState(false);

  const handleUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (file.size > 5 * 1024 * 1024) {
      alert('File too large (Max 5MB)');
      return;
    }

    const formData = new FormData();
    formData.append('file', file);

    setIsUploading(true);
    try {
      const res = await fetch('/api/admin/upload', {
        method: 'POST',
        body: formData,
      });

      if (!res.ok) {
        const error = await res.text();
        throw new Error(error || 'Upload failed');
      }

      const data = await res.json();
      onChange(data.url);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Unknown error';
      console.error('Upload error:', err);
      alert(`Failed to upload image: ${message}`);
    } finally {
      setIsUploading(false);
      e.target.value = '';
    }
  };

  return (
    <div className="flex flex-col p-2.5 rounded bg-bg-header/30 hover:bg-bg-header/50 transition-colors border border-border-main/50 space-y-1">
      <div className="flex items-center justify-between">
        <span className="text-[10px] font-bold text-text-secondary uppercase">{label}</span>
        {helperText && <span className="text-[8px] font-mono text-text-muted italic">{helperText}</span>}
      </div>
      <div className="flex gap-2">
        <input 
          type="text" 
          value={value} 
          onChange={e => onChange(e.target.value)}
          placeholder="https://..."
          className="bg-bg-page border border-border-main/30 rounded px-2 py-1.5 text-[11px] outline-none focus:border-accent-primary transition-colors text-text-main font-medium flex-1"
        />
        <label className={`shrink-0 w-8 h-8 flex items-center justify-center rounded border transition-all cursor-pointer ${isUploading ? 'bg-bg-header animate-pulse' : 'bg-bg-page border-border-main/30 hover:border-accent-primary hover:text-accent-primary'}`}>
          {isUploading ? (
            <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none"></circle>
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
          ) : (
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
              <polyline points="17 8 12 3 7 8"></polyline>
              <line x1="12" y1="3" x2="12" y2="15"></line>
            </svg>
          )}
          <input type="file" className="hidden" accept="image/*" onChange={handleUpload} disabled={isUploading} />
        </label>
      </div>
    </div>
  );
}

const CARDBOARD_DEFAULT: ThemeInput = {
  name: 'Cardboard',
  bg_page: '#e6dac3',
  bg_header: '#1a1f2e',
  bg_surface: '#fdfbf7',
  bg_card: '#ffffff',
  text_main: '#3b3127',
  text_secondary: '#5c4e3d',
  text_muted: '#8b795c',
  text_on_accent: '#2c251d',
  text_on_header: '#ffffff',
  accent_primary: '#d4af37',
  accent_primary_hover: '#b8961e',
  border_main: '#d4c5ab',
  border_focus: '#3b3127',
  status_nm: '#2e7d32',
  status_lp: '#558b2f',
  status_mp: '#ef6c00',
  status_hp: '#c62828',
  status_dmg: '#455a64',
  accent_header: '#fbbf24',
  status_hp_header: '#f87171',
  btn_primary_bg: '#1a1f2e',
  btn_primary_text: '#ffffff',
  btn_secondary_bg: 'transparent',
  btn_secondary_text: '#3b3127',
  checkbox_border: '#8b795c',
  checkbox_checked: '#d4af37',
  radius_base: '8px',
  padding_card: '12px',
  gap_grid: '16px',
  bg_image_url: '',
  font_heading: '',
  font_body: '',
  accent_secondary: '',
};

export default function AdminThemesPage() {
  const { token } = useAdmin();
  const [themes, setThemes] = useState<Theme[]>([]);
  const [settings, setSettings] = useState<Settings | null>(null);
  const [loading, setLoading] = useState(true);
  const [editingTheme, setEditingTheme] = useState<Theme | null>(null);
  const [form, setForm] = useState<ThemeInput>(CARDBOARD_DEFAULT);
  const [isNew, setIsNew] = useState(false);
  const [message, setMessage] = useState<{ text: string, type: 'success' | 'error' } | null>(null);
  const [expanded, setExpanded] = useState<Record<string, boolean>>({
    identity: true,
    surface: true,
    signature: true,
    interactive: true,
    type: false,
    logic: false,
    geometry: false,
    advanced: true
  });
  const [previewMode, setPreviewMode] = useState<'desktop' | 'mobile'>('desktop');

  useEffect(() => {
    async function loadData() {
      setLoading(true);
      try {
        const [tData, sData] = await Promise.all([
          fetchThemes(),
          getAdminSettings()
        ]);
        setThemes(tData);
        setSettings(sData);
      } catch (err) {
        console.error(err);
      } finally {
        setLoading(false);
      }
    }

    if (token) {
      loadData();
    }
  }, [token]);

  const handleEdit = (theme: Theme) => {
    setEditingTheme(theme);
    setForm({
      name: theme.name,
      bg_page: theme.bg_page,
      bg_header: theme.bg_header,
      bg_surface: theme.bg_surface,
      bg_card: theme.bg_card,
      text_main: theme.text_main,
      text_secondary: theme.text_secondary,
      text_muted: theme.text_muted,
      text_on_accent: theme.text_on_accent,
      text_on_header: theme.text_on_header,
      accent_primary: theme.accent_primary,
      accent_primary_hover: theme.accent_primary_hover,
      border_main: theme.border_main,
      border_focus: theme.border_focus,
      status_nm: theme.status_nm,
      status_lp: theme.status_lp,
      status_mp: theme.status_mp,
      status_hp: theme.status_hp,
      status_dmg: theme.status_dmg,
      btn_primary_bg: theme.btn_primary_bg,
      btn_primary_text: theme.btn_primary_text,
      btn_secondary_bg: theme.btn_secondary_bg,
      btn_secondary_text: theme.btn_secondary_text,
      checkbox_border: theme.checkbox_border,
      checkbox_checked: theme.checkbox_checked,
      radius_base: theme.radius_base,
      padding_card: theme.padding_card,
      gap_grid: theme.gap_grid,
      bg_image_url: theme.bg_image_url || '',
      font_heading: theme.font_heading || '',
      font_body: theme.font_body || '',
      accent_header: theme.accent_header || '',
      status_hp_header: theme.status_hp_header || '',
      accent_secondary: theme.accent_secondary || '',
    });
    setIsNew(false);
    setMessage(null);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  const handleCreateNew = () => {
    setEditingTheme(null);
    setForm({ ...CARDBOARD_DEFAULT, name: 'New Theme' });
    setIsNew(true);
    setMessage(null);
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      if (isNew) {
        await createTheme(form);
        setMessage({ text: 'Theme created successfully!', type: 'success' });
      } else if (editingTheme) {
        await updateTheme(editingTheme.id, form);
        setMessage({ text: 'Theme updated successfully!', type: 'success' });
      }
      const tData = await fetchThemes();
      setThemes(tData);
    } catch (err: unknown) {
      const error = err as Error;
      setMessage({ text: error.message || 'Operation failed', type: 'error' });
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this theme?')) return;
    setLoading(true);
    try {
      await deleteTheme(id);
      setMessage({ text: 'Theme deleted', type: 'success' });
      const tData = await fetchThemes();
      setThemes(tData);
      if (editingTheme?.id === id) handleCreateNew();
    } catch (err: unknown) {
      const error = err as Error;
      setMessage({ text: error.message, type: 'error' });
    } finally {
      setLoading(false);
    }
  };

  const handleSetDefault = async (id: string) => {
    setLoading(true);
    try {
      await updateAdminSettings({ default_theme_id: id });
      setMessage({ text: 'Default theme updated', type: 'success' });
      const sData = await getAdminSettings();
      setSettings(sData);
    } catch (err: unknown) {
      const error = err as Error;
      setMessage({ text: error.message, type: 'error' });
    } finally {
      setLoading(false);
    }
  };

  const handleResetToSystem = () => {
    if (editingTheme?.is_system) {
      setForm(CARDBOARD_DEFAULT);
      setMessage({ text: 'Reset to system default values (not saved yet)', type: 'success' });
    }
  };

  if (loading && themes.length === 0) {
    return <div className="p-8 text-text-muted animate-pulse font-mono uppercase tracking-widest">Initializing visual matrix...</div>;
  }

  return (
    <div className="flex-1 overflow-y-auto p-4 lg:p-8 bg-bg-page text-text-main custom-scrollbar">
      <div className="max-w-[1600px] mx-auto space-y-8">
        
        <header className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3 border-b border-border-main pb-6">
          <div>
            <h1 className="text-3xl font-display font-bold text-accent-primary">THEME ENGINE</h1>
            <p className="text-text-secondary text-sm font-mono mt-1 opacity-60 uppercase tracking-wider">Configure visual identity and storefront skins</p>
          </div>
          <button 
            onClick={handleCreateNew}
            className="flex items-center gap-2 px-4 py-2 bg-accent-primary text-text-on-accent rounded shadow-lg hover:brightness-110 active:scale-95 transition-all font-bold uppercase text-xs tracking-widest"
          >
            <Icons.Plus />
            Forge New Theme
          </button>
        </header>

        {message && (
          <div className={`p-4 rounded border flex items-center gap-3 animate-in fade-in slide-in-from-top-2 overflow-hidden relative ${
            message.type === 'success' ? 'bg-emerald-500/10 border-emerald-500/30 text-emerald-400' : 'bg-hp-color/10 border-hp-color/30 text-hp-color'
          }`}>
            <div className={`absolute top-0 left-0 w-1 h-full ${message.type === 'success' ? 'bg-emerald-500' : 'bg-hp-color'}`} />
            {message.type === 'success' ? <Icons.Check /> : <Icons.Alert />}
            <span className="font-medium">{message.text}</span>
          </div>
        )}

        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
          
          {/* 1. Left Sidebar: Available Skins (Compact) */}
          <div className="lg:col-span-2 space-y-4 lg:sticky lg:top-8">
            <h2 className="text-[10px] font-mono font-bold text-text-muted uppercase tracking-[0.2em] flex items-center gap-2 mb-2">
              <Icons.Grid /> Skins
            </h2>
            <div className="flex lg:grid gap-2 overflow-x-auto lg:overflow-visible no-scrollbar pb-2 lg:pb-0">
              {themes.map(t => (
                <div 
                  key={t.id}
                  onClick={() => handleEdit(t)}
                  className={`
                    p-3 rounded border cursor-pointer transition-all relative group flex flex-col gap-2 min-w-[140px] lg:min-w-0
                    ${editingTheme?.id === t.id ? 'border-accent-primary bg-accent-primary/5 shadow-sm' : 'border-border-main hover:border-text-muted bg-bg-surface'}
                  `}
                >
                  <div className="flex items-center justify-between">
                    <span className="font-bold text-[11px] truncate pr-4">{t.name}</span>
                    {t.is_system && <Icons.Shield />}
                  </div>
                  
                  <div className="flex gap-1">
                    <div className="w-2.5 h-2.5 rounded-full border border-white/5" style={{ backgroundColor: t.bg_page }} />
                    <div className="w-2.5 h-2.5 rounded-full border border-white/5" style={{ backgroundColor: t.accent_primary }} />
                    <div className="w-2.5 h-2.5 rounded-full border border-white/5" style={{ backgroundColor: t.text_main }} />
                  </div>

                  {settings?.default_theme_id === t.id && (
                    <div className="absolute top-1 right-1 w-1.5 h-1.5 bg-accent-primary rounded-full shadow-[0_0_8px_rgba(212,175,55,0.8)]" />
                  )}

                  {/* Hover Actions */}
                  <div className="absolute inset-0 bg-bg-surface/90 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2 rounded">
                    <button 
                      onClick={(e) => { e.stopPropagation(); handleSetDefault(t.id); }}
                      className="p-1.5 text-accent-primary hover:bg-accent-primary/10 rounded transition-colors"
                      title="Set as Default"
                    >
                      <Icons.Check />
                    </button>
                    {!t.is_system && (
                      <button 
                        onClick={(e) => { e.stopPropagation(); handleDelete(t.id); }}
                        className="p-1.5 text-hp-color hover:bg-hp-color/10 rounded transition-colors"
                        title="Delete"
                      >
                        <Icons.Trash />
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* 2. Center Stage: Large Sticky Preview */}
          <div className="lg:col-span-7 lg:sticky lg:top-8 space-y-4 hidden lg:block">
            <div className="flex items-center justify-between border-b border-border-main/20 pb-4">
              <div className="flex items-center gap-3">
                <div className="flex items-center gap-1.5 p-1 bg-bg-header/20 rounded-md border border-border-main/30">
                  <span className={`w-2 h-2 rounded-full ${form.accent_primary && 'bg-accent-primary animate-pulse'}`} style={{ backgroundColor: form.accent_primary }} />
                  <span className="text-[10px] font-mono font-bold tracking-tight opacity-80">LIVE MATRIX</span>
                </div>
                
                <div className="flex items-center bg-bg-header/40 rounded-lg p-0.5 border border-border-main/50">
                  <button 
                    type="button"
                    onClick={() => setPreviewMode('desktop')}
                    className={`p-1.5 rounded-md transition-all ${previewMode === 'desktop' ? 'bg-accent-primary text-text-on-accent shadow-sm' : 'text-text-muted hover:text-text-main'}`}
                    title="Desktop View"
                  >
                    <Icons.Monitor />
                  </button>
                  <button 
                    type="button"
                    onClick={() => setPreviewMode('mobile')}
                    className={`p-1.5 rounded-md transition-all ${previewMode === 'mobile' ? 'bg-accent-primary text-text-on-accent shadow-sm' : 'text-text-muted hover:text-text-main'}`}
                    title="Mobile View"
                  >
                    <Icons.Smartphone />
                  </button>
                </div>
              </div>
              <div className="hidden lg:flex items-center gap-2 px-3 py-1 bg-accent-primary/10 rounded-full border border-accent-primary/20">
                <span className="text-[10px] font-bold text-accent-primary uppercase tracking-wider">High Fidelity</span>
              </div>
            </div>

            <div className="flex-1 min-h-0 flex items-start justify-center p-8 overflow-y-auto bg-bg-page/5 bg-[radial-gradient(circle_at_center,var(--border-main)_0.5px,transparent_0.5px)] bg-[size:16px_16px] rounded-lg border border-border-main/20">
              <MiniSinglesPreview form={form} mode={previewMode} />
            </div>
          </div>

          {/* 3. Right Sidebar: Property Editor (Single Column) */}
          <div className="lg:col-span-3">
            <form onSubmit={handleSave} className="bg-bg-surface border border-border-main rounded-lg shadow-xl overflow-hidden flex flex-col h-full max-h-[calc(100vh-12rem)]">
              <div className="p-4 border-b border-border-main bg-bg-header/50 flex items-center justify-between shrink-0">
                <div className="flex items-center gap-2">
                  <h3 className="text-xs font-display font-bold uppercase tracking-widest">{isNew ? 'Forge' : 'Edit'} Identity</h3>
                </div>
                {editingTheme?.is_system && (
                  <button 
                    type="button" 
                    onClick={handleResetToSystem}
                    title="Reset to Original"
                    className="text-text-muted hover:text-accent-primary transition-colors"
                  >
                    <Icons.Refresh />
                  </button>
                )}
              </div>

              <div className="flex-1 overflow-y-auto p-4 space-y-4 custom-scrollbar">
                {/* 1. Identity Section */}
                <Collapsible 
                  title="Base Info" 
                  icon={<Icons.Type />} 
                  isOpen={expanded.identity} 
                  onToggle={() => setExpanded(p => ({...p, identity: !p.identity}))}
                >
                  <div className="space-y-1">
                    <label className="text-[10px] text-text-secondary uppercase font-bold tracking-tighter">Theme Name</label>
                    <input 
                      type="text" 
                      value={form.name} 
                      onChange={e => setForm({...form, name: e.target.value})}
                      className="bg-bg-page border border-border-main rounded px-3 py-1.5 text-xs focus:border-accent-primary outline-none w-full shadow-inner"
                      required
                    />
                  </div>
                </Collapsible>

                {/* 2. Surface Palette */}
                <Collapsible 
                  title="Backgrounds & Surfaces" 
                  isOpen={expanded.surface} 
                  onToggle={() => setExpanded(p => ({...p, surface: !p.surface}))}
                >
                  <div className="space-y-2">
                    <ColorInput label="Global Page Background" value={form.bg_page} onChange={val => setForm({...form, bg_page: val})} />
                    <ColorInput label="Navigation Header" value={form.bg_header} onChange={val => setForm({...form, bg_header: val})} />
                    <ColorInput label="Card & Panel Surface" value={form.bg_surface} onChange={val => setForm({...form, bg_surface: val})} />
                    <ColorInput label="Inner Card Background" value={form.bg_card} onChange={val => setForm({...form, bg_card: val})} />
                    <ColorInput label="Borders & Dividers" value={form.border_main} onChange={val => setForm({...form, border_main: val})} />
                  </div>
                </Collapsible>

                {/* 3. Brand Palette */}
                <Collapsible 
                  title="Brand & Highlights" 
                  isOpen={expanded.signature} 
                  onToggle={() => setExpanded(p => ({...p, signature: !p.signature}))}
                >
                  <div className="space-y-2">
                    <ColorInput label="Primary Accent Color" value={form.accent_primary} onChange={val => setForm({...form, accent_primary: val})} />
                    <ColorInput label="Hover / Secondary Accent" value={form.accent_primary_hover} onChange={val => setForm({...form, accent_primary_hover: val})} />
                    <ColorInput label="Text on Accent Backgrounds" value={form.text_on_accent} onChange={val => setForm({...form, text_on_accent: val})} />
                    <ColorInput label="Text on Navigation Header" value={form.text_on_header} onChange={val => setForm({...form, text_on_header: val})} />
                    <ColorInput label="Accent for Header Areas" value={form.accent_header} onChange={val => setForm({...form, accent_header: val})} />
                    <ColorInput label="High Contrast HP (Header)" value={form.status_hp_header} onChange={val => setForm({...form, status_hp_header: val})} />
                    <ColorInput label="Focus State Color" value={form.border_focus} onChange={val => setForm({...form, border_focus: val})} />
                  </div>
                </Collapsible>

                <Collapsible 
                  title="Buttons & Inputs" 
                  isOpen={expanded.interactive} 
                  onToggle={() => setExpanded(p => ({...p, interactive: !p.interactive}))}
                >
                  <div className="space-y-4">
                    <div className="space-y-2">
                       <p className="text-[9px] font-bold text-text-muted uppercase tracking-widest pl-1">Primary Button</p>
                       <ColorInput label="Button Background" value={form.btn_primary_bg} onChange={val => setForm({...form, btn_primary_bg: val})} />
                       <ColorInput label="Button Text" value={form.btn_primary_text} onChange={val => setForm({...form, btn_primary_text: val})} />
                    </div>
                    <div className="space-y-2">
                       <p className="text-[9px] font-bold text-text-muted uppercase tracking-widest pl-1">Secondary Button</p>
                       <ColorInput label="Button Background" value={form.btn_secondary_bg} onChange={val => setForm({...form, btn_secondary_bg: val})} />
                       <ColorInput label="Button Text" value={form.btn_secondary_text} onChange={val => setForm({...form, btn_secondary_text: val})} />
                    </div>
                    <div className="space-y-2">
                       <p className="text-[9px] font-bold text-text-muted uppercase tracking-widest pl-1">Checkboxes</p>
                       <ColorInput label="Checkbox Border" value={form.checkbox_border} onChange={val => setForm({...form, checkbox_border: val})} />
                       <ColorInput label="Checkbox Checked BG" value={form.checkbox_checked} onChange={val => setForm({...form, checkbox_checked: val})} />
                    </div>
                  </div>
                </Collapsible>

                {/* 4. Type Palette */}
                <Collapsible 
                  title="Typography & Content" 
                  isOpen={expanded.type} 
                  onToggle={() => setExpanded(p => ({...p, type: !p.type}))}
                >
                  <div className="space-y-2">
                    <ColorInput label="Primary Heading Text" value={form.text_main} onChange={val => setForm({...form, text_main: val})} />
                    <ColorInput label="Secondary / Body Text" value={form.text_secondary} onChange={val => setForm({...form, text_secondary: val})} />
                    <ColorInput label="Muted / Small Text" value={form.text_muted} onChange={val => setForm({...form, text_muted: val})} />
                  </div>
                </Collapsible>

                {/* 5. Logic Palette */}
                <Collapsible 
                  title="Condition Status Indicators" 
                  isOpen={expanded.logic} 
                  onToggle={() => setExpanded(p => ({...p, logic: !p.logic}))}
                >
                  <div className="space-y-2">
                    <ColorInput label="Near Mint (NM) Status" value={form.status_nm} onChange={val => setForm({...form, status_nm: val})} />
                    <ColorInput label="Lightly Played (LP) Status" value={form.status_lp} onChange={val => setForm({...form, status_lp: val})} />
                    <ColorInput label="Moderately Played (MP) Status" value={form.status_mp} onChange={val => setForm({...form, status_mp: val})} />
                    <ColorInput label="Heavily Played (HP) Status" value={form.status_hp} onChange={val => setForm({...form, status_hp: val})} />
                    <ColorInput label="Damaged (DMG) Status" value={form.status_dmg} onChange={val => setForm({...form, status_dmg: val})} />
                  </div>
                </Collapsible>

                {/* 6. Layout Geometry */}
                <Collapsible 
                  title="Layout & Card Geometry" 
                  icon={<Icons.Maximize />} 
                  isOpen={expanded.geometry} 
                  onToggle={() => setExpanded(p => ({...p, geometry: !p.geometry}))}
                >
                  <div className="space-y-2">
                    <LayoutPropertyInput label="Card Corner Rounding" value={form.radius_base} onChange={val => setForm({...form, radius_base: val})} helperText="(px/rem)" />
                    <LayoutPropertyInput label="Internal Card Padding" value={form.padding_card} onChange={val => setForm({...form, padding_card: val})} helperText="(px/rem)" />
                    <LayoutPropertyInput label="Grid Spacing (Gap)" value={form.gap_grid} onChange={val => setForm({...form, gap_grid: val})} helperText="(px/rem)" />
                  </div>
                </Collapsible>

                {/* 7. Advanced Extensions */}
                <Collapsible 
                  title="Advanced Branding & Overrides" 
                  icon={<Icons.Palette />} 
                  isOpen={expanded.advanced} 
                  onToggle={() => setExpanded(p => ({...p, advanced: !p.advanced}))}
                >
                  <div className="space-y-2">
                    <ImageUploadInput label="Background Overlay Image URL" value={form.bg_image_url || ''} onChange={val => setForm({...form, bg_image_url: val})} helperText="(Can be SVG data URI or absolute URL)" />
                    <ColorInput label="Secondary Edge/Accent Color" value={form.accent_secondary || ''} onChange={val => setForm({...form, accent_secondary: val})} />
                    <FontSelector label="Header Typography Family" value={form.font_heading || ''} onChange={val => setForm({...form, font_heading: val})} helperText="(Logos and Headings)" />
                    <FontSelector label="Body Typography Family" value={form.font_body || ''} onChange={val => setForm({...form, font_body: val})} helperText="(UI and Content)" />
                  </div>
                </Collapsible>
              </div>

              <div className="p-4 border-t border-border-main bg-bg-header/20 flex flex-col gap-3 shrink-0">
                <button 
                  type="submit" 
                  disabled={loading}
                  className="w-full flex items-center justify-center gap-2 py-3 bg-accent-primary text-text-on-accent rounded shadow-lg hover:brightness-110 font-bold uppercase text-xs tracking-widest disabled:opacity-50 transition-all"
                >
                  <Icons.Save /> {isNew ? 'Initialize Skin' : 'Persist Blueprint'}
                </button>
                
                {!isNew && !editingTheme?.is_system && (
                  <button 
                    type="button"
                    onClick={() => handleDelete(editingTheme!.id)}
                    className="w-full flex items-center justify-center gap-2 py-2 border border-hp-color/30 text-hp-color rounded hover:bg-hp-color/10 transition-colors uppercase text-[10px] font-bold tracking-widest"
                  >
                    <Icons.Trash /> Purge Matrix
                  </button>
                )}
              </div>
            </form>
          </div>

        </div>
      </div>
    </div>
  );
}

function Collapsible({ title, icon, isOpen, onToggle, children }: { title: string, icon?: React.ReactNode, isOpen: boolean, onToggle: () => void, children: React.ReactNode }) {
  return (
    <section className="border border-border-main/30 rounded-md overflow-hidden bg-bg-surface/30">
      <button 
        type="button"
        onClick={onToggle}
        className="w-full flex items-center justify-between p-3 bg-bg-header/5 hover:bg-bg-header/10 transition-colors"
      >
        <span className="text-[10px] font-mono font-bold text-text-main uppercase tracking-[0.15em] flex items-center gap-2">
          {icon} {title}
        </span>
        <Icons.Chevron open={isOpen} />
      </button>
      {isOpen && (
        <div className="p-4 space-y-4 border-t border-border-main/10 animate-in slide-in-from-top-1 duration-200">
          {children}
        </div>
      )}
    </section>
  );
}

function MiniSinglesPreview({ form, mode = 'desktop' }: { form: ThemeInput, mode?: 'desktop' | 'mobile' }) {
  const isMobile = mode === 'mobile';
  
  // Convert standard hex/string values to CSS variables for preview scoping
  const scopedVars = {
    // Base Variables
    '--bg-page': form.bg_page,
    '--bg-header': form.bg_header,
    '--bg-surface': form.bg_surface,
    '--bg-card': form.bg_card,
    '--text-main': form.text_main,
    '--text-secondary': form.text_secondary,
    '--text-muted': form.text_muted,
    '--text-on-accent': form.text_on_accent,
    '--text-on-header': form.text_on_header,
    '--accent-primary': form.accent_primary,
    '--accent-primary-hover': form.accent_primary_hover,
    '--border-main': form.border_main,
    '--border-focus': form.border_focus,
    '--status-nm': form.status_nm,
    '--status-lp': form.status_lp,
    '--status-mp': form.status_mp,
    '--status-hp': form.status_hp,
    '--status-dmg': form.status_dmg,
    '--accent-header': form.accent_header,
    '--status-hp-header': form.status_hp_header,
    '--btn-primary-bg': form.btn_primary_bg,
    '--btn-primary-text': form.btn_primary_text,
    '--btn-secondary-bg': form.btn_secondary_bg,
    '--btn-secondary-text': form.btn_secondary_text,
    '--checkbox-border': form.checkbox_border,
    '--checkbox-checked': form.checkbox_checked,
    '--radius-base': form.radius_base,
    '--padding-card': form.padding_card,
    '--gap-grid': form.gap_grid,
    '--theme-bg-image': form.bg_image_url ? `url(${form.bg_image_url})` : 'none',
    '--theme-font-heading': form.font_heading || 'Inherit',
    '--theme-font-body': form.font_body || 'Inherit',
    '--accent-secondary': form.accent_secondary || form.border_main,

    // Tailwind Resolved Variables (Prevents bleeding from root theme)
    '--color-bg-page': form.bg_page,
    '--color-bg-header': form.bg_header,
    '--color-bg-surface': form.bg_surface,
    '--color-bg-card': form.bg_card,
    '--color-text-main': form.text_main,
    '--color-text-secondary': form.text_secondary,
    '--color-text-muted': form.text_muted,
    '--color-text-on-accent': form.text_on_accent,
    '--color-text-on-header': form.text_on_header,
    '--color-accent-primary': form.accent_primary,
    '--color-accent-primary-hover': form.accent_primary_hover,
    '--color-border-main': form.border_main,
    '--color-border-focus': form.border_focus,
    '--color-status-nm': form.status_nm,
    '--color-status-lp': form.status_lp,
    '--color-status-mp': form.status_mp,
    '--color-status-hp': form.status_hp,
    '--color-status-dmg': form.status_dmg,
    '--color-accent-header': form.accent_header,
    '--color-status-hp-header': form.status_hp_header,
    '--color-btn-primary-bg': form.btn_primary_bg,
    '--color-btn-primary-text': form.btn_primary_text,
    '--color-btn-secondary-bg': form.btn_secondary_bg,
    '--color-btn-secondary-text': form.btn_secondary_text,
    '--color-checkbox-border': form.checkbox_border,
    '--color-checkbox-checked': form.checkbox_checked,
    '--color-accent-secondary': form.accent_secondary || form.border_main,
  } as React.CSSProperties;

  return (
    <div 
      className={`rounded-lg border shadow-2xl overflow-hidden origin-top flex flex-col transition-all duration-500 ease-in-out border-border-main scrollbar-none`} 
      style={{ 
        ...scopedVars,
        backgroundColor: 'var(--bg-page)', 
        backgroundImage: 'var(--theme-bg-image)',
        backgroundSize: 'cover',
        backgroundAttachment: 'fixed',
        fontFamily: 'var(--theme-font-body)',
        width: isMobile ? '375px' : '100%',
        minHeight: isMobile ? '667px' : '700px',
        maxHeight: isMobile ? '667px' : '800px',
      }}
    >
      {/* Header */}
      <header className="p-4 border-b border-border-main flex items-center justify-between sticky top-0 z-20 bg-bg-header text-text-on-header">
        <div className="flex items-center gap-4">
          <div className="w-8 h-8 rounded bg-accent-primary text-text-on-accent flex items-center justify-center font-display" style={{ fontFamily: 'var(--theme-font-heading)' }}>EB</div>
          {!isMobile && (
            <nav className="flex gap-4 text-[10px] font-bold uppercase tracking-widest">
              <span className="opacity-60">SINGLES</span>
              <span className="text-accent-header border-b-2 border-accent-header pb-1">EXPLORE</span>
              <span className="opacity-60">DECKS</span>
            </nav>
          )}
        </div>
        <div className="flex items-center gap-3">
           <div className="px-2 py-0.5 rounded-full text-[8px] font-bold uppercase border border-status-hp-header text-status-hp-header bg-status-hp-header/10 backdrop-blur-sm shadow-sm">Critical Alert</div>
           <div className="hidden sm:flex items-center gap-2">
             <div className="w-5 h-5 rounded-full border border-border-main bg-bg-surface flex items-center justify-center text-[8px] text-text-main shadow-inner">2</div>
           </div>
           <div className="w-7 h-7 rounded-full border-2 border-accent-header shrink-0"></div>
        </div>
      </header>

      <div className="flex-1 flex overflow-hidden">
        {/* Sidebar */}
        {!isMobile && (
          <aside className="w-56 p-4 border-r border-accent-secondary overflow-y-auto bg-bg-surface/60 backdrop-blur-sm custom-scrollbar">
            <div className="space-y-6">
              <div>
                <p className="text-[8px] font-bold uppercase tracking-[0.2em] mb-2 text-text-muted">Vault Matrix Search</p>
                <input type="search" placeholder="Search cards..." className="w-full" />
              </div>
              
              <div>
                <p className="text-[8px] font-bold uppercase tracking-[0.2em] mb-3 text-text-muted">Filter State</p>
                <div className="space-y-2.5">
                  <label className="flex items-center gap-2.5 cursor-pointer group">
                    <input type="checkbox" className="group-hover:border-accent-primary" />
                    <span className="text-[10px] text-text-secondary group-hover:text-text-main transition-colors">Normal Inactive</span>
                  </label>
                  <label className="flex items-center gap-2.5 cursor-pointer group">
                    <input type="checkbox" defaultChecked className="group-hover:border-accent-primary" />
                    <span className="text-[10px] font-bold text-text-main">Active Selected</span>
                  </label>
                </div>
              </div>

              <div>
                <p className="text-[10px] font-display font-bold uppercase text-text-main mb-3 border-b border-border-main pb-1" style={{ fontFamily: 'var(--theme-font-heading)' }}>Interactive ▼</p>
                <div className="space-y-3 flex flex-col">
                  <button className="btn-primary text-[10px] py-1.5 px-3">Primary Action</button>
                  <button className="btn-secondary text-[10px] py-1.5 px-3">Secondary Action</button>
                </div>
              </div>

              <div className="space-y-3">
                <p className="text-[8px] font-bold uppercase tracking-[0.2em] mb-3 text-text-muted">Dynamic Status Flags</p>
                <div className="flex flex-wrap gap-2">
                  <span className="border border-status-nm text-status-nm bg-status-nm/10 px-1.5 py-0.5 rounded-sm text-[8px] font-bold uppercase tracking-wider">NM</span>
                  <span className="border border-status-lp text-status-lp bg-status-lp/10 px-1.5 py-0.5 rounded-sm text-[8px] font-bold uppercase tracking-wider">LP</span>
                  <span className="border border-status-mp text-status-mp bg-status-mp/10 px-1.5 py-0.5 rounded-sm text-[8px] font-bold uppercase tracking-wider">MP</span>
                  <span className="border border-status-hp text-status-hp bg-status-hp/10 px-1.5 py-0.5 rounded-sm text-[8px] font-bold uppercase tracking-wider">HP</span>
                  <span className="border border-status-dmg text-status-dmg bg-status-dmg/10 px-1.5 py-0.5 rounded-sm text-[8px] font-bold uppercase tracking-wider">DMG</span>
                </div>
              </div>
            </div>
          </aside>
        )}

        {/* Main Content */}
        <main className="flex-1 p-6 overflow-y-auto custom-scrollbar relative bg-[radial-gradient(circle_at_center,var(--color-border-main)_0.5px,transparent_0.5px)] bg-[size:16px_16px]">
          
          <div className="mb-6 p-6 rounded-xl flex flex-col md:flex-row items-center justify-between gap-4 bg-accent-primary text-text-on-accent relative overflow-hidden group shadow-[0_10px_30px_-10px_var(--color-accent-primary)] border border-white/10" style={{ borderRadius: 'calc(var(--radius-base) * 1.5)' }}>
            <div className="absolute inset-0 opacity-10 bg-[url('https://www.transparenttextures.com/patterns/carbon-fibre.png')] pointer-events-none mix-blend-overlay" />
            <div className="relative z-10 text-center md:text-left flex-1">
              <span className="text-[8px] font-mono font-bold uppercase tracking-[0.3em] opacity-80 mb-1 block">Limited Release</span>
              <h3 className="text-2xl sm:text-3xl font-display leading-none mb-2" style={{ fontFamily: 'var(--theme-font-heading)' }}>STRIXHAVEN ARCHIVES</h3>
              <p className="text-[10px] sm:text-xs opacity-90 max-w-md">Discover the forbidden knowledge of the Multiverse with our exclusive archive singles.</p>
            </div>
            <button className="relative z-10 px-6 py-2.5 bg-text-on-accent text-accent-primary hover:scale-105 active:scale-95 border border-white/30 text-[10px] font-bold uppercase tracking-widest transition-all shadow-lg" style={{ borderRadius: 'var(--radius-base)' }}>
              View Collection
            </button>
          </div>

          <div className="mb-6 flex flex-col sm:flex-row items-start sm:items-end justify-between gap-4 border-b border-border-main pb-4">
             <div>
               <h2 className={`${isMobile ? 'text-xl' : 'text-3xl'} font-display leading-tight text-text-main group-hover:text-accent-primary transition-colors`} style={{ fontFamily: 'var(--theme-font-heading)' }}>PREVIEW ENGINE</h2>
               <p className="text-[10px] mt-1 italic text-text-secondary">The curator&apos;s choice for high-fidelity inventory management.</p>
             </div>
             <div className="flex gap-2">
                 <select className="text-xs py-1.5 w-auto pr-8">
                    <option>Sort by Date</option>
                    <option>Sort by Price</option>
                 </select>
             </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-12 gap-6 items-start">
            <div className="md:col-span-8">
              <div className="grid grid-cols-1 sm:grid-cols-2" style={{ gap: 'var(--gap-grid)' }}>
                {/* Product Card 1 */}
                <div className="card flex flex-col group p-0 border border-border-main bg-bg-card shadow-sm hover:shadow-xl transition-all duration-300" style={{ borderRadius: 'var(--radius-base)' }}>
                  <div className="aspect-[4/3] relative flex items-center justify-center p-4 border-b border-border-main/50" style={{ backgroundColor: 'var(--bg-page-dark, rgba(0,0,0,0.05))' }}>
                     <div className="w-24 h-32 bg-bg-card rounded shadow-md flex items-center justify-center border border-border-main transition-transform group-hover:scale-105 duration-500">
                        <span className="text-text-muted opacity-30 font-display text-4xl">MTG</span>
                     </div>
                     <div className="absolute top-2 left-2 flex flex-col gap-1.5">
                       <span className="border border-status-nm text-status-nm bg-bg-card/90 backdrop-blur-sm px-1.5 py-0.5 rounded-sm text-[8px] font-bold shadow-sm">NM</span>
                       <span className="border border-accent-primary text-accent-primary bg-bg-card/90 backdrop-blur-sm px-1.5 py-0.5 rounded-sm text-[8px] font-bold shadow-sm flex items-center gap-1">✨ FOIL</span>
                     </div>
                  </div>
                  <div className="flex flex-col flex-1" style={{ padding: 'var(--padding-card)' }}>
                    <h5 className="text-[13px] font-bold truncate leading-tight mb-0.5 text-text-main group-hover:text-accent-primary transition-colors" style={{ fontFamily: 'var(--theme-font-body)' }}>Black Lotus</h5>
                    <p className="text-[10px] truncate text-text-muted font-mono tracking-tight">Limited Edition Alpha</p>
                    <div className="mt-4 pt-3 border-t border-dashed border-border-main flex items-center justify-between">
                      <span className="text-[12px] font-bold text-text-main font-mono tracking-tight">$50,000,000</span>
                      <button className="btn-primary text-[10px] py-1.5 px-4 shadow-[0_2px_10px_-2px_var(--color-accent-primary)] hover:shadow-[0_4px_15px_-2px_var(--color-accent-primary)]">Buy</button>
                    </div>
                  </div>
                </div>

                {/* Product Card 2 - out of stock / sale */}
                <div className="card flex flex-col group p-0 border border-border-main bg-bg-card shadow-sm transition-all duration-300 opacity-90" style={{ borderRadius: 'var(--radius-base)' }}>
                  <div className="aspect-[4/3] bg-bg-page/40 relative flex items-center justify-center p-4 border-b border-border-main/50" style={{ backgroundColor: 'var(--bg-page-dark, rgba(0,0,0,0.05))' }}>
                     <div className="w-24 h-32 bg-bg-card opacity-50 rounded shadow-sm border flex items-center justify-center border-border-main grayscale transition-all group-hover:grayscale-0">
                        <span className="text-text-muted opacity-30 font-display text-4xl">MTG</span>
                     </div>
                     <div className="absolute top-2 left-2 flex flex-col gap-1.5">
                       <span className="border border-status-hp text-status-hp bg-bg-card/90 backdrop-blur-sm px-1.5 py-0.5 rounded-sm text-[8px] font-bold shadow-sm">HP</span>
                       <span className="border border-transparent bg-status-hp text-white px-1.5 py-0.5 rounded-sm text-[8px] font-bold shadow-sm">SALE</span>
                     </div>
                     <div className="absolute inset-0 bg-bg-page/20 flex flex-col items-center justify-center backdrop-blur-[1px] opacity-0 group-hover:opacity-100 transition-opacity">
                         <span className="bg-bg-card text-text-main border border-border-main px-3 py-1 font-bold text-[10px] uppercase shadow-lg rounded">Out of Stock</span>
                     </div>
                  </div>
                  <div className="flex flex-col flex-1" style={{ padding: 'var(--padding-card)' }}>
                    <h5 className="text-[13px] font-bold truncate leading-tight mb-0.5 text-text-main">Lightning Bolt</h5>
                    <p className="text-[10px] truncate text-text-muted font-mono tracking-tight">Fourth Edition</p>
                    <div className="mt-4 pt-3 border-t border-dashed border-border-main flex items-center justify-between">
                      <div className="flex flex-col">
                        <span className="text-[12px] font-bold text-text-main font-mono tracking-tight">$15.00</span>
                        <span className="text-[9px] text-status-hp line-through font-mono opacity-80">$21.00</span>
                      </div>
                      <button className="btn-secondary text-[10px] py-1.5 px-3 opacity-60">Waitlist</button>
                    </div>
                  </div>
                </div>
              </div>
              
              {/* Pagination Mock */}
              <div className="mt-6 pt-4 border-t border-border-main flex items-center justify-center gap-2">
                 <button className="w-8 h-8 flex items-center justify-center rounded border border-border-main bg-bg-surface text-text-secondary hover:bg-bg-card hover:border-accent-primary transition-colors text-xs font-mono">&lt;</button>
                 <button className="w-8 h-8 flex items-center justify-center rounded border border-accent-primary bg-accent-primary text-text-on-accent text-xs font-mono shadow-sm">1</button>
                 <button className="w-8 h-8 flex items-center justify-center rounded border border-border-main bg-bg-surface text-text-secondary hover:bg-bg-card hover:border-accent-primary transition-colors text-xs font-mono">2</button>
                 <button className="w-8 h-8 flex items-center justify-center rounded border border-border-main bg-bg-surface text-text-secondary hover:bg-bg-card hover:border-accent-primary transition-colors text-xs font-mono">...</button>
                 <button className="w-8 h-8 flex items-center justify-center rounded border border-border-main bg-bg-surface text-text-secondary hover:bg-bg-card hover:border-accent-primary transition-colors text-xs font-mono">&gt;</button>
              </div>
            </div>

            <div className="md:col-span-4 space-y-6">
               <div className="p-5 border border-border-main shadow-inner bg-bg-surface flex flex-col gap-3 group" style={{ borderRadius: 'var(--radius-base)' }}>
                  <p className="text-[10px] font-bold uppercase tracking-widest text-accent-primary group-hover:text-accent-primary-hover transition-colors">Newsletter Alert</p>
                  <p className="text-xs leading-relaxed text-text-secondary">Join our guild of collectors for instant stock alerts and drops.</p>
                  <input type="email" placeholder="collector@matrix.com" className="text-xs mt-1 shadow-inner focus:shadow-[0_0_0_2px_var(--color-border-focus)]" />
                  <button className="btn-primary w-full text-[11px] py-2.5 mt-1">Subscribe</button>
               </div>

               <div className="p-5 border border-border-main bg-bg-card flex flex-col gap-4 border-l-[3px] shadow-md hover:shadow-lg transition-shadow" style={{ borderLeftColor: 'var(--accent-primary)', borderRadius: 'var(--radius-base)' }}>
                  <div className="flex items-center justify-between pb-3 border-b border-border-main/50">
                     <span className="text-[11px] font-bold uppercase tracking-widest text-text-main flex items-center gap-2">
                        Cart Summary
                        <span className="bg-bg-surface text-text-muted px-1.5 py-0.5 rounded text-[9px]">1 Item</span>
                     </span>
                     <span className="text-[14px] font-bold text-accent-primary font-mono">$50,000,015</span>
                  </div>
                  
                  <div className="flex justify-between items-center text-[10px]">
                     <span className="text-text-secondary">Subtotal</span>
                     <span className="font-mono text-text-main">$50,000,000</span>
                  </div>
                  <div className="flex justify-between items-center text-[10px]">
                     <span className="text-text-secondary">Shipping</span>
                     <span className="font-mono text-accent-primary uppercase font-bold tracking-widest">Free</span>
                  </div>
                  
                  <div className="flex flex-col gap-1.5 pt-2 border-t border-border-main/50">
                     <div className="h-1.5 rounded-full bg-border-main overflow-hidden shadow-inner flex">
                        <div className="h-full bg-accent-primary w-[100%] animate-pulse" />
                     </div>
                     <span className="text-[8px] font-mono uppercase text-text-muted flex justify-between">
                        <span>Free Shipping Unlocked!</span>
                        <span>100%</span>
                     </span>
                  </div>
                  <button className="btn-primary w-full text-[11px] py-3 mt-1 shadow-[0_4px_14px_-4px_var(--color-accent-primary)] hover:shadow-[0_6px_20px_-4px_var(--color-accent-primary)]">Proceed to Checkout</button>
               </div>
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}

function LayoutPropertyInput({ label, value, onChange, helperText }: { label: string, value: string, onChange: (val: string) => void, helperText?: string }) {
  return (
    <div className="flex flex-col p-2.5 rounded bg-bg-header/30 hover:bg-bg-header/50 transition-colors border border-border-main/50 space-y-1">
      <div className="flex items-center justify-between">
        <span className="text-[10px] font-bold text-text-secondary uppercase">{label}</span>
        {helperText && <span className="text-[8px] font-mono text-text-muted italic">{helperText}</span>}
      </div>
      <input 
        type="text" 
        value={value} 
        onChange={e => onChange(e.target.value)}
        className="bg-bg-page border border-border-main/30 rounded px-2 py-1 text-[11px] outline-none focus:border-accent-primary transition-colors text-text-main font-mono"
      />
    </div>
  );
}

function ColorInput({ label, value, onChange }: { label: string, value: string, onChange: (val: string) => void }) {
  return (
    <div className="flex items-center justify-between p-2 rounded bg-bg-header/30 hover:bg-bg-header/50 transition-colors border border-border-main/50">
      <div className="flex flex-col">
        <span className="text-[10px] font-bold text-text-secondary uppercase">{label}</span>
        <span className="text-[9px] font-mono text-text-muted tracking-wide">{value.toUpperCase()}</span>
      </div>
      <div className="relative group overflow-hidden w-10 h-10 rounded shadow-inner border border-border-main cursor-pointer">
        <input 
          type="color" 
          value={value} 
          onChange={e => onChange(e.target.value)}
          className="absolute inset-0 w-[150%] h-[150%] -translate-x-1/4 -translate-y-1/4 cursor-pointer opacity-0"
        />
        <div 
          className="w-full h-full" 
          style={{ backgroundColor: value }}
        />
      </div>
    </div>
  );
}
