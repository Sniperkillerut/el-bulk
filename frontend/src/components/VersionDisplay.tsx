'use client';

import { useEffect, useState } from 'react';

export default function VersionDisplay({ className = "" }: { className?: string }) {
  const [apiVersion, setApiVersion] = useState<string>('loading...');
  const frontendVersion = process.env.NEXT_PUBLIC_APP_VERSION || '0.0.0-dev';

  useEffect(() => {
    const fetchApiVersion = async () => {
      try {
        const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL || ''}/health`);
        const data = await res.json();
        setApiVersion(data.version || 'unknown');
      } catch {
        setApiVersion('unavailable');
      }
    };

    fetchApiVersion();
  }, []);

  return (
    <div className={`flex items-center gap-3 text-[10px] font-mono-stack uppercase tracking-wider opacity-40 hover:opacity-100 transition-opacity ${className}`}>
      <div className="flex items-center gap-1.5 grayscale">
        <span className="bg-hp-color/20 text-hp-color px-1 rounded">Web</span>
        <span className="font-bold">v{frontendVersion}</span>
      </div>
      <span className="text-HP-color/30">|</span>
      <div className="flex items-center gap-1.5 grayscale">
        <span className="bg-HP-color/20 text-HP-color px-1 rounded">API</span>
        <span className="font-bold">v{apiVersion}</span>
      </div>
    </div>
  );
}
