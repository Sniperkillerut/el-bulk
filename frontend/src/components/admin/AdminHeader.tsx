'use client';

import React from 'react';

interface AdminHeaderProps {
  title: string;
  subtitle?: string;
  actions?: React.ReactNode;
  customMargin?: string;
}

export default function AdminHeader({ title, subtitle, actions, customMargin = 'mb-6' }: AdminHeaderProps) {
  return (
    <header className={`flex flex-col lg:flex-row lg:justify-between lg:items-center ${customMargin} gap-2 lg:gap-4 flex-shrink-0`}>
      <div className="space-y-0.5">
        <h1 className="font-display text-2xl sm:text-3xl lg:text-4xl tracking-tight text-ink-deep m-0 uppercase leading-none">
          {title}
        </h1>
        {subtitle && (
          <p className="font-mono-stack text-[9px] sm:text-[10px] text-text-muted opacity-60 font-bold uppercase tracking-widest border-l-2 border-gold/30 pl-2 lg:pl-0 lg:border-0">
            {subtitle}
          </p>
        )}
      </div>
      {actions && (
        <div className="flex flex-wrap items-center gap-1.5 sm:gap-2 lg:justify-end">
          {actions}
        </div>
      )}
    </header>
  );
}
