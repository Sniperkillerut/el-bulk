'use client';

import { useSearchParams } from 'next/navigation';
import { Suspense } from 'react';
import Link from 'next/link';

function LoginContent() {
  const searchParams = useSearchParams();
  const callbackUrl = searchParams.get('callbackUrl') || '/';

  // In a real app, we'd append callbackUrl to the login request
  const handleLogin = (provider: string) => {
    // Redirect to backend OAuth endpoint with return URL
    const apiUrl = process.env.NEXT_PUBLIC_API_URL || '';
    const redirectUrl = encodeURIComponent(callbackUrl);
    window.location.href = `${apiUrl}/api/auth/${provider}?redirect_url=${redirectUrl}`;
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-kraft-paper p-6 relative overflow-hidden">
      {/* Dynamic Background Elements */}
      <div className="absolute top-[-10%] left-[-10%] w-[40%] h-[40%] bg-gold/10 rounded-full blur-[120px] animate-pulse"></div>
      <div className="absolute bottom-[-10%] right-[-10%] w-[40%] h-[40%] bg-hp-color/5 rounded-full blur-[120px] animate-pulse" style={{ animationDelay: '2s' }}></div>

      <div className="w-full max-w-md z-10">
        <div className="cardbox p-8 backdrop-blur-md bg-white/40 border border-white/20 shadow-2xl relative overflow-hidden group">
          {/* Decorative Gold Corner */}
          <div className="absolute top-0 right-0 w-24 h-24 bg-gold/10 rotate-45 translate-x-12 -translate-y-12 transition-transform group-hover:scale-110"></div>
          
          <div className="text-center mb-10 mt-4">
            <h1 className="font-display text-4xl tracking-tighter mb-2 text-ink-deep">
              WELCOME <span className="text-gold">BACK</span>
            </h1>
            <p className="font-mono-stack text-[10px] uppercase tracking-[0.2em] text-text-muted opacity-70">
              Secure Access // El Bulk Collective
            </p>
          </div>

          <div className="space-y-4">
            <button 
              onClick={() => handleLogin('google')}
              className="w-full flex items-center justify-center gap-4 bg-white border border-kraft-dark/20 py-4 px-6 rounded-lg text-sm font-bold text-ink-deep hover:bg-gold/5 hover:border-gold transition-all shadow-sm hover:shadow-gold/10 group/btn"
            >
              <svg width="20" height="20" viewBox="0 0 24 24">
                <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
                <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-1 .67-2.28 1.07-3.71 1.07-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
                <path fill="#FBBC05" d="M5.84 14.11c-.22-.67-.35-1.39-.35-2.11s.13-1.44.35-2.11V7.06H2.18c-.71 1.48-1.11 3.13-1.11 4.94s.4 3.46 1.11 4.94l3.66-2.83z"/>
                <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.06l3.66 2.83c.87-2.6 3.3-4.53 12-4.53z"/>
              </svg>
              SIGN IN WITH GOOGLE
              <span className="opacity-0 group-hover/btn:opacity-100 transition-opacity ml-auto">→</span>
            </button>

            <button 
              onClick={() => handleLogin('facebook')}
              className="w-full flex items-center justify-center gap-4 bg-white border border-kraft-dark/20 py-4 px-6 rounded-lg text-sm font-bold text-ink-deep hover:bg-gold/5 hover:border-gold transition-all shadow-sm hover:shadow-gold/10 group/btn"
            >
              <svg width="20" height="20" fill="#1877F2" viewBox="0 0 24 24">
                <path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z"/>
              </svg>
              SIGN IN WITH FACEBOOK
              <span className="opacity-0 group-hover/btn:opacity-100 transition-opacity ml-auto">→</span>
            </button>

            <div className="py-4 flex items-center gap-4 opacity-30">
              <div className="h-px bg-ink-deep grow"></div>
              <span className="font-mono-stack text-[9px] font-bold">OR</span>
              <div className="h-px bg-ink-deep grow"></div>
            </div>

            <button 
              onClick={() => handleLogin('apple')}
              className="w-full flex items-center justify-center gap-4 bg-ink-deep py-4 px-6 rounded-lg text-sm font-bold text-white hover:bg-black transition-all shadow-lg group/btn"
            >
              <svg width="18" height="18" fill="white" viewBox="0 0 24 24">
                <path d="M17.05 20.28c-.96.95-2.06 1.43-3.29 1.43-1.19 0-2.31-.48-3.37-1.43-1.07-.95-2.18-1.43-3.33-1.43-1.25 0-2.39.48-3.41 1.43l-.44.41c-1.46-2.17-2.19-4.52-2.19-7.05 0-2.61.76-4.78 2.27-6.52 1.51-1.74 3.42-2.61 5.73-2.61 1.15 0 2.25.26 3.32.79.97.48 1.82.72 2.54.72.64 0 1.43-.24 2.37-.72 1.03-.53 2.1-.79 3.22-.79 1.7 0 3.14.53 4.31 1.59-2.02 1.25-3.03 3.03-3.03 5.34 0 1.84.66 3.34 1.98 4.51-.55 1.53-1.39 2.81-2.51 3.86l-.77.81zM11.96 4.31c-.02-1.18.39-2.22 1.24-3.13.88-.93 1.93-1.45 3.16-1.54.1 1.21-.33 2.27-1.29 3.18-.89.84-1.93 1.34-3.11 1.49z"/>
              </svg>
              SIGN IN WITH APPLE
              <span className="opacity-0 group-hover/btn:opacity-100 transition-opacity ml-auto">→</span>
            </button>
          </div>

          <div className="mt-12 pt-6 border-t border-kraft-dark/10 text-center">
            <p className="font-mono-stack text-[9px] text-text-muted leading-relaxed max-w-[240px] mx-auto">
              BY JOINING THE COLLECTIVE, YOU AGREE TO OUR <a href="/terms" className="text-hp-color underline">TERMS OF ENGAGEMENT</a> AND <a href="/privacy" className="text-hp-color underline">DATA PROTOCOLS</a>.
            </p>
          </div>
        </div>

        <div className="mt-6 text-center animate-in fade-in slide-in-from-bottom-2 duration-700">
          <Link href="/" className="text-[10px] font-mono-stack font-bold text-text-muted hover:text-gold no-underline tracking-widest uppercase">
            ← Return to Homebase
          </Link>
        </div>
      </div>
      
      {/* Visual Accents */}
      <div className="absolute top-12 left-12 font-display text-[80px] opacity-[0.03] select-none pointer-events-none">COLLECTIVE</div>
      <div className="absolute bottom-12 right-12 font-display text-[80px] opacity-[0.03] select-none pointer-events-none rotate-180">LOGISTICS</div>
    </div>
  );
}

export default function LoginPage() {
  return (
    <Suspense fallback={<div className="min-h-screen bg-kraft-paper flex items-center justify-center p-12 text-gold font-mono-stack animate-pulse uppercase tracking-widest font-bold">Initializing Auth Matrix...</div>}>
      <LoginContent />
    </Suspense>
  );
}
