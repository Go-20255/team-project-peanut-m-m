/**
 * Reusable Card component with bubbly styling
 */

import React from 'react';

interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
  children: React.ReactNode;
  variant?: 'default' | 'elevated';
}

export function Card({
  children,
  variant = 'default',
  className = '',
  ...props
}: CardProps) {
  const variantClasses = {
    default:
      'bg-white border-3 border-rit-orange shadow-bubble',
    elevated:
      'bg-white shadow-bubble-lg border-3 border-rit-orange',
  };

  return (
    <div
      className={`
        rounded-2xl
        p-6
        transition-all duration-200
        ${variantClasses[variant]}
        ${className}
      `}
      {...props}
    >
      {children}
    </div>
  );
}
