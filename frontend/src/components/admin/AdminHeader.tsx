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
    <header className={`flex flex-col lg:flex-row lg:justify-between lg:items-start ${customMargin} gap-4 lg:gap-8 flex-shrink-0`}>
      <div className="space-y-1">
        <h1 className="font-display text-3xl sm:text-4xl lg:text-5xl tracking-tight text-ink-deep m-0 uppercase leading-tight lg:leading-none">
          {title}
        </h1>
        {subtitle && (
          <p className="font-mono-stack text-[10px] sm:text-xs text-text-muted opacity-60 font-bold uppercase tracking-widest border-l-2 border-gold/30 pl-2 lg:pl-0 lg:border-0">
            {subtitle}
          </p>
        )}
      </div>
      {actions && (
        <div className="flex flex-wrap items-center gap-2 sm:gap-4 lg:justify-end">
          {actions}
        </div>
      )}
    </header>
  );
}
