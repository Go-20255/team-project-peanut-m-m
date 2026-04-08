/**
 * Reusable Button component with cute bubbly styling
 */

import React from 'react';

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'accent';
  size?: 'sm' | 'md' | 'lg';
  children: React.ReactNode;
  isLoading?: boolean;
}

export function Button({
  variant = 'primary',
  size = 'md',
  children,
  isLoading = false,
  disabled,
  className = '',
  ...props
}: ButtonProps) {
  const variantClasses = {
    primary:
      'bg-rit-orange bg-orange-600 font-bold shadow-bubble hover:shadow-bubble-lg border-2 border-white text-black',
    secondary:
      'bg-rit-orange bg-orange-600 font-bold shadow-bubble hover:shadow-bubble-lg border-2 border-white text-black',
    accent:
      'bg-gradient-to-br from-accent-gold to-orange-400 hover:from-orange-400 hover:to-orange-500 text-white font-bold shadow-bubble hover:shadow-bubble-lg border-2 border-white',
  };

  const getVariantStyle = (v: typeof variant) => {
    return {};
  };

  const sizeClasses = {
    sm: 'px-4 py-2 text-sm rounded-lg',
    md: 'px-6 py-3 text-base rounded-xl',
    lg: 'px-8 py-4 text-lg rounded-2xl',
  };

  return (
    <button
      disabled={disabled || isLoading}
      style={getVariantStyle(variant)}
      className={`
        font-semibold
        transition-all duration-200
        active:scale-95
        disabled:opacity-50 disabled:cursor-not-allowed disabled:scale-100
        ${variantClasses[variant]}
        ${sizeClasses[size]}
        ${className}
      `}
      {...props}
    >
      {isLoading ? (
        <span className="flex items-center gap-2">
          <span className="animate-spin">⏳</span>
          Loading...
        </span>
      ) : (
        children
      )}
    </button>
  );
}
