'use client';

export default function LoadingSpinner() {
  return (
    <div className="flex flex-col items-center justify-center p-20 w-full min-h-[400px]">
      <div className="w-12 h-12 border-4 border-gold border-t-transparent rounded-full animate-spin mb-4 shadow-lg shadow-gold/20"></div>
      <p className="font-display text-sm tracking-widest text-text-muted animate-pulse uppercase">Synchronizing with Core...</p>
    </div>
  );
}
