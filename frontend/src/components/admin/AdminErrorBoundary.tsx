'use client';

import React, { Component, ErrorInfo, ReactNode } from 'react';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export default class AdminErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Admin Error Boundary caught an error:', error, errorInfo);
    
    // Log to remote server if possible
    if (typeof window !== 'undefined' && (window as any).RemoteLogger) {
      (window as any).RemoteLogger.error('React #310 / Admin Crash', {
        message: error.message,
        stack: error.stack,
        componentStack: errorInfo.componentStack,
        url: window.location.href
      });
    }
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }

      const isHydrationError = this.state.error?.message?.includes('hydration') || 
                               this.state.error?.message?.includes('310');

      return (
        <div className="min-h-[400px] flex items-center justify-center p-8 bg-ink-surface/30 rounded-lg border-2 border-dashed border-hp-color/20">
          <div className="text-center max-w-md">
            <div className="text-4xl mb-4">⚠️</div>
            <h2 className="font-display text-2xl mb-2 text-ink-deep uppercase">
              {isHydrationError ? 'Interface Sync Error' : 'Interface Exception'}
            </h2>
            <p className="text-sm text-text-muted mb-6">
              {isHydrationError 
                ? 'The administrative interface encountered a synchronization issue (React #310). This often happens due to browser extensions or unstable connections.' 
                : 'Something went wrong while rendering this section of the vault.'}
            </p>
            
            <div className="flex gap-3 justify-center">
              <button 
                onClick={() => window.location.reload()}
                className="btn-primary px-6"
              >
                Refresh Interface
              </button>
              <button 
                onClick={() => this.setState({ hasError: false, error: null })}
                className="btn-secondary px-6"
              >
                Try Again
              </button>
            </div>
            
            {this.state.error && (
              <pre className="mt-8 p-4 bg-ink-navy text-hp-color text-[10px] text-left overflow-auto rounded font-mono border border-hp-color/20 max-h-40">
                {this.state.error.message}
                {"\n"}
                {this.state.error.stack}
              </pre>
            )}
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
