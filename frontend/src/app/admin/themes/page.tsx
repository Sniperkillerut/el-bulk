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

const CARDBOARD_DEFAULT: ThemeInput = {
  name: 'Cardboard',
  bg_page: '#e6dac3',
  bg_header: '#1a1f2e',
  bg_surface: '#fdfbf7',
  text_main: '#3b3127',
  text_secondary: '#5c4e3d',
  text_muted: '#8b795c',
  text_on_accent: '#2c251d',
  accent_primary: '#d4af37',
  accent_primary_hover: '#b8961e',
  border_main: '#d4c5ab',
  status_nm: '#2e7d32',
  status_lp: '#558b2f',
  status_mp: '#ef6c00',
  status_hp: '#c62828',
  radius_base: '8px',
  padding_card: '12px',
  gap_grid: '16px',
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
    type: false,
    logic: false,
    geometry: false
  });
  const [previewMode, setPreviewMode] = useState<'desktop' | 'mobile'>('desktop');

  useEffect(() => {
    async function loadData() {
      setLoading(true);
      try {
        const [tData, sData] = await Promise.all([
          fetchThemes(),
          getAdminSettings(token!)
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
      text_main: theme.text_main,
      text_secondary: theme.text_secondary,
      text_muted: theme.text_muted,
      text_on_accent: theme.text_on_accent,
      accent_primary: theme.accent_primary,
      accent_primary_hover: theme.accent_primary_hover,
      border_main: theme.border_main,
      status_nm: theme.status_nm,
      status_lp: theme.status_lp,
      status_mp: theme.status_mp,
      status_hp: theme.status_hp,
      radius_base: theme.radius_base,
      padding_card: theme.padding_card,
      gap_grid: theme.gap_grid,
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
      await updateAdminSettings(token!, { default_theme_id: id });
      setMessage({ text: 'Default theme updated', type: 'success' });
      const sData = await getAdminSettings(token!);
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
        
        <header className="flex items-center justify-between border-b border-border-main pb-6">
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
            <div className="grid gap-2">
              {themes.map(t => (
                <div 
                  key={t.id}
                  onClick={() => handleEdit(t)}
                  className={`
                    p-3 rounded border cursor-pointer transition-all relative group flex flex-col gap-2
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
          <div className="lg:col-span-7 lg:sticky lg:top-8 space-y-4">
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
                  title="Surface Layout" 
                  isOpen={expanded.surface} 
                  onToggle={() => setExpanded(p => ({...p, surface: !p.surface}))}
                >
                  <div className="space-y-2">
                    <ColorInput label="Page Canvas" value={form.bg_page} onChange={val => setForm({...form, bg_page: val})} />
                    <ColorInput label="Header Strip" value={form.bg_header} onChange={val => setForm({...form, bg_header: val})} />
                    <ColorInput label="Card Solid" value={form.bg_surface} onChange={val => setForm({...form, bg_surface: val})} />
                    <ColorInput label="Wireframes" value={form.border_main} onChange={val => setForm({...form, border_main: val})} />
                  </div>
                </Collapsible>

                {/* 3. Brand Palette */}
                <Collapsible 
                  title="Signature" 
                  isOpen={expanded.signature} 
                  onToggle={() => setExpanded(p => ({...p, signature: !p.signature}))}
                >
                  <div className="space-y-2">
                    <ColorInput label="Main Soul" value={form.accent_primary} onChange={val => setForm({...form, accent_primary: val})} />
                    <ColorInput label="Trace Soul" value={form.accent_primary_hover} onChange={val => setForm({...form, accent_primary_hover: val})} />
                    <ColorInput label="Contrast In" value={form.text_on_accent} onChange={val => setForm({...form, text_on_accent: val})} />
                  </div>
                </Collapsible>

                {/* 4. Type Palette */}
                <Collapsible 
                  title="Type System" 
                  isOpen={expanded.type} 
                  onToggle={() => setExpanded(p => ({...p, type: !p.type}))}
                >
                  <div className="space-y-2">
                    <ColorInput label="Primary Ink" value={form.text_main} onChange={val => setForm({...form, text_main: val})} />
                    <ColorInput label="Sub Ink" value={form.text_secondary} onChange={val => setForm({...form, text_secondary: val})} />
                    <ColorInput label="Ghost Ink" value={form.text_muted} onChange={val => setForm({...form, text_muted: val})} />
                  </div>
                </Collapsible>

                {/* 5. Logic Palette */}
                <Collapsible 
                  title="Condition Logic" 
                  isOpen={expanded.logic} 
                  onToggle={() => setExpanded(p => ({...p, logic: !p.logic}))}
                >
                  <div className="space-y-2">
                    <ColorInput label="Mint (NM)" value={form.status_nm} onChange={val => setForm({...form, status_nm: val})} />
                    <ColorInput label="Played (LP)" value={form.status_lp} onChange={val => setForm({...form, status_lp: val})} />
                    <ColorInput label="Worn (MP)" value={form.status_mp} onChange={val => setForm({...form, status_mp: val})} />
                    <ColorInput label="Heavy (HP)" value={form.status_hp} onChange={val => setForm({...form, status_hp: val})} />
                  </div>
                </Collapsible>

                {/* 6. Layout Geometry */}
                <Collapsible 
                  title="Geometry & Spacing" 
                  icon={<Icons.Maximize />} 
                  isOpen={expanded.geometry} 
                  onToggle={() => setExpanded(p => ({...p, geometry: !p.geometry}))}
                >
                  <div className="space-y-2">
                    <LayoutPropertyInput label="Corner Radius" value={form.radius_base} onChange={val => setForm({...form, radius_base: val})} helperText="Card rounding (px/rem)" />
                    <LayoutPropertyInput label="Card Spacing" value={form.padding_card} onChange={val => setForm({...form, padding_card: val})} helperText="Internal padding" />
                    <LayoutPropertyInput label="Grid Gap" value={form.gap_grid} onChange={val => setForm({...form, gap_grid: val})} helperText="Spacing between cards" />
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
  const mockProducts = [
    { id: '1', name: 'Black Lotus', set_name: 'Limited Edition Alpha', price: 50000000, stock: 1, condition: 'NM', image_url: 'https://cards.scryfall.io/large/front/b/d/bd8fa327-dd41-4737-8f19-2cf5eb1f7cdd.jpg' },
    { id: '2', name: 'Lightning Bolt', set_name: 'Fourth Edition', price: 15000, stock: 24, condition: 'LP', image_url: 'https://cards.scryfall.io/large/front/f/2/f29ba16f-c8fb-42fe-aabf-87089cb214a7.jpg' },
    { id: '3', name: 'Ragavan', set_name: 'Modern Horizons 2', price: 320000, stock: 0, condition: 'HP', image_url: 'https://cards.scryfall.io/large/front/a/1/a1111111-1111-1111-1111-111111111111.jpg' }
  ];

  return (
    <div 
      className={`rounded-lg border shadow-2xl overflow-hidden origin-top flex flex-col transition-all duration-500 ease-in-out border-border-main scrollbar-none`} 
      style={{ 
        backgroundColor: form.bg_page, 
        borderColor: form.border_main, 
        borderRadius: form.radius_base,
        width: isMobile ? '375px' : '100%',
        minHeight: isMobile ? '667px' : '700px',
        maxHeight: isMobile ? '667px' : '800px',
      }}
    >
      {/* Mini Header */}
      <div className="p-4 border-b border-border-main flex items-center justify-between sticky top-0 z-20" style={{ backgroundColor: form.bg_header, borderColor: form.border_main }}>
        <div className="flex items-center gap-4">
          <div className="w-8 h-8 rounded bg-accent-primary" style={{ backgroundColor: form.accent_primary }} />
          {!isMobile && (
            <nav className="flex gap-4 text-[10px] font-bold uppercase tracking-widest" style={{ color: '#ffffff' }}>
              <span className="opacity-60">SINGLES</span>
              <span className="opacity-60">SEALED</span>
              <span className="opacity-60">DECKS</span>
            </nav>
          )}
        </div>
        <div className="w-6 h-6 rounded-full border-2" style={{ borderColor: form.accent_primary }} />
      </div>

      <div className="flex-1 flex overflow-hidden">
        {/* Mini Sidebar - Hidden on mobile preview */}
        {!isMobile && (
          <aside className="w-48 p-4 border-r border-border-main hidden md:block overflow-y-auto" style={{ borderColor: form.border_main }}>
            <div className="space-y-6">
              <div>
                <p className="text-[8px] font-bold uppercase tracking-[0.2em] mb-2" style={{ color: form.text_muted }}>Keywords</p>
                <div className="h-8 rounded border px-2 flex items-center" style={{ backgroundColor: form.bg_surface, borderColor: form.border_main }}>
                  <span className="text-[10px]" style={{ color: form.text_muted }}>Search cards...</span>
                </div>
              </div>

              <div>
                <p className="text-[8px] font-bold uppercase tracking-[0.2em] mb-3" style={{ color: form.text_muted }}>Availability</p>
                <div className="flex items-center gap-2">
                  <div className="w-3 h-3 rounded-sm border" style={{ borderColor: form.accent_primary, backgroundColor: form.accent_primary }} />
                  <span className="text-[10px] font-bold" style={{ color: form.text_main }}>In Stock Only</span>
                </div>
              </div>

              <div className="space-y-3">
                <p className="text-[10px] font-display font-bold uppercase" style={{ color: form.text_main }}>Condition ▼</p>
                <div className="space-y-2 pl-1">
                  {['NM', 'LP', 'MP', 'HP'].map(c => (
                    <div key={c} className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <div className="w-3 h-3 rounded-sm border" style={{ borderColor: form.border_main, backgroundColor: form.bg_surface }} />
                        <span className="text-[10px] font-bold" style={{ color: form.text_secondary }}>{c}</span>
                      </div>
                      <span className="text-[8px] font-bold" style={{ color: form.accent_primary }}>(12)</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </aside>
        )}

        {/* Mini Grid Container */}
        <main className="flex-1 p-6 overflow-y-auto custom-scrollbar">
          <div className="mb-6">
             <p className="text-[8px] font-mono mb-1 uppercase tracking-widest" style={{ color: form.text_muted }}>MTG / SINGLES</p>
             <h2 className={`${isMobile ? 'text-xl' : 'text-3xl'} font-display leading-tight`} style={{ color: form.text_main }}>PREVIEW SINGLES</h2>
             <div className="h-0.5 mt-2" style={{ backgroundColor: form.text_main }} />
          </div>

          <div className="flex items-center justify-between mb-4">
            <span className="text-[10px] font-mono" style={{ color: form.text_muted }}>3 results</span>
            <div className="flex items-center gap-2">
               <span className="text-[8px] font-bold uppercase" style={{ color: form.text_muted }}>Sort</span>
               <div className="h-6 w-20 md:w-24 rounded border px-2 flex items-center justify-between" style={{ backgroundColor: form.bg_surface, borderColor: form.border_main }}>
                  <span className="text-[9px]" style={{ color: form.text_main }}>Newest</span>
                  <span className="text-[8px]" style={{ color: form.text_muted }}>▼</span>
               </div>
            </div>
          </div>

          <div className={`grid ${isMobile ? 'grid-cols-1' : 'grid-cols-2'}`} style={{ gap: form.gap_grid }}>
            {mockProducts.slice(0, isMobile ? 2 : 3).map(p => (
              <div key={p.id} className="border flex flex-col overflow-hidden transition-all hover:-translate-y-1" style={{ backgroundColor: form.bg_surface, borderColor: form.border_main, borderRadius: form.radius_base }}>
                <div className={`${isMobile ? 'aspect-[21/9]' : 'aspect-[3/4]'} bg-bg-page/20 relative`} style={{ backgroundColor: form.bg_page + '33' }}>
                  <div className="absolute inset-0 flex items-center justify-center p-4">
                    <div className="w-full h-full rounded bg-bg-header/20 border border-white/5" style={{ backgroundColor: form.bg_header + '22' }} />
                  </div>
                  <div className="absolute top-2 left-2 flex gap-1">
                    <span className="text-[7px] font-bold px-1 py-0.5 rounded border" style={{ 
                      backgroundColor: p.condition === 'NM' ? form.status_nm + '22' : p.condition === 'LP' ? form.status_lp + '22' : form.status_hp + '22',
                      borderColor: p.condition === 'NM' ? form.status_nm : p.condition === 'LP' ? form.status_lp : form.status_hp,
                      color: p.condition === 'NM' ? form.status_nm : p.condition === 'LP' ? form.status_lp : form.status_hp
                    }}>
                      {p.condition}
                    </span>
                  </div>
                </div>
                <div className="flex flex-col flex-1 gap-2" style={{ padding: form.padding_card }}>
                  <div>
                    <h5 className="text-[10px] font-bold truncate leading-tight" style={{ color: form.text_main }}>{p.name}</h5>
                    <p className="text-[8px] truncate opacity-60" style={{ color: form.text_secondary }}>{p.set_name}</p>
                  </div>
                  <div className="mt-auto pt-2 border-t flex items-center justify-between" style={{ borderColor: form.border_main }}>
                    <span className="text-[10px] font-bold font-mono" style={{ color: form.text_main }}>${p.price.toLocaleString()}</span>
                    <button className="text-[8px] font-display font-bold px-2 py-1 rounded" style={{ backgroundColor: form.bg_header, color: '#ffffff' }}>
                      {p.stock > 0 ? 'ADD' : 'SOLD'}
                    </button>
                  </div>
                </div>
              </div>
            ))}
          </div>

          <div className="mt-8 flex justify-center gap-1">
            {[1, 2, 3].map(n => (
              <div key={n} className="w-6 h-6 flex items-center justify-center rounded border text-[9px] font-mono" style={{ 
                backgroundColor: n === 1 ? form.bg_header : form.bg_surface,
                borderColor: form.border_main,
                color: n === 1 ? '#ffffff' : form.text_main
              }}>
                {n}
              </div>
            ))}
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
