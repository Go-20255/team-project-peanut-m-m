/**
 * Reusable Badge component for status indicators
 */

import React from 'react';

interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  children: React.ReactNode;
  variant?: 'info' | 'success' | 'warning' | 'error';
}

export function Badge({
  children,
  variant = 'info',
  className = '',
  ...props
}: BadgeProps) {
  const variantClasses = {
    info: 'bg-orange-100 text-rit-charcoal',
    success: 'bg-orange-200 text-rit-charcoal',
    warning: 'bg-orange-300 text-white',
    error: 'bg-orange-600 text-white',
  };

  return (
    <span
      className={`
        inline-block
        px-3 py-1
        rounded-full
        text-sm
        font-semibold
        ${variantClasses[variant]}
        ${className}
      `}
      {...props}
    >
      {children}
    </span>
  );
}
