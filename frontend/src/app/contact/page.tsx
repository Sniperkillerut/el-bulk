'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import { fetchPublicSettings } from '@/lib/api';
import { PublicSettings } from '@/lib/types';
import { useLanguage } from '@/context/LanguageContext';

function BusinessHours({ hoursJson, t }: { hoursJson: string, t: any }) {
  try {
    const hours = JSON.parse(hoursJson);
    const dayOrder = ['mon', 'tue', 'wed', 'thu', 'fri', 'sat', 'sun'];
    
    return (
      <div className="space-y-1.5 mt-2 max-w-[240px]">
        {dayOrder.map(day => {
          const schedule = hours[day];
          if (!schedule) return null;
          
          return (
            <div key={day} className="flex justify-between items-center text-sm">
              <span className="font-mono-stack uppercase text-text-muted text-[10px] w-8">
                {t(`common.days.${day}.short`, day.toUpperCase())}
              </span>
              <div className="flex-1 border-b border-dotted border-kraft-dark mx-2 h-3" />
              <span className="font-bold text-ink-deep min-w-[85px] text-right">
                {schedule.open 
                  ? `${schedule.start} - ${schedule.end}` 
                  : t('common.hours.closed', 'CLOSED')}
              </span>
            </div>
          );
        })}
      </div>
    );
  } catch (e) {
    return <p className="font-semibold text-ink-deep whitespace-pre-wrap">{hoursJson}</p>;
  }
}

export default function ContactPage() {
  const [settings, setSettings] = useState<PublicSettings | undefined>(undefined);
  const { t } = useLanguage();

  useEffect(() => {
    fetchPublicSettings().then(setSettings).catch(console.error);
  }, []);

  if (!settings) {
    return (
      <div className="max-w-4xl mx-auto px-4 py-16 text-center">
        <div className="skeleton h-64 w-full mb-8" style={{ borderRadius: 8 }} />
        <div className="skeleton h-12 w-1/2 mx-auto" style={{ borderRadius: 4 }} />
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto px-4 py-16">
      {/* Header Stamp */}
      <div className="text-center mb-16">
        <div className="stamp-border inline-block p-1 bg-surface rotate-1">
          <div className="border border-dashed border-kraft-shadow px-10 py-4">
            <h1 className="font-display text-6xl text-ink-deep m-0 uppercase tracking-tight">
              {t('pages.contact.page.title', 'Contact Us').split(' ').map((word, i) => (
                <span key={i} style={i === 1 ? { color: 'var(--gold-dark)' } : {}}>
                  {word}{' '}
                </span>
              ))}
            </h1>
          </div>
        </div>
        <p className="mt-8 font-mono-stack text-sm tracking-widest text-text-muted">
          {'// EL BULK TCG STORE // BOGOTÁ, COLOMBIA'}
        </p>
      </div>

      <div className="grid md:grid-cols-2 gap-12">
        {/* Contact Info - Box Side Style */}
        <div className="cardbox p-8 bg-surface relative overflow-hidden" style={{ borderLeft: '4px solid var(--kraft-dark)' }}>
          <div className="absolute top-0 right-0 opacity-10 pointer-events-none p-4">
             <span style={{ fontSize: '8rem' }}>📍</span>
          </div>
          
          <h2 className="font-display text-3xl mb-6 text-gold-dark">{t('pages.contact.section.visit', 'VISIT THE BOX')}</h2>
          
          <div className="space-y-6">
            <div>
              <p className="text-xs font-mono-stack mb-1 text-text-muted">{t('pages.contact.info.address', 'ADDRESS')}</p>
              <p className="font-semibold text-ink-deep leading-snug whitespace-pre-wrap">
                {settings.contact_address}
              </p>
            </div>
            
            <div>
              <p className="text-xs font-mono-stack mb-1 text-text-muted">{t('pages.contact.info.hours', 'HOURS')}</p>
              <BusinessHours hoursJson={settings.contact_hours} t={t} />
            </div>

            <div className="pt-4 mt-4 border-t border-dashed border-kraft-dark">
              <p className="text-sm italic text-text-secondary">
                {t('pages.contact.info.quote', "\"Just look for the stack of shoeboxes near the back entrance.\"")}
              </p>
            </div>
          </div>
        </div>

        {/* Digital Contact - Shipping Label Style */}
        <div className="card shadow-md flex flex-col items-center justify-center p-8 bg-kraft-light relative h-full">
           <div className="w-full mb-8">
             <h2 className="font-display text-3xl mb-6 text-ink-deep">{t('pages.contact.section.digital', 'DIGITAL COMMS')}</h2>
             
             <div className="space-y-6">
               <a href={`https://wa.me/${settings.contact_phone.replace(/\D/g, '')}`} target="_blank" rel="noopener noreferrer" className="flex items-center gap-4 transition-transform hover:translate-x-1 group">
                 <div className="w-10 h-10 rounded flex items-center justify-center bg-nm-color text-white shadow-sm">
                   <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"></path></svg>
                 </div>
                 <div>
                   <p className="text-xs font-mono-stack text-text-muted">{t('pages.contact.info.whatsapp', 'WHATSAPP / SALES')}</p>
                   <p className="font-bold text-ink-deep group-hover:text-nm-color text-sm lg:text-base">{settings.contact_phone}</p>
                 </div>
               </a>

               <a href={`mailto:${settings.contact_email}`} className="flex items-center gap-4 transition-transform hover:translate-x-1 group">
                 <div className="w-10 h-10 rounded flex items-center justify-center bg-gold-dark text-white shadow-sm">
                   <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path><polyline points="22,6 12,13 2,6"></polyline></svg>
                 </div>
                 <div>
                   <p className="text-xs font-mono-stack text-text-muted">{t('pages.contact.info.email', 'EMAIL')}</p>
                   <p className="font-bold text-ink-deep group-hover:text-gold-dark text-sm lg:text-base">{settings.contact_email}</p>
                 </div>
               </a>

               <a href={`https://instagram.com/${settings.contact_instagram}`} target="_blank" rel="noopener noreferrer" className="flex items-center gap-4 transition-transform hover:translate-x-1 group">
                 <div className="w-10 h-10 rounded flex items-center justify-center bg-hp-color text-white shadow-sm">
                   <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><rect x="2" y="2" width="20" height="20" rx="5" ry="5"></rect><path d="M16 11.37A4 4 0 1 1 12.63 8 4 4 0 0 1 16 11.37z"></path><line x1="17.5" y1="6.5" x2="17.51" y2="6.5"></line></svg>
                 </div>
                 <div>
                   <p className="text-xs font-mono-stack text-text-muted">{t('pages.contact.info.instagram', 'INSTAGRAM')}</p>
                   <p className="font-bold text-ink-deep group-hover:text-hp-color text-sm lg:text-base">@{settings.contact_instagram}</p>
                 </div>
               </a>
             </div>
           </div>
           
           <div className="w-full mt-auto pt-6 border-t border-ink-border text-center">
              <Link href="/bulk" className="btn-primary w-full shadow-sm">{t("pages.contact.buttons.sell_bulk", "SELL US YOUR BULK →")}</Link>
           </div>
        </div>
      </div>

      {/* Map integration hidden for future use */}
    </div>
  );
}
