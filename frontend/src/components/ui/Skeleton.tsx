import React from 'react';

interface SkeletonProps {
  className?: string;
  width?: string | number;
  height?: string | number;
  borderRadius?: string;
  variant?: 'rectangular' | 'circular' | 'text' | 'card';
}

export default function Skeleton({ 
  className = '', 
  width, 
  height, 
  borderRadius, 
  variant = 'rectangular' 
}: SkeletonProps) {
  const baseStyle: React.CSSProperties = {
    width: width || '100%',
    height: height || '1em',
    borderRadius: borderRadius || (variant === 'circular' ? '50%' : 'var(--radius-sm, 2px)'),
  };

  // Kraft paper aesthetic: subtle gradient plus a "rough" grain look if possible
  const kraftSkeletonStyle: React.CSSProperties = {
    ...baseStyle,
    background: 'linear-gradient(90deg, var(--kraft-light) 25%, var(--kraft-mid) 50%, var(--kraft-light) 75%)',
    backgroundSize: '200% 100%',
    animation: 'skeleton-shimmer 1.5s infinite ease-in-out',
  };

  if (variant === 'card') {
    return (
      <div className={`skeleton-card ${className}`} style={{ 
        ...kraftSkeletonStyle, 
        height: height || '350px',
        border: '2px solid var(--kraft-dark)',
        opacity: 0.7
      }}>
        <div className="skeleton-thumb" style={{ height: '70%', background: 'rgba(0,0,0,0.05)', marginBottom: '1rem' }} />
        <div className="skeleton-text" style={{ width: '80%', height: '1rem', background: 'rgba(0,0,0,0.05)', margin: '0 1rem 0.5rem' }} />
        <div className="skeleton-text" style={{ width: '60%', height: '0.8rem', background: 'rgba(0,0,0,0.05)', margin: '0 1rem' }} />
      </div>
    );
  }

  return <div className={`skeleton ${variant} ${className}`} style={kraftSkeletonStyle} />;
}
