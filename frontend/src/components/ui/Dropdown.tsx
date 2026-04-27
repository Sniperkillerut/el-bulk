'use client';

import React, { useState, useRef, useEffect } from 'react';
import Link from 'next/link';

interface DropdownItem {
  label: string;
  href: string;
  onClick?: () => void;
}

interface DropdownProps {
  label?: React.ReactNode;
  trigger?: React.ReactNode;
  items?: DropdownItem[];
  children?: React.ReactNode;
  className?: string;
  dropdownClassName?: string;
  align?: 'left' | 'right' | 'end';
}

export default function Dropdown({ 
  label, 
  trigger,
  items = [], 
  children,
  className = '', 
  dropdownClassName = '',
  align = 'left'
}: DropdownProps) {
  const [isOpen, setIsOpen] = useState(false);
  const timeoutRef = useRef<NodeJS.Timeout | null>(null);

  const handleMouseEnter = () => {
    if (timeoutRef.current) clearTimeout(timeoutRef.current);
    setIsOpen(true);
  };

  const handleMouseLeave = () => {
    timeoutRef.current = setTimeout(() => {
      setIsOpen(false);
    }, 150);
  };

  useEffect(() => {
    return () => {
      if (timeoutRef.current) clearTimeout(timeoutRef.current);
    };
  }, []);

  const alignmentClass = align === 'right' || align === 'end' ? 'right-0' : 'left-0';

  return (
    <div 
      className={`relative ${className}`}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
    >
      <div className="flex items-center gap-1 cursor-pointer">
        {trigger || label}
        {!trigger && (
          <svg 
            width="12" 
            height="12" 
            viewBox="0 0 12 12" 
            fill="currentColor"
            className={`transition-transform duration-200 ${isOpen ? 'rotate-180' : ''}`}
          >
            <path d="M2 4l4 4 4-4" stroke="currentColor" strokeWidth="1.5" fill="none" strokeLinecap="round"/>
          </svg>
        )}
      </div>

      {isOpen && (
        <div className={`absolute top-full ${alignmentClass} pt-1 min-w-max z-50 animate-in fade-in slide-in-from-top-1 duration-200 ${dropdownClassName}`}>
          {children ? (
            <div className="rounded-sm overflow-hidden">
              {children}
            </div>
          ) : (
            <div className="bg-bg-surface border border-border-main rounded-sm shadow-xl overflow-hidden">
              {items.map((item, index) => (
                <Link
                  key={index}
                  href={item.href}
                  onClick={() => {
                    setIsOpen(false);
                    item.onClick?.();
                  }}
                  className="block px-4 py-2 text-sm text-text-secondary transition-colors hover:bg-bg-page hover:text-text-main no-underline"
                >
                  {item.label}
                </Link>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
