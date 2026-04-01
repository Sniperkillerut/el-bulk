'use client';

import React from 'react';

interface AdminHeaderProps {
  title: string;
  subtitle?: string;
  actions?: React.ReactNode;
}

export default function AdminHeader({ title, subtitle, actions }: AdminHeaderProps) {
  return (
    <header className="flex justify-between items-start mb-2 flex-shrink-0">
      <div className="space-y-1">
        <h1 className="font-display text-5xl tracking-tight text-ink-deep m-0 uppercase leading-none">
          {title}
        </h1>
        {subtitle && (
          <p className="font-mono-stack text-xs text-text-muted opacity-60 font-bold uppercase tracking-widest">
            {subtitle}
          </p>
        )}
      </div>
      {actions && (
        <div className="flex gap-4">
          {actions}
        </div>
      )}
    </header>
  );
}
