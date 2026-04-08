/**
 * Reusable TextInput component with cute styling
 */

import React from 'react';

interface TextInputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
}

export function TextInput({
  label,
  error,
  className = '',
  ...props
}: TextInputProps) {
  return (
    <div className="w-full">
      {label && (
        <label className="block text-sm font-bold text-rit-orange mb-2\">
          {label}
        </label>
      )}
      <input
        className={`
          w-full
          px-4 py-3
          rounded-xl
          border-2 border-rit-gray-light
          bg-white
          text-rit-charcoal
          placeholder-gray-400
          focus:outline-none
          focus:border-rit-orange
          focus:ring-2
          focus:ring-orange-200
          transition-all duration-200
          ${error ? 'border-red-500 focus:border-red-500 focus:ring-red-200' : ''}
          ${className}
        `}
        {...props}
      />
      {error && (
        <p className="mt-1 text-sm text-red-500 font-medium">{error}</p>
      )}
    </div>
  );
}
