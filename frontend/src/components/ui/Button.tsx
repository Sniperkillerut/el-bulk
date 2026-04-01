'use client';

import React from 'react';

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'danger' | 'success' | 'ghost';
  size?: 'sm' | 'md' | 'lg' | 'xl';
  loading?: boolean;
  fullWidth?: boolean;
}

export default function Button({
  children,
  variant = 'primary',
  size = 'md',
  loading = false,
  fullWidth = false,
  className = '',
  disabled,
  ...props
}: ButtonProps) {
  const baseStyles = 'inline-flex items-center justify-center font-bold transition-all disabled:opacity-40 disabled:cursor-not-allowed uppercase tracking-wider cursor-pointer';
  
  const variants = {
    primary: 'btn-primary shadow-lg shadow-accent-primary/10 hover:shadow-accent-primary/20',
    secondary: 'btn-secondary border-border-main hover:bg-bg-page/5',
    danger: 'bg-status-hp/10 text-status-hp border border-status-hp/30 hover:bg-status-hp/20 hover:text-status-hp shadow-lg shadow-status-hp/5',
    success: 'bg-status-nm/10 text-status-nm border border-status-nm/30 hover:bg-status-nm/20 hover:text-status-nm shadow-lg shadow-status-nm/5',
    ghost: 'hover:bg-bg-page/5 text-text-muted hover:text-text-main',
  };

  const sizes = {
    sm: 'px-3 py-1.5 text-[10px]',
    md: 'px-6 py-2.5 text-xs',
    lg: 'px-8 py-3 text-sm',
    xl: 'px-10 py-4 text-base font-display',
  };

  const combinedClasses = `
    ${baseStyles} 
    ${variants[variant]} 
    ${sizes[size]} 
    ${fullWidth ? 'w-full' : ''} 
    ${className}
  `.trim();

  return (
    <button
      disabled={disabled || loading}
      className={combinedClasses}
      data-theme-area="button"
      data-variant={variant}
      {...props}
    >
      {loading ? (
        <span className="flex items-center gap-2">
          <svg className="animate-spin h-3 w-3" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
          </svg>
          {typeof children === 'string' ? (children.endsWith('...') ? children : `${children}...`) : children}
        </span>
      ) : (
        children
      )}
    </button>
  );
}
