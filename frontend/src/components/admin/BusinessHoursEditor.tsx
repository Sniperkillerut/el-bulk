'use client';

import React, { useState, useEffect } from 'react';
import { useLanguage } from '@/context/LanguageContext';

interface DayConfig {
  open: boolean;
  start: string;
  end: string;
}

interface WeekConfig {
  [key: string]: DayConfig;
}

interface BusinessHoursEditorProps {
  value: string;
  onChange: (value: string) => void;
}

const DAYS = [
  { key: 'mon', label: 'Monday', labelEs: 'Lunes' },
  { key: 'tue', label: 'Tuesday', labelEs: 'Martes' },
  { key: 'wed', label: 'Wednesday', labelEs: 'Miércoles' },
  { key: 'thu', label: 'Thursday', labelEs: 'Jueves' },
  { key: 'fri', label: 'Friday', labelEs: 'Viernes' },
  { key: 'sat', label: 'Saturday', labelEs: 'Sábado' },
  { key: 'sun', label: 'Sunday', labelEs: 'Domingo' },
];

const DEFAULT_CONFIG: WeekConfig = {
  mon: { open: true, start: '11:00', end: '19:00' },
  tue: { open: true, start: '11:00', end: '19:00' },
  wed: { open: true, start: '11:00', end: '19:00' },
  thu: { open: true, start: '11:00', end: '19:00' },
  fri: { open: true, start: '11:00', end: '19:00' },
  sat: { open: true, start: '11:00', end: '19:00' },
  sun: { open: false, start: '11:00', end: '19:00' },
};

export default function BusinessHoursEditor({ value, onChange }: BusinessHoursEditorProps) {
  const { t, locale } = useLanguage();
  const [config, setConfig] = useState<WeekConfig>(DEFAULT_CONFIG);

  useEffect(() => {
    try {
      if (value && value.startsWith('{')) {
        const parsed = JSON.parse(value);
        if (parsed.mon && typeof parsed.mon.open === 'boolean') {
          setConfig(parsed);
        }
      } else {
        // If it's not JSON, initialize with defaults and save as JSON immediately
        onChange(JSON.stringify(DEFAULT_CONFIG));
      }
    } catch {
      onChange(JSON.stringify(DEFAULT_CONFIG));
    }
  }, [value]);

  const handleUpdateDay = (day: string, updates: Partial<DayConfig>) => {
    const newConfig = {
      ...config,
      [day]: { ...config[day], ...updates }
    };
    setConfig(newConfig);
    onChange(JSON.stringify(newConfig));
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between border-b border-ink-border/10 pb-2">
        <label className="text-[10px] font-mono-stack block uppercase font-bold text-text-muted">
          {t('pages.admin.settings.hours_label', 'Business Hours Scheduler')}
        </label>
      </div>

      <div className="grid gap-2">
        {DAYS.map(({ key, label, labelEs }) => (
          <div key={key} className="flex items-center gap-3 p-2 bg-white rounded border border-ink-border/10 hover:border-gold/30 transition-colors">
            <div className="w-20">
              <span className="text-xs font-bold text-ink-deep">{locale === 'es' ? labelEs : label}</span>
            </div>
            
            <button
              onClick={() => handleUpdateDay(key, { open: !config[key].open })}
              className={`px-3 py-1 text-[10px] font-bold rounded-full border transition-all ${
                config[key].open 
                  ? 'bg-green-500/10 text-green-600 border-green-500/30' 
                  : 'bg-hp-color/5 text-hp-color border-hp-color/20 opacity-50'
              }`}
            >
              {config[key].open ? 'OPEN' : 'CLOSED'}
            </button>

            {config[key].open && (
              <div className="flex items-center gap-2 ml-auto">
                <input 
                  type="time" 
                  value={config[key].start}
                  onChange={e => handleUpdateDay(key, { start: e.target.value })}
                  className="bg-ink-surface/10 border-none rounded px-2 py-1 text-xs font-mono font-bold outline-none focus:bg-gold/10"
                />
                <span className="text-[10px] text-text-muted">to</span>
                <input 
                  type="time" 
                  value={config[key].end}
                  onChange={e => handleUpdateDay(key, { end: e.target.value })}
                  className="bg-ink-surface/10 border-none rounded px-2 py-1 text-xs font-mono font-bold outline-none focus:bg-gold/10"
                />
              </div>
            )}
          </div>
        ))}
      </div>
      <p className="text-[9px] text-text-muted italic leading-tight">
        Bogotá Express will automatically show as <b>INACTIVE</b> during closed hours or on days marked as &quot;CLOSED&quot;.
      </p>
    </div>
  );
}
